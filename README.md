# teamtailor-go-sdk-mcp

A Go SDK + MCP server for the [Teamtailor API](https://docs.teamtailor.com/), generated from Protocol Buffer definitions using [sebuf](https://github.com/SebastienMelki/sebuf).

> Status: early scaffolding. Recruitment-core resources first (candidates, jobs, job-applications, stages, users, departments, locations, custom-fields).

## What it ships

- **`pkg/teamtailor`** — type-safe Go client for the Teamtailor API
- **`mcp/cmd/teamtailor-mcp`** — Model Context Protocol stdio server exposing the SDK as tools for any MCP-compatible client
- **`docs/`** — generated OpenAPI 3.1 specs

## Quick start (SDK)

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/sebastienmelki/teamtailor-go-sdk-mcp/pkg/teamtailor"
)

func main() {
    client := teamtailor.NewClient("YOUR_API_KEY")

    resp, err := client.ListCandidates(context.Background(), &teamtailor.ListCandidatesRequest{
        PageSize: 10,
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, c := range resp.Data {
        fmt.Printf("%s %s <%s>\n", c.Attributes.FirstName, c.Attributes.LastName, c.Attributes.Email)
    }
}
```

## Quick start (MCP)

Add to your MCP client config:

```json
{
  "mcpServers": {
    "teamtailor": {
      "command": "/path/to/teamtailor-mcp",
      "env": {
        "TEAMTAILOR_API_KEY": "your_api_key"
      }
    }
  }
}
```

## Architecture

Proto definitions in `teamtailor/v1/` are the single source of truth. Running `make generate` produces:
- Go client code (`internal/gen/`)
- OpenAPI 3.1 specs (`docs/`)

The `pkg/teamtailor` package is a thin hand-written wrapper that configures auth + base URL + API version on the generated client. The MCP server in `mcp/` wraps the SDK with one tool per operation.

See `AGENTS.md` for conventions.

## Development

```bash
make install-tools   # protoc plugins + goimports + golangci-lint
make generate        # buf generate + goimports
make buf-lint        # lint proto files
make lint            # lint Go
make check           # everything
```

## License

MIT (see LICENSE).
