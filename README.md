<p align="center">
  <img src="assets/dotenvai-logo.png" width="180" alt="dotenv.ai logo">
</p>

# dotenv.ai

Your repository isn't the only place secrets leak anymore.

`dotenvai audit` finds credentials retained in local AI coding-agent transcripts,
scratchpads, tool output, caches, downloads, and workspace artifacts. It supports
Claude Code, Codex, OpenCode, and Cursor.

Everything stays on your machine. SQLite histories are opened read-only, binary
and oversized artifacts are skipped, and reports never include matched values.

## Usage

Build the CLI:

```sh
go build -o dotenvai ./cmd/dotenvai
```

Audit every supported agent:

```sh
./dotenvai audit
```

Select one or more agents:

```sh
./dotenvai audit --agent claude --agent codex
```

Produce a machine-readable report:

```sh
./dotenvai audit --format json
```

Exit codes are `0` when no candidates are found, `1` for usage or runtime errors,
and `2` when the audit finds possible exposures.

## Development

```sh
go test ./...
go vet ./...
go build ./cmd/dotenvai
```

Third-party licenses and attribution are recorded in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
