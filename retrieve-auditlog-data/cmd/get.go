package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bytes"
	"context"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// newGetCmd creates the "get" subcommand that fetches and prints audit log records
// from the SAP Audit Log Service v2 API as pretty-printed JSON.
func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get audit log data",
		Run: func(cmd *cobra.Command, args []string) {

			client, err := newClient()
			if err != nil {
				fmt.Println("Error creating OAuth2 client:", err)
				return
			}

			resp, err := client.Get(config.serviceBinding.URL + "/auditlog/v2/auditlogrecords")
			if err != nil {
				fmt.Println("Error making request:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println("Failed to fetch audit logs, status code:", resp.StatusCode)
				return
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return
			}

			var prettyJSON bytes.Buffer
			if err = json.Indent(&prettyJSON, bodyBytes, "", "\t"); err != nil {
				log.Println("JSON formatting failed:", err)
				return
			}
			fmt.Println(prettyJSON.String())
		},
	}
}

// newClient creates an OAuth2 HTTP client using client credentials from the service binding.
// The client automatically fetches and refreshes the bearer token as needed.
func newClient() (*http.Client, error) {
	cfg := &clientcredentials.Config{
		ClientID:     config.serviceBinding.UAA.ClientID,
		ClientSecret: config.serviceBinding.UAA.ClientSecret,
		TokenURL:     config.serviceBinding.UAA.URL + "/oauth/token",
	}

	tokenSource := cfg.TokenSource(context.TODO())
	client := oauth2.NewClient(context.TODO(), tokenSource)

	return client, nil
}
