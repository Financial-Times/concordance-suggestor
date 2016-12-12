package suggestor

import (
	"fmt"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
	"github.com/satori/go.uuid"
)

// Driver interface
type Driver interface {
	getOrganisationSuggestionsForAContentItem(contentId uuid.UUID) ([]Organisation, bool, error)
	CheckConnectivity() error
	getAllMajorMentions(contentUuid uuid.UUID) ([]uuid.UUID, error)
	isConcordedToTmeAlready(organisationUuid uuid.UUID) (bool, error)
	getSuggestionsForOrganisation(organisationUuid uuid.UUID) ([]Organisation, error)
}

// CypherDriver struct
type CypherDriver struct {
	conn neoutils.NeoConnection
	env  string
}

//NewCypherDriver instantiate driver
func NewCypherDriver(conn neoutils.NeoConnection, env string) CypherDriver {
	return CypherDriver{conn, env}
}

// CheckConnectivity tests neo4j by running a simple cypher query
func (pcw CypherDriver) CheckConnectivity() error {
	return neoutils.Check(pcw.conn)
}

func (pcw CypherDriver) getOrganisationSuggestionsForAContentItem(contentId uuid.UUID) ([]Organisation, bool, error) {

	// Get all major mentions
	majorMentionsOrgUuids, err := pcw.getAllMajorMentions(contentId)

	if err != nil {
		return []Organisation{}, false, err
	}

	for _, majorMentionUuid := range majorMentionsOrgUuids {

		// Get all v2 mentions annotations for all content that has that major mentions
		log.Infof("MajorMention UUID: %v", majorMentionUuid)
		//MATCH (organisation:Organisation{uuid:"a18e6c95-e69a-3b3b-ab7d-5f5c7d9bbada"})<-[rel:MAJOR_MENTIONS]-(content:Content)
		//MATCH (content)-[v:MENTIONS]-(t:Thing)
		//RETURN t.uuid, t.prefLabel, count(t.uuid) ORDER BY count(t.uuid) desc
		// Could we just bring back the top results?
		// Not an already concorded org?

	}

	// Get all mentions for that content

	return []Organisation{}, true, nil
}

func (pcw CypherDriver) isConcordedToTmeAlready(organisationUuid uuid.UUID) (bool, error) {
	results := []struct {
		organisationUuid uuid.UUID
	}{}

	query := &neoism.CypherQuery{
		Statement: `
                        MATCH (organisation:Organisation)<-[]-(:UPPIdentifier{value:{organisationUuid}})
                        MATCH (organisation)<-[]-(:FactsetIdentifier)
			MATCH (organisation)<-[]-(:TMEIdentifier)
			RETURN organisation.uuid
                        `,
		Parameters: neoism.Props{"organisationUuid": organisationUuid.String()},
		Result:     &results,
	}

	err := pcw.conn.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return false, err
	}
	if len(results) == 0 {
		// No results -> Should we just log this? Or do we want to try something else? Could we query the gazeteer for any possibilities?
		return false, nil
	}
	return true, nil

}

type neoReadStruct struct {
	UUID      string `json:"uuid"`
	PrefLabel string `json:"prefLabel"`
}

func (pcw CypherDriver) getSuggestionsForOrganisation(organisationUuid uuid.UUID) ([]Organisation, error) {
	results := []struct {
		rs []neoReadStruct
	}{}

	// As we only have a months worth of V1 suggestions as of december then we don't get too many suggestions back
	// Ideally we would want to only look at the highest count ones which I can't seem to do at the moment
	println("Org id: %v", organisationUuid.String())
	query := &neoism.CypherQuery{
		Statement: `
                        MATCH (organisation:Organisation{uuid:{organisationUuid}})<-[rel:MAJOR_MENTIONS]-(content:Content)
			MATCH (content)-[v:MENTIONS]-(t:Thing)
			RETURN collect({uuid:t.uuid, prefLabel:t.prefLabel}) as rs
                        `,
		Parameters: neoism.Props{"organisationUuid": organisationUuid.String()},
		Result:     &results,
	}

	err := pcw.conn.CypherBatch([]*neoism.CypherQuery{query})

	organisations := []Organisation{}

	fmt.Printf("QUERY: %v", query)
	for _, result := range results[0].rs {
		fmt.Printf("result: %v", result)
		org := Organisation{UUID: result.UUID, PrefLabel: result.PrefLabel}
		organisations = append(organisations, org)
	}
	if err != nil {
		return []Organisation{}, err
	}
	if len(results) == 0 {
		// No results -> Should we just log this? Or do we want to try something else? Could we query the gazeteer for any possibilities?
		return []Organisation{}, nil
	}

	return organisations, nil
}

func (pcw CypherDriver) getAllMajorMentions(contentUuid uuid.UUID) ([]uuid.UUID, error) {

	results := []struct {
		organisationUuids []uuid.UUID
	}{}

	query := &neoism.CypherQuery{
		Statement: `
                        MATCH (organisation:Organisation)<-[rel:MAJOR_MENTIONS]-(content:Content{uuid:{contentUuid}})
			RETURN organisation.uuid as organisationUuids
                        `,
		Parameters: neoism.Props{"contentUuid": contentUuid.String()},
		Result:     &results,
	}

	err := pcw.conn.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		// No results -> Should we just log this? Or do we want to try something else? Could we query the gazeteer for any possibilities?
		return nil, nil
	}
	return results[0].organisationUuids, nil

}
