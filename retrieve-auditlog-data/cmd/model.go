package cmd

// serviceBinding holds the connection details parsed from the SAP Audit Log Service binding file.
type serviceBinding struct {
	URL          string `json:"url"`
	TokenURL     string `json:"tokenUrl"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}
