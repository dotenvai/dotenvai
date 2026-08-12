# dotenv.ai product boundary

dotenv.ai catches credentials in code changes before they become incidents.

The open-source product is deliberately small:

1. `dotenvai` scans added diff lines locally with Gitleaks' maintained provider
   fingerprints plus dotenv.ai-specific contextual rules.
2. The Docker-based GitHub Action runs the same binary and emits inline annotations.

The scanner never includes matched values in its output. The default CLI makes no
network requests. Hosted semantic classification, organization policy, credential
verification, and rotation workflows are possible paid extensions, not requirements
for the core product.

## Not in v0

- Secret storage or delivery
- Agent sandboxes and capability brokers
- Automatic credential verification or revocation
- Whole-history scanning
- Billing and organization administration
- LLM calls from the default scanner

## Validation

The primary signal is a retained installation on a real repository. Track time to
first scan, confirmed findings, suppressions, 30-day retention, and requests for
central policy or hosted classification.
