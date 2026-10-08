# Turnstile validation

`verify.go` executes documented Cloudflare Siteverify POST requests through the
existing bounded upstream queue. Successful tokens must match the configured
panel hostname and `txc_first_entry` action. Tokens are single-use and expire at
Cloudflare after five minutes; neither tokens nor secrets are persisted here.
Redirects, malformed responses, provider outages and invalid configuration fail
closed. Only sanitized sentinel errors leave this module. The injected HTTP
client supports hosted wire-contract regression checks in `verify_test.go`.

Reference: https://developers.cloudflare.com/turnstile/get-started/server-side-validation/
