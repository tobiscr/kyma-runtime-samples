package cmd

// serviceBinding holds the connection details parsed from the SAP Audit Log Service binding file.
type serviceBinding struct {
	URL string `json:"url"`
	UAA uaa    `json:"uaa"`
}

// uaa holds the OAuth2 credentials for the UAA token endpoint.
type uaa struct {
	URL          string `json:"url"`
	ClientID     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
}
