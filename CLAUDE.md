# edinet-cli Development Guide

## Project Overview

EDINET API v2 CLI wrapper in Go. AI agent-first design.

## Build & Test

```bash
make build    # go build -o edinet-cli .
make test     # go test -race ./...
make lint     # go vet + golangci-lint
make coverage # coverage report
```

## TDD Rules (Strictly Enforced)

1. **Test first**: Write `*_test.go` before implementation
2. **RED → GREEN → REFACTOR**: One test at a time
3. **Tests are immutable**: Fix implementation to match tests, never the reverse
4. **External APIs**: Always mock with `httptest.NewServer`
5. **File I/O**: Always use `t.TempDir()` for isolation
6. **`-race` always on**: All test runs include `-race`
7. **Naming**: `Test<Function>_<Scenario>` (e.g., `TestClient_Get_WithExpiredToken`)
8. **Table-driven tests**: Preferred for multiple scenarios

## Architecture

```
cmd/         → Thin CLI adapters (cobra commands)
internal/
  api/       → EDINET API wire models & HTTP client
  cache/     → Cache interface + implementations
  config/    → Config loading (env vars, paths)
  company/   → EDINET code list registry & search
  extract/   → ZIP/CSV/HTML extraction
  output/    → JSON/table formatters
  service/   → Business logic (document listing, downloads, company search)
  schema/    → CLI self-description data
  testutil/  → Shared test helpers (httptest mocks)
```

## Key Design Decisions

### API Key
- **Environment variable only**: `EDINET_API_KEY`
- Never read from config file, never written to disk
- Passed as `Subscription-Key` query parameter (not header)

### Error Handling
- EDINET API always returns HTTP 200. Errors are in JSON body
- Two different error JSON formats:
  - Normal (400/404/500): `{"metadata": {"status": "400", "message": "..."}}`
  - Auth (401): `{"StatusCode": 401, "message": "..."}` (int StatusCode, uppercase S/C)
- Download API: check `Content-Type` header (`application/json` = error)
- Use `mime.ParseMediaType()` for Content-Type normalization

### Error Codes (unified taxonomy)
```
AUTH_REQUIRED, AUTH_FAILED, BAD_REQUEST, NOT_FOUND,
SERVER_ERROR, NETWORK_ERROR, TIMEOUT, VALIDATION_ERROR, INTERNAL_ERROR
```

### Exit Codes
- 0: Success (including partial-success with warnings)
- 1: General failure
- 2: Validation/argument error
- 3: Authentication error
- 4: API error (400/404/500)

### Output Contract
- stdout: Always structured data (JSON default, `--format table` for humans)
- stderr: Errors (JSON), progress, debug info
- `os.Exit()` only in `main.go`. Commands return errors.

### Nullable Fields
- API wire model (`api.Document`): All fields are `*string` (EDINET nulls many fields when viewing period expires)
- CLI output DTO (`service.DocumentInfo`): `string` + `omitempty`, flags converted to `bool`

### Cache
- Interface: `cache.Cache` with `ErrCacheMiss` sentinel
- Keys: `doclist/{date}.json`, `files/{docID}/{type}`, `codelist/edinetcode.csv`
- Atomic writes: temp file + `os.Rename()`
- Only cache successful responses

### Dependencies
- `App` struct for dependency injection (not `context.WithValue`)
- Services receive dependencies via constructor
