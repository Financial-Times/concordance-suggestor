package suggestor

import (
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
	"github.com/satori/go.uuid"
)

// Driver interface
type Driver interface {
	Read(contentId uuid.UUID) (concordanceSuggestion []ConcordanceSuggestion, found bool, err error)
	CheckConnectivity() error
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

func (pcw CypherDriver) Read(contentId uuid.UUID) (concordanceSuggestion []ConcordanceSuggestion, found bool, err error) {

	// Get all major mentions
	majorMentionsOrgUuids := getAllMajorMentions(pcw, contentId)

	for _, majorMentionUuid := range majorMentionsOrgUuids {

		// Get all mentions annotations for all content that has that major mentions
		log.Infof("MajorMention UUID: %v", majorMentionUuid)
		//MATCH (organisation:Organisation{uuid:"a18e6c95-e69a-3b3b-ab7d-5f5c7d9bbada"})<-[rel:MAJOR_MENTIONS]-(content:Content)
		//MATCH (content)-[v:MENTIONS]-(t:Thing)
		//RETURN t.uuid, t.prefLabel, count(t.uuid) ORDER BY count(t.uuid) desc
		// Could we just bring back the top results?

	}

	// Get all mentions for that content

	return []ConcordanceSuggestion{}, true, nil
}

// I don't like passing the driver this way!!!
func getAllMajorMentions(pcw CypherDriver, contentUuid uuid.UUID) (uuids []uuid.UUID) {

	results := []struct {
		organisationUuids []uuid.UUID
	}{}

	query := &neoism.CypherQuery{
		Statement: `
                        MATCH (organisation:Organisation)<-[rel:MAJOR_MENTIONS]-(content:Content{uuid})
			RETURN organisation.uuid as organisationUuids
                        `,
		Parameters: neoism.Props{"contentUuid": contentUuid.String()},
		Result:     &results,
	}

	pcw.conn.CypherBatch(query)
	return results

}
