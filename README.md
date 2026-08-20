<p align="center">
  <img src="assets/dotenvai-logo.png" width="160" alt="dotenv.ai logo">
</p>

# dotenv.ai

**Official skills for the products agents work with.**

dotenv.ai is an open registry and discovery protocol for company-authored
[Agent Skills](https://agentskills.io). It gives publishers a domain-verified way
to tell agents which skills are official, while keeping the skill packages in
ordinary, inspectable repositories.

This repository contains:

- the v0 discovery protocol and JSON Schema;
- a curated index of publishers;
- reference skill packages, beginning with Twisp;
- `dotenvai`, a CLI for validating skills and discovery documents.

## The model

A skill remains a normal directory containing `SKILL.md` and optional scripts,
references, and assets. A publisher exposes a discovery document at:

```text
https://example.com/.well-known/agent-skills.json
```

The document points to versioned source and downloadable artifacts. dotenv.ai
indexes it; the publisher remains the authority.

```text
publisher domain ──> .well-known/agent-skills.json ──> versioned skill package
                              │
                              └── indexed by dotenv.ai
```

Read [the v0 protocol](docs/discovery.md) or inspect the
[Twisp example](examples/twisp.agent-skills.json).

## CLI

Requires Go 1.23 or newer.

```sh
go build -o dotenvai ./cmd/dotenvai
./dotenvai validate skills/twisp
./dotenvai validate-manifest examples/twisp.agent-skills.json
./dotenvai discover example.com
```

Validation follows the Agent Skills naming and frontmatter constraints. The
discovery command requires HTTPS, limits response size, and never executes code
from a package.

## Repository layout

```text
cmd/dotenvai/           CLI
internal/               validator and discovery implementation
docs/discovery.md       protocol specification
schema/                 machine-readable discovery schema
registry/index.json     curated publisher index
skills/                 reference packages
examples/               example discovery documents
```

## Publishing

1. Publish a conforming skill in a public, versioned repository.
2. Host an `agent-skills.json` document on your company domain.
3. Run `dotenvai discover your-company.example`.
4. Open a PR adding the discovery URL to `registry/index.json`.

The registry is intentionally an index, not a package host. See
[CONTRIBUTING.md](CONTRIBUTING.md) for acceptance criteria.

## Status

The discovery protocol is an experimental v0. We intend to learn from real
company packages before freezing a v1.
