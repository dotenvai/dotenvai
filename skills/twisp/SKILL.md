---
name: twisp
description: Call the Twisp financial ledger GraphQL API, obtain an AWS identity token, choose a public Twisp Cloud regional endpoint, download the schema SDL, write queries and mutations, and troubleshoot 401, ACCESS_DENIED, and BAD_REQUEST responses. Use when building or debugging an integration with accounts, journals, transactions, balances, entries, account sets, velocity controls, or ACH on Twisp.
compatibility: Requires curl and jq; token helper requires AWS CLI 2.32.0 or newer and a configured AWS identity.
metadata:
  publisher: Twisp
  version: "0.1.0"
---

# Working with the Twisp API

Twisp is a financial ledger exposed through one GraphQL endpoint per region:

```text
POST https://api.<region>.cloud.twisp.com/financial/v1/graphql
authorization: Bearer <JWT>
x-twisp-account-id: <tenant-account-id>
content-type: application/json
```

Supported public Cloud regions are `us-east-1`, `us-east-2`, `us-west-1`,
`us-west-2`, and `ap-southeast-2`.

## Before making a request

Collect three values from the user or their environment:

- `TWISP_REGION`
- `TWISP_ACCOUNT_ID`
- `TWISP_TOKEN`, or an AWS identity capable of minting one

Never print a bearer token or write it into the repository. Before a mutation,
show the target account and the operation's effect. Do not invent non-public
deployment names or endpoints.

## Authentication

Twisp accepts an OIDC JWT whose issuer has been registered as an auth client.
With AWS IAM Outbound Identity Federation:

```sh
aws sts get-web-identity-token \
  --audience ledger-service \
  --signing-algorithm RS256
```

Use the returned `WebIdentityToken` as the bearer token. The command requires
AWS CLI 2.32.0 or newer. `scripts/twisp-token.sh` extracts only the token and
does not cache it.

If the response is HTTP 401 with a null body, verify that the JWT `iss` claim
matches a Twisp auth client's principal and that the client is registered for
the tenant. For authorization and GraphQL errors, read
[`references/troubleshooting.md`](references/troubleshooting.md).

## Inspect the live schema

Prefer the endpoint's schema over guessing field names:

```graphql
query sdl { _service { sdl } }
```

Download it with `scripts/twisp-schema.sh schema.graphql`, then search only the
relevant type, input, or enum.

## Run an operation

`scripts/twisp-gql.sh` accepts an operation file and an optional variables file:

```sh
scripts/twisp-gql.sh operation.graphql variables.json
```

Twisp list queries require a named `index` and `first`. Pagination uses Relay
cursors. Represent monetary values as strings, never floating-point numbers.
One request is one database transaction: if any operation fails, the request is
rolled back.

For the documentation index, fetch `https://www.twisp.com/docs/llms.txt`, choose
the relevant page, then fetch that page as Markdown. Avoid loading the entire
documentation corpus unless a targeted search requires it.
