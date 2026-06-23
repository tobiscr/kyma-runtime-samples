# Audit Log CLI

Audit log CLI is a command-line tool to retrieve audit logs from the [SAP Audit Log Service](https://help.sap.com/docs/btp/sap-business-technology-platform/audit-log-retrieval-api-usage-for-subaccounts-in-cloud-foundry-environment) on SAP Business Technology Platform (BTP).

## What It Does

The tool connects to your SAP BTP Audit Log Service instance and downloads audit log records as JSON. You can filter by time range to narrow down what you retrieve. If the result is large, all pages are fetched and printed automatically.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/) installed
- A service binding file for the `auditlog-management` service instance (see [Audit Log Retrieval API for Global Accounts in the Cloud Foundry Environment](https://help.sap.com/docs/btp/sap-business-technology-platform/audit-log-retrieval-api-for-global-accounts-in-cloud-foundry-environment))

## Setup

Place your service binding file in the project directory as `servicebinding.json`. It should look like this:

```json
{
  "url": "https://<your-auditlog-service>.cfapps.<region>.hana.ondemand.com",
  "tokenUrl": "https://<your-subdomain>.authentication.<region>.hana.ondemand.com",
  "clientId": "<your-client-id>",
  "clientSecret": "<your-client-secret>"
}
```

## Usage

### Get Logs with the Service Default Time Range

```bash
go run main.go get
```

### Get Logs for a Specific Time Range

All times are in UTC, format `YYYY-MM-DDTHH:MM:SS`.

```bash
go run main.go get --time-from 2026-06-01T00:00:00 --time-to 2026-06-22T23:59:59
```

### Get Logs for the Last 15 Minutes

```bash
go run main.go get \
  --time-from $(date -u -v-15M '+%Y-%m-%dT%H:%M:%S') \
  --time-to   $(date -u '+%Y-%m-%dT%H:%M:%S')
```

### Use a Different Binding File

```bash
go run main.go get --bindingFile /path/to/my-binding.json
```

### Save Output to a File

```bash
go run main.go get --time-from 2026-06-01T00:00:00 --time-to 2026-06-22T23:59:59 > logs.json
```

## Flags

| Flag | Description | Default |
|---|---|---|
| `--time-from` | Start of the time range (UTC) | Service default |
| `--time-to` | End of the time range (UTC) | Service default |
| `--bindingFile` / `-b` | Path to the service binding file | `./servicebinding.json` |

## Output

Records are printed as pretty-printed JSON, one page at a time. Each record contains fields such as **message_uuid**, **time**, **category**, **user**, and a **message** field with the full event detail.

Categories you may see:

- `audit.security-events` — login attempts, token issuance
- `audit.configuration` — configuration changes
- `audit.data-access` — data read operations
- `audit.data-modification` — data write or delete operations

## Notes

- The service returns up to 500 records per page. The tool follows pagination automatically so you always get the full result.
- Logs are not immediately available after an event occurs — there may be a short delay before they appear.
- The API rate limit is 4–8 requests per second depending on your region. For large time ranges this is handled transparently.
