# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

## Project overview

Go SDK + MCP server for the [Teamtailor API](https://docs.teamtailor.com/), built with [sebuf](https://github.com/SebastienMelki/sebuf). Proto definitions in `teamtailor/v1/` are the single source of truth; everything else (`internal/gen/`, `docs/`) is generated. The MCP server in `mcp/` is hand-written using the official Go MCP SDK and wraps the generated SDK 1:1.

## Build commands

```bash
make install-tools     # protoc plugins + goimports + golangci-lint
make generate          # buf generate + goimports on internal/gen/
make buf-lint          # lint proto files
make lint              # golangci-lint
make lint-fix          # golangci-lint --fix
make build             # go build ./...
make test              # go test ./...
make check             # buf-lint + lint + generate + build + test
make apitest           # run cmd/apitest against live Teamtailor API
make mcp-build         # build the stdio MCP binary
make release VERSION=v0.1.0   # tag + push
```

## Architecture

### Protobuf-first design

All API definitions live in `.proto` files under `teamtailor/v1/`. The sebuf generators produce:
- A typed Go HTTP client (`protoc-gen-go-client`)
- Request validation hooks (`buf.validate`)
- OpenAPI 3.1 documentation (`protoc-gen-openapiv3`)

### Teamtailor API specifics

Teamtailor uses **JSON:API** (`application/vnd.api+json`):
- Hyphenated attribute names — model with `[json_name = "first-name"]` on proto fields
- `data` / `attributes` envelope — model explicitly as proto wrapper messages
- Cursor pagination via `page[after]`, `page[before]`, `page[size]` (max 30)
- Sparse fieldsets: `fields[<type>]=name,email`
- Relationship sideloading: `include=department,location`

Auth is `Authorization: Token token=<API_KEY>` + `X-Api-Version: <YYYYMMDD>` (Teamtailor uses dated versions; default `20240904`).

Region-specific base URLs:
- EU: `https://api.teamtailor.com`
- NA: `https://api.na.teamtailor.com`
- AU: `https://api.au.teamtailor.com`

### Proto file organization

```
teamtailor/v1/
├── service.proto              # TeamtailorService + service_headers + all RPC declarations
├── common.proto               # Links, Meta, Relationship, JsonApiError(Response)
├── candidate.proto            # Candidate + CandidateAttributes
├── job.proto                  # Job + JobAttributes
├── ...                        # one model file per resource
├── list_candidates.proto      # one file per RPC, holding both Request and Response
├── get_candidate.proto
└── ...
```

### Generated layout

```
internal/gen/teamtailor/v1/    # generated Go client (private; consumers go through pkg/)
pkg/teamtailor/                # public wrapper: NewClient, options, type aliases
mcp/cmd/teamtailor-mcp/        # MCP stdio binary entrypoint
mcp/                           # MCP server bootstrap + tool registration
cmd/apitest/                   # manual integration harness
docs/                          # generated OpenAPI 3.1
```

## Key patterns

### Service-level headers

```proto
service TeamtailorService {
  option (sebuf.http.service_config) = { base_path: "/v1" };

  option (sebuf.http.service_headers) = {
    required_headers: [
      { name: "Authorization",  description: "Token token=<API_KEY>",        type: "string", required: true },
      { name: "X-Api-Version",  description: "Dated API version YYYYMMDD",   type: "string", required: true, example: "20240904" },
      { name: "Content-Type",   description: "application/vnd.api+json",     type: "string" },
      { name: "Accept",         description: "application/vnd.api+json",     type: "string" }
    ]
  };
}
```

The `pkg/teamtailor` wrapper formats `"Token token=" + apiKey` and passes it via the generated `WithTeamtailorServiceAuthorization` option.

### HTTP methods and paths

```proto
rpc ListCandidates(ListCandidatesRequest) returns (ListCandidatesResponse) {
  option (sebuf.http.config) = { path: "/candidates", method: HTTP_METHOD_GET };
}

rpc GetCandidate(GetCandidateRequest) returns (GetCandidateResponse) {
  option (sebuf.http.config) = { path: "/candidates/{id}", method: HTTP_METHOD_GET };
}
```

### JSON:API attribute names (hyphenated keys)

```proto
message CandidateAttributes {
  string first_name = 1 [json_name = "first-name"];
  string last_name  = 2 [json_name = "last-name"];
  string email      = 3;
  string created_at = 4 [json_name = "created-at"];
}
```

### Bracketed query parameters

```proto
message ListCandidatesRequest {
  int32  page_size   = 1 [(sebuf.http.query) = { name: "page[size]" }];
  string page_after  = 2 [(sebuf.http.query) = { name: "page[after]" }];
  string filter_email = 3 [(sebuf.http.query) = { name: "filter[email]" }];
  string include     = 4 [(sebuf.http.query) = { name: "include" }];
  string sort        = 5 [(sebuf.http.query) = { name: "sort" }];
}
```

### Validation annotations

```proto
message GetCandidateRequest {
  string id = 1 [
    (buf.validate.field).required = true,
    (sebuf.http.field_examples) = { values: ["12345"] }
  ];
}
```

### Error messages

Messages whose name ends in `Error` auto-implement Go's `error` interface (sebuf convention):

```proto
message JsonApiError {
  string status = 1;
  string title  = 2;
  string detail = 3;
}
```

## Conventions

- One proto file per RPC operation (request + response together).
- One proto file per resource model (e.g. `candidate.proto` holds `Candidate`, `CandidateAttributes`, related enums).
- All resource attributes go in a `{Resource}Attributes` nested message — mirrors JSON:API's `data.attributes` envelope.
- Commits use Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`).
- Generated code lives in `internal/gen/` — never edit by hand.
- Run `make check` before committing.

## API testing

```bash
# .env (gitignored)
TEAMTAILOR_API_KEY=...
TEAMTAILOR_REGION=eu            # eu|na|au, defaults to eu
TEAMTAILOR_API_VERSION=20240904 # defaults to 20240904

make apitest
```

## References

- Teamtailor docs: https://docs.teamtailor.com/
- sebuf: https://github.com/SebastienMelki/sebuf
- JSON:API spec: https://jsonapi.org/
