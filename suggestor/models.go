package suggestor

// Thing is the base entity, all nodes in neo4j should have these properties
type Thing struct {
	ID        string `json:"id"`
	APIURL    string `json:"apiUrl"` // self ?
	PrefLabel string `json:"prefLabel,omitempty"`
}

type ConcordanceSuggestion struct {
	V1MajorMentions           Organisation   `json:"V1MajorMentions"`
	V2OrganisationSuggestions []Organisation `json:"V2OrganisationSuggestions,omitempty"`
}

// Organisation
type Organisation struct {
	*Thing
}
