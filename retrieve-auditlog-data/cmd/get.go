package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// newGetCmd creates the "get" subcommand that fetches and prints audit log records
// from the SAP Audit Log Service v2 API as pretty-printed JSON.
//
// Supported flags:
//
//	--time-from  Start of the time range (format: 2006-01-02T15:04:05, UTC). Defaults to 30 days back.
//	--time-to    End of the time range   (format: 2006-01-02T15:04:05, UTC). Defaults to now.
//
// When the result set exceeds 500 records the API paginates via a Paging response header.
// All pages are fetched automatically and printed in order.
func newGetCmd() *cobra.Command {
	var timeFrom, timeTo string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get audit log data",
		Long: `Retrieve audit log records from the SAP Audit Log Service v2 API.

If no time filter is provided the API returns the last 30 days of logs.
Times must be in UTC using the format: 2006-01-02T15:04:05

Example:
  auditlog get --time-from 2024-01-01T00:00:00 --time-to 2024-01-31T23:59:59`,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := newClient()
			if err != nil {
				fmt.Println("Error creating OAuth2 client:", err)
				return
			}

			if err := fetchAllPages(client, timeFrom, timeTo); err != nil {
				fmt.Println("Error fetching audit logs:", err)
			}
		},
	}

	cmd.Flags().StringVar(&timeFrom, "time-from", "", "Start of time range in UTC (format: 2006-01-02T15:04:05)")
	cmd.Flags().StringVar(&timeTo, "time-to", "", "End of time range in UTC (format: 2006-01-02T15:04:05)")

	return cmd
}

// fetchAllPages retrieves all pages of audit log records, following the server-side
// paging handle returned in the Paging response header until no further pages exist.
func fetchAllPages(client *http.Client, timeFrom, timeTo string) error {
	baseURL := config.serviceBinding.URL + "/auditlog/v2/auditlogrecords"
	handle := ""

	for {
		pageURL, err := buildURL(baseURL, timeFrom, timeTo, handle)
		if err != nil {
			return fmt.Errorf("building request URL: %w", err)
		}

		resp, err := client.Get(pageURL)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}

		body, err := readResponse(resp)
		resp.Body.Close()
		if err != nil {
			return err
		}

		var prettyJSON bytes.Buffer
		if err = json.Indent(&prettyJSON, body, "", "\t"); err != nil {
			log.Println("JSON formatting failed:", err)
			return err
		}
		fmt.Println(prettyJSON.String())

		// The API signals more pages via the "Paging" response header: handle=<value>.
		handle = extractHandle(resp.Header.Get("Paging"))
		if handle == "" {
			break
		}
	}
	return nil
}

// buildURL constructs the request URL with optional time_from, time_to, and handle parameters.
func buildURL(baseURL, timeFrom, timeTo, handle string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	if timeFrom != "" {
		q.Set("time_from", timeFrom)
	}
	if timeTo != "" {
		q.Set("time_to", timeTo)
	}
	if handle != "" {
		q.Set("handle", handle)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// extractHandle parses the handle value from the Paging response header (e.g. "handle=abc123").
func extractHandle(pagingHeader string) string {
	const prefix = "handle="
	if len(pagingHeader) > len(prefix) && pagingHeader[:len(prefix)] == prefix {
		return pagingHeader[len(prefix):]
	}
	return ""
}

// readResponse reads and validates the HTTP response body.
// Returns an error for non-200/204 status codes.
func readResponse(resp *http.Response) ([]byte, error) {
	if resp.StatusCode == http.StatusNoContent {
		return []byte("[]"), nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return body, nil
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
