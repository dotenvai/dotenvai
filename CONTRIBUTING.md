# Contributing

## Listing a publisher

A registry submission should:

1. point to `https://<publisher-domain>/.well-known/agent-skills.json`;
2. pass `dotenvai discover <publisher-domain>`;
3. link only to public, versioned skill source;
4. identify licensing for each distributable package;
5. contain no credentials, private endpoints, customer identifiers, or internal
   operating procedures.

Add one entry to `registry/index.json` and include the successful validator
output in the pull request. A maintainer independently verifies that the domain
and linked source represent the named publisher.

Registry inclusion verifies provenance only. It is not a security certification
or endorsement of the publisher's product.

## Protocol changes

For v0, open an issue describing the concrete publisher or client behavior that
the change enables. Protocol additions should be driven by working packages,
not anticipated package-manager features.

## Development

```sh
go test ./...
go vet ./...
go build ./cmd/dotenvai
```
