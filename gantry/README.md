# Gantry

Provides the control plane for the object store.

## Development Setup

### entr

`entr` is used to automatically restart the server and tests when files change.
Install `entr` with:

```bash
brew install entr
```

### Database Migrations

The project relies on the official `migrate` CLI with the SQLite driver enabled. Build it locally so SQLite support is compiled in:

```bash
CGO_ENABLED=1 go install -tags 'sqlite sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

> The command requires SQLite development headers (e.g., `libsqlite3`) to be installed on your system.
