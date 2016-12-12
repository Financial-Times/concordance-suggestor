package main

import (
	"net/http"
	"os"
	"time"

	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/concordance-suggestor/suggestor"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
)

func main() {
	app := cli.App("concordance-suggestor-neo4j", "A public RESTful API to try and suggest organisation concordances")
	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://ftaps67109-law1a-eu-p:8080/db/data",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "log-level",
		Value:  "INFO",
		Desc:   "Log level to use",
		EnvVar: "LOG_LEVEL",
	})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphiteTCPAddress",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)",
		EnvVar: "GRAPHITE_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphitePrefix",
		Value:  "",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. content.test.public.content.by.concept.api.ftaps59382-law1a-eu-t",
		EnvVar: "GRAPHITE_PREFIX",
	})
	logMetrics := app.Bool(cli.BoolOpt{
		Name:   "logMetrics",
		Value:  false,
		Desc:   "Whether to log metrics. Set to true if running locally and you want metrics output",
		EnvVar: "LOG_METRICS",
	})
	env := app.String(cli.StringOpt{
		Name:  "env",
		Value: "local",
		Desc:  "environment this app is running in",
	})

	app.Action = func() {
		parsedLogLevel, err := log.ParseLevel(*logLevel)
		if err != nil {
			log.WithFields(log.Fields{"logLevel": logLevel, "err": err}).Fatal("Incorrect log level")
		}
		log.SetLevel(parsedLogLevel)

		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)

		log.Infof("concordance-suggestor will listen on port: %s, connecting to: %s", *port, *neoURL)
		runServer(*neoURL, *port, *env)
	}
	log.Infof("Application started with args %s", os.Args)
	app.Run(os.Args)
}

func runServer(neoURL string, port string, env string) {

	conf := neoutils.ConnectionConfig{
		BatchSize:     1024,
		Transactional: false,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
			},
			Timeout: 1 * time.Minute,
		},
		BackgroundConnect: true,
	}
	db, err := neoutils.Connect(neoURL, &conf)

	if err != nil {
		log.Fatalf("Error connecting to neo4j %s", err)
	}

	suggestor.SuggestorDriver = suggestor.NewCypherDriver(db, env)

	servicesRouter := mux.NewRouter()

	// Health checks and standards first
	servicesRouter.HandleFunc("/__health", v1a.Handler("ConcordanceSuggestor Healthchecks",
		"Checks for accessing neo4j", suggestor.HealthCheck()))

	// Then API specific ones:
	// Should be /suggestions?contentId=uuid and /suggestions?v1OrganisationId=uuid ????
	servicesRouter.HandleFunc("/content/{uuid}", suggestor.GetConcordanceSuggestionForContentItem).Methods("GET")
	servicesRouter.HandleFunc("/organisations/{uuid}", suggestor.GetConcordanceSuggestionForOrganisation).Methods("GET")

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	// The following endpoints should not be monitored or logged (varnish calls one of these every second, depending on config)
	// The top one of these build info endpoints feels more correct, but the lower one matches what we have in Dropwizard,
	// so it's what apps expect currently same as ping, the content of build-info needs more definition
	http.HandleFunc(status.PingPath, status.PingHandler)
	http.HandleFunc(status.PingPathDW, status.PingHandler)
	http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	http.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)
	http.HandleFunc("/__gtg", suggestor.GoodToGo)
	http.Handle("/", monitoringRouter)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
