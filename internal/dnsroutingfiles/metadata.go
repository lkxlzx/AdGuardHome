package dnsroutingfiles

// Metadata contains all persisted rule information.
type Metadata struct {
	// DomainListRules are rules loaded from URL sources
	DomainListRules []*DomainListRule `json:"domain_list_rules"`

	// CustomRules are user-defined rules
	CustomRules []*CustomRule `json:"custom_rules"`

	// Version is the metadata format version for future migrations
	Version int `json:"version"`
}

const (
	// metadataVersion is the current metadata format version
	metadataVersion = 1

	// metadataFileName is the name of the metadata file
	metadataFileName = "metadata.json"
)
