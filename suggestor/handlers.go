package suggestor

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
)

// PeopleDriver for cypher queries
var SuggestorDriver Driver
var CacheControlHeader string

//var maxAge = 24 * time.Hour

// HealthCheck does something
func HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact: "Unable to respond to concordance suggestor requests",
		Name:           "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:     "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/concordance-suggestor",
		Severity:       1,
		TechnicalSummary: `Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find
				the neoUrl as a parameter in hieradata for this service. `,
		Checker: Checker,
	}
}

// Checker does more stuff
func Checker() (string, error) {
	err := SuggestorDriver.CheckConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

// Ping says pong
func Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

//GoodToGo returns a 503 if the healthcheck fails - suitable for use from varnish to check availability of a node
func GoodToGo(writer http.ResponseWriter, req *http.Request) {
	if _, err := Checker(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}

}

// BuildInfoHandler - This is a stop gap and will be added to when we can define what we should display here
func BuildInfoHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "build-info")
}

// MethodNotAllowedHandler handles 405
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

// GetConcordanceSuggestion is the public API
func GetConcordanceSuggestionForContentItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedId, err := uuid.FromString(vars["uuid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	concordance, found, err := SuggestorDriver.getOrganisationSuggestionsForAContentItem(requestedId)

	// TODO Should handle error

	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Concordance not found."}`))
		return
	}

	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(concordance); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Concordance could not be retrieved, err=` + err.Error() + `"}`))
	}
}

func GetConcordanceSuggestionForOrganisation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedId, err := uuid.FromString(vars["uuid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	concorded, err := SuggestorDriver.isConcordedToTmeAlready(requestedId)
	// TODO Should handle error

	if concorded {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`{"message":"Already concorded."}`))
		return
	}

	concordance, err := SuggestorDriver.getSuggestionsForOrganisation(requestedId)

	// TODO Should handle error

	if (len(concordance) == 0) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Concordance not found."}`))
		return
	}

	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(concordance); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Concordance could not be retrieved, err=` + err.Error() + `"}`))
	}
}
