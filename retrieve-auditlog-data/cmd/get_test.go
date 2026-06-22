package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// auditLogRecord mirrors the structure returned by the SAP Audit Log Service v2 API.
// The message field is a JSON-encoded string containing the full event detail.
type auditLogRecord struct {
	MessageUUID    string `json:"message_uuid"`
	Time           string `json:"time"`
	Tenant         string `json:"tenant"`
	OrgID          string `json:"org_id"`
	SpaceID        string `json:"space_id"`
	AppOrServiceID string `json:"app_or_service_id"`
	AlsServiceID   string `json:"als_service_id"`
	User           string `json:"user"`
	Category       string `json:"category"`
	FormatVersion  string `json:"format_version"`
	Message        string `json:"message"`
}

// auditLogMessage mirrors the structure of the JSON-encoded message field.
type auditLogMessage struct {
	UUID          string                 `json:"uuid"`
	User          string                 `json:"user"`
	Time          string                 `json:"time"`
	IP            string                 `json:"ip"`
	Data          string                 `json:"data"`
	Attributes    []any          `json:"attributes"`
	ID            string         `json:"id"`
	Category      string         `json:"category"`
	Tenant        string         `json:"tenant"`
	CustomDetails map[string]any `json:"customDetails"`
}

// newTestRecord builds a synthetic audit log record with the given category and timestamp.
// All IDs and values are generic placeholders — no real data from test-data.json.
func newTestRecord(category, timestamp string) auditLogRecord {
	msg := auditLogMessage{
		UUID:          "00000000-0000-0000-0000-000000000001",
		User:          "test-user",
		Time:          timestamp,
		IP:            "192.0.2.1",
		Data:          `{"level":"INFO","message":"test event"}`,
		Attributes:    []any{},
		ID:            "00000000-0000-0000-0000-000000000002",
		Category:      category,
		Tenant:        "00000000-0000-0000-0000-000000000003",
		CustomDetails: map[string]any{},
	}
	msgBytes, _ := json.Marshal(msg)

	return auditLogRecord{
		MessageUUID:    "00000000-0000-0000-0000-000000000001",
		Time:           timestamp,
		Tenant:         "00000000-0000-0000-0000-000000000003",
		OrgID:          "00000000-0000-0000-0000-000000000004",
		SpaceID:        "00000000-0000-0000-0000-000000000005",
		AppOrServiceID: "00000000-0000-0000-0000-000000000006",
		AlsServiceID:   "00000000-0000-0000-0000-000000000007",
		User:           "test-user",
		Category:       category,
		FormatVersion:  "",
		Message:        string(msgBytes),
	}
}

func TestBuildURL_WithTimestamps(t *testing.T) {
	base := "https://auditlog.example.com/auditlog/v2/auditlogrecords"

	result, err := buildURL(base, "2024-01-01T00:00:00", "2024-01-31T23:59:59", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u, _ := url.Parse(result)
	q := u.Query()

	if got := q.Get("time_from"); got != "2024-01-01T00:00:00" {
		t.Errorf("time_from = %q, want %q", got, "2024-01-01T00:00:00")
	}
	if got := q.Get("time_to"); got != "2024-01-31T23:59:59" {
		t.Errorf("time_to = %q, want %q", got, "2024-01-31T23:59:59")
	}
}

func TestBuildURL_WithoutTimestamps(t *testing.T) {
	base := "https://auditlog.example.com/auditlog/v2/auditlogrecords"

	result, err := buildURL(base, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u, _ := url.Parse(result)
	q := u.Query()

	if q.Has("time_from") {
		t.Error("time_from should not be present when empty")
	}
	if q.Has("time_to") {
		t.Error("time_to should not be present when empty")
	}
}

func TestBuildURL_WithHandle(t *testing.T) {
	base := "https://auditlog.example.com/auditlog/v2/auditlogrecords"
	handle := "dGVzdC1oYW5kbGU="

	result, err := buildURL(base, "2024-01-01T00:00:00", "2024-01-31T23:59:59", handle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u, _ := url.Parse(result)
	q := u.Query()

	if got := q.Get("handle"); got != handle {
		t.Errorf("handle = %q, want %q", got, handle)
	}
}

func TestExtractHandle(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{"handle=dGVzdA==", "dGVzdA=="},
		{"handle=", ""},
		{"", ""},
		{"something-else", ""},
	}

	for _, tt := range tests {
		got := extractHandle(tt.header)
		if got != tt.want {
			t.Errorf("extractHandle(%q) = %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestReadResponse_OK(t *testing.T) {
	records := []auditLogRecord{newTestRecord("audit.security-events", "2024-01-15T10:00:00.000Z")}
	body, _ := json.Marshal(records)

	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)
	rec.Write(body)

	got, err := readResponse(rec.Result())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []auditLogRecord
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if len(parsed) != 1 {
		t.Errorf("expected 1 record, got %d", len(parsed))
	}
	if parsed[0].Category != "audit.security-events" {
		t.Errorf("category = %q, want %q", parsed[0].Category, "audit.security-events")
	}
}

func TestReadResponse_NoContent(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusNoContent)

	got, err := readResponse(rec.Result())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "[]" {
		t.Errorf("body = %q, want %q", string(got), "[]")
	}
}

func TestReadResponse_ErrorStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusUnauthorized)
	fmt.Fprint(rec, "unauthorized")

	_, err := readResponse(rec.Result())
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestFetchAllPages_WithTimestamps(t *testing.T) {
	records := []auditLogRecord{
		newTestRecord("audit.security-events", "2024-01-10T08:00:00.000Z"),
		newTestRecord("audit.configuration", "2024-01-15T12:30:00.000Z"),
	}
	body, _ := json.Marshal(records)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if got := q.Get("time_from"); got != "2024-01-01T00:00:00" {
			http.Error(w, fmt.Sprintf("unexpected time_from: %q", got), http.StatusBadRequest)
			return
		}
		if got := q.Get("time_to"); got != "2024-01-31T23:59:59" {
			http.Error(w, fmt.Sprintf("unexpected time_to: %q", got), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	config.serviceBinding.URL = server.URL

	if err := fetchAllPages(server.Client(), "2024-01-01T00:00:00", "2024-01-31T23:59:59"); err != nil {
		t.Fatalf("fetchAllPages returned error: %v", err)
	}
}

func TestFetchAllPages_Pagination(t *testing.T) {
	page1 := []auditLogRecord{newTestRecord("audit.security-events", "2024-01-10T08:00:00.000Z")}
	page2 := []auditLogRecord{newTestRecord("audit.data-access", "2024-01-20T14:00:00.000Z")}
	body1, _ := json.Marshal(page1)
	body2, _ := json.Marshal(page2)

	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")

		if page == 1 {
			w.Header().Set("Paging", "handle=dGVzdC1oYW5kbGU=")
			w.WriteHeader(http.StatusOK)
			w.Write(body1)
			return
		}

		if got := r.URL.Query().Get("handle"); got != "dGVzdC1oYW5kbGU=" {
			http.Error(w, fmt.Sprintf("missing or wrong handle: %q", got), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body2)
	}))
	defer server.Close()

	config.serviceBinding.URL = server.URL

	if err := fetchAllPages(server.Client(), "2024-01-01T00:00:00", "2024-01-31T23:59:59"); err != nil {
		t.Fatalf("fetchAllPages returned error: %v", err)
	}
	if page != 2 {
		t.Errorf("expected 2 pages fetched, got %d", page)
	}
}
