# Twisp troubleshooting

## HTTP 401 with a null body

The JWT issuer is not registered for the tenant, the token has expired, or the
request is targeting the wrong public Cloud region. Decode the JWT payload
locally and inspect `iss`, `aud`, and `exp`; never paste the complete token into
chat or logs.

## `ACCESS_DENIED`

Authentication succeeded, but the auth client's policies do not permit the
requested operation or resource. Do not retry unchanged. Ask the tenant
administrator to inspect the client's policies.

## `BAD_REQUEST`

Compare the operation against a freshly downloaded schema. For list fields,
confirm that:

- `index` and `first` are present;
- the selected index exists on that resource;
- the index partition key has an equality constraint;
- filter fields belong to the selected index.

## Retrying

GraphQL errors include `extensions.code` and may include `retriableError`. Retry
only when `retriableError` is `true`, using bounded exponential backoff. Syntax,
validation, authentication, and authorization failures require a changed
request or configuration.
