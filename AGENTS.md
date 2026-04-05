# AGENTS.md

## Project Overview

A Go CLI tool and web server for converting backup files between two AI chat applications: [Kelivo](https://github.com/nicepkg/kelivo) (Flutter) and [RikkaHub](https://github.com/rikka/rikkahub) (Android/Kotlin). Converts providers, assistants, MCP servers, world books/lorebooks, quick messages, instruction injections, and full chat history with branching/versions.

The UI and error messages are in Chinese (中文). Keep them in Chinese when modifying.

## Build & Run

```bash
# Build (requires Go 1.21+, CGO needed for SQLite)
go build -o backup-converter .

# CLI usage
./backup-converter k2r -i kelivo_backup.zip -o output.zip      # Kelivo → RikkaHub
./backup-converter r2k -i rikkahub_backup.zip -o output.zip    # RikkaHub → Kelivo
./backup-converter info -i backup.zip                           # Inspect backup
./backup-converter serve -p 8080                                # Web server mode

# Interactive shell script (builds if needed)
./run.sh
```

The CI workflow (`.github/workflows/build.yml`) cross-compiles with `CGO_ENABLED=0` and `go build -ldflags="-s -w"` for 6 OS/arch combos. It triggers on `v*` tags and auto-creates a GitHub Release with binaries.

## Architecture

```
main.go          → CLI entry, arg parsing, command dispatch
server.go        → HTTP server (embeds web/index.html via go:embed)
models/          → Data structures for both formats (no logic)
  kelivo.go      → Kelivo types (settings, providers, assistants, messages...)
  rikkahub.go    → RikkaHub types (settings, providers, assistants, message nodes...)
parser/          → Deserialize backup zips into models
  kelivo_parser.go   → Reads settings.json + chats.json from zip
  rikkahub_parser.go → Reads settings.json + SQLite DB from zip (with WAL support)
converter/       → Bidirectional conversion logic (pure data transformation, no I/O)
  kelivo_to_rikkahub.go  → K→R conversion
  rikkahub_to_kelivo.go  → R→K conversion
  builtins.go            → Built-in provider ID mappings between the two apps
writer/          → Serialize models back to backup zips
  kelivo_writer.go   → Writes settings.json + chats.json to zip
  rikkahub_writer.go → Creates fresh SQLite DB + writes to zip
```

### Data Flow

**CLI path:** `parser.Parse*Backup(zip) → converter.Convert*(backup) → writer.Write*Backup(result, zip)`

**Web server path:** Upload → parse to `models.*Backup` → store in `Server` struct (mutex-protected) → API endpoints read/transform → convert + marshal on demand → download as zip.

### Key Architectural Differences Between Formats

- **Kelivo** stores chats as flat JSON: `chats.json` with `conversations[]` and `messages[]`, using `groupId` + `version` fields for branching.
- **RikkaHub** stores chats in SQLite (`rikka_hub.db`): `ConversationEntity` and `message_node` tables. Each node contains an array of messages (versions). Messages within nodes have a `select_index` for the active version.
- **Kelivo settings** uses double-serialization: complex fields (providers, assistants, MCP servers, etc.) are JSON strings stored as top-level string fields (e.g., `provider_configs_v1`). The parser deserializes these into the `json:"-"` tagged struct fields. The writer re-serializes them back.
- **RikkaHub settings** uses normal nested JSON structures.

## Critical Gotchas

### Built-in Provider Mapping (`converter/builtins.go`)

Both apps ship with preset providers (OpenAI, DeepSeek, Gemini, etc.) that share API endpoints but have different internal IDs. During conversion:

- **K→R direction**: Matched by `(baseUrl, providerType)`. Only apiKey and user-added models are transferred; the built-in provider's RikkaHub UUID is used.
- **R→K direction**: Matched by RikkaHub provider UUID. Maps to the corresponding Kelivo built-in provider key.

When adding new built-in providers, add entries to the `builtinProviderMappings` slice in `builtins.go`.

### Model ID Mapping

Models are identified differently in each app:
- **Kelivo**: `"ProviderID::modelId"` format (e.g., `"OpenAI::gpt-4o"`)
- **RikkaHub**: Each model gets a UUID. A `modelIDMap` is built during conversion to translate between formats.

The mapping resolution is fuzzy — falls back to matching by model ID suffix if the full key isn't found.

### Kelivo's Double-Serialized Settings

`models/kelivo.go` has paired fields: a `string` "raw" field (e.g., `ProviderConfigsRaw`) that holds JSON text deserialized from `settings.json`, and a `json:"-"` parsed field (e.g., `ProviderConfigs`) populated by the parser. The writer reverses this. When constructing `KelivoSettings` in converter code, you must populate **both** the raw and parsed fields (or the writer will handle the re-serialization if only parsed fields are set).

### SQLite for RikkaHub

RikkaHub backups contain `rikka_hub.db` and optionally `rikka_hub-wal` / `rikka_hub-shm`. The parser writes the DB to a temp file, opens it with `mattn/go-sqlite3`, and does a WAL checkpoint before reading. The writer creates a fresh SQLite DB with the expected schema. **CGO is required** because of the SQLite dependency.

### Web UI is Embedded

`server.go` uses `//go:embed web/index.html` to embed the frontend. The web UI is a single HTML file using React (via CDN/Babel standalone), TailwindCSS, and Lucide icons — no build step.

### Attachment Format

Kelivo uses inline tags in message content for attachments: `[image:<path>]` and `[file:<path>|<name>|<mime>]`. RikkaHub uses structured `RikkaHubMessagePart` objects with `type`, `url`, etc. The converter regex-parses these back and forth.

## Dependencies

- `github.com/google/uuid` — generating UUIDs for new model/assistant IDs
- `github.com/mattn/go-sqlite3` — SQLite driver (requires CGO)
- Go standard library only otherwise

## Build, Lint & Test Commands

```bash
# Build
go build -o backup-converter .

# Build with flags (used in CI for smaller binaries)
CGO_ENABLED=0 go build -ldflags="-s -w" -o backup-converter .

# Run
./backup-converter --help

# Format code (uses gofmt)
go fmt ./...

# Vet (static analysis)
go vet ./...

# Run all checks (fmt + vet + build)
go build . && go fmt ./... && go vet ./...

# Run a specific test (if tests exist)
go test -v -run TestName ./...

# Run all tests (if tests exist)
go test ./...

# Interactive build script
./run.sh
```

## Code Style Guidelines

### General Principles

- **Go 1.21+ required**, Go 1.24.4 used in development
- **CGO required** for SQLite (github.com/mattn/go-sqlite3)
- Use standard Go tooling (`go fmt`, `go vet`)
- Keep code idiomatic — follow Go conventions, not other languages

### Imports

```go
import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/converter/backup-converter/converter"
    "github.com/converter/backup-converter/parser"
    "github.com/converter/backup-converter/writer"
)
```

- Group imports: stdlib first, then third-party, then project-local
- Use named imports for packages used multiple times; can omit for one-off use
- No blank imports unless necessary

### Formatting

- Run `go fmt ./...` before committing
- Use tabs for indentation (Go default)
- Max line length ~100 chars; break long lines at logical points
- No trailing whitespace
- Add blank lines between logical sections (e.g., between functions, between import groups)

### Naming Conventions

- **Variables/functions**: `camelCase` (e.g., `parseArgs`, `rikkaBackup`)
- **Constants**: `PascalCase` or `UPPER_SNAKE_CASE` for exported, `camelCase` for unexported
- **Types/structs**: `PascalCase` (e.g., `KelivoSettings`, `RikkaHubBackup`)
- **Packages**: short, lowercase, no underscores (e.g., `parser`, `converter`)
- **Files**: `snake_case.go` (e.g., `kelivo_parser.go`, `rikkahub_writer.go`)
- **Acronyms**: preserve case (e.g., `ID`, `URL`) not `id`, `url`
- Be descriptive: `handleKelivo2RikkaHub` is fine, avoid `h2r`

### Types & Structs

- Prefer concrete types over interfaces unless needed
- Use struct tags for JSON: `json:"field_name"`, `json:"-"` for skipped fields
- Add `omitempty` for optional fields: `json:"field_name,omitempty"`
- Document exported types with comments:

```go
// KelivoSettings holds the parsed settings from a Kelivo backup.
type KelivoSettings struct {
    // ...
}
```

### Error Handling

- Return errors rather than using global state
- Wrap errors with context using `fmt.Errorf("description: %w", err)`:

```go
if err != nil {
    return nil, fmt.Errorf("解析 Kelivo 备份失败: %w", err)
}
```

- Use `os.Exit(1)` in `main()` for fatal errors; return errors from library functions
- Chinese error messages in CLI/main code are acceptable (UI uses Chinese)

### Functions

- Keep functions focused and reasonably sized
- Unexported functions can be shorter; exported should be documented
- Use receiver methods on types when behavior relates to a type:

```go
func (s *Server) HandleUpload(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

### Concurrency

- Use mutexes (`sync.Mutex`) to protect shared state (see `server.go`)
- Always `defer` unlock if using `Lock()`:

```go
func (s *Server) DoSomething() {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ...
}
```

- Use goroutines only when appropriate (background tasks, parallel processing)

### Database (SQLite)

- Use parameterized queries to prevent SQL injection
- Close resources with `defer`:

```go
defer rows.Close()
```

- WAL checkpoint after writes (see `parser/rikkahub_parser.go`)

### Testing

- Tests should go in `*_test.go` files alongside the code they test
- Use `t.Fatalf` or `t.Error` for assertions
- Table-driven tests are encouraged for multiple cases:

```go
func TestParseArgs(t *testing.T) {
    tests := []struct {
        name string
        args []string
        // ...
    }{
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

## Directory Layout Note

The `code/` directory contains full source trees of the Kelivo (Flutter) and RikkaHub (Kotlin/Android) applications for reference. These are not compiled or used by the Go project — they exist for understanding the source/target data formats.
