# Agent Skill Discovery Protocol v0.1

Status: experimental.

## Purpose

This protocol lets an organization publish an authoritative list of its Agent
Skills from a domain it controls. It complements the Agent Skills package
specification; it does not redefine `SKILL.md`.

## Discovery

Given a publisher domain `example.com`, a client performs an HTTPS `GET` for:

```text
https://example.com/.well-known/agent-skills.json
```

The server should return `200 OK` with `Content-Type: application/json`. Clients
must reject non-HTTPS URLs in normal operation and should impose response size
and timeout limits. Redirects may be followed, but the final URL must use HTTPS.

The origin serving the well-known document establishes control of the publisher
domain. It does not establish that every linked package is safe. Clients should
display the final artifact origin and verify `sha256` before installation.

## Document

The normative machine-readable definition is
[`schema/discovery-v0.1.schema.json`](../schema/discovery-v0.1.schema.json).

```json
{
  "schema_version": "0.1",
  "publisher": {
    "name": "Example, Inc.",
    "url": "https://example.com"
  },
  "skills": [
    {
      "name": "example",
      "description": "Integrate with the Example API.",
      "version": "1.0.0",
      "source": "https://github.com/example/skills/tree/example-v1.0.0/example",
      "archive": "https://github.com/example/skills/releases/download/example-v1.0.0/example.tar.gz",
      "sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
      "license": "Apache-2.0",
      "compatibility": ["claude-code", "codex"]
    }
  ]
}
```

### Required fields

- `schema_version`: exactly `0.1` for this revision.
- `publisher.name`: publisher's human-readable name.
- `publisher.url`: canonical HTTPS homepage.
- `skills`: one or more skill records, with unique names.
- `skills[].name`: must conform to the Agent Skills name rules.
- `skills[].description`: discovery description, 1–1024 characters.
- `skills[].version`: semantic version without a leading `v`.
- `skills[].source`: HTTPS URL for human-inspectable, versioned source.

### Optional fields

- `archive`: HTTPS URL for an immutable `.tar.gz` or `.zip` artifact.
- `sha256`: lowercase SHA-256 digest. Required when `archive` is present.
- `license`: SPDX identifier or concise license reference.
- `compatibility`: identifiers for clients or environments verified by the
  publisher. Absence means portable, not universally tested.

Unknown fields are rejected in v0.1 so mistakes surface early. New semantics
will use a new `schema_version`.

## Versioning and updates

Publishers should use semantic versions and immutable source tags. Changing a
package without changing its version is invalid. Consumers should pin the
resolved version and digest in their own lock data rather than silently tracking
the publisher's latest release.

## Registry relationship

The dotenv.ai registry stores a discovery URL and verification status, not the
canonical package. Removal from the index does not prevent direct discovery.
Likewise, inclusion does not imply a security audit or endorsement.

## Security considerations

Skills are active instructions and may include executable scripts. Discovery
clients must not execute code while indexing, validating, or rendering a skill.
Installers should show package contents and dependencies, verify hashes, prevent
path traversal during extraction, and require confirmation before replacing an
existing package.
