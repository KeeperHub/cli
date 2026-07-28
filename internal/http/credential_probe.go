package khhttp

// CredentialProbePath is the endpoint used to decide whether a stored
// credential is actually accepted by the server.
//
// It must be a route that **requires** authentication. This is not a free
// choice: `/api/workflows` and `/api/auth/get-session` both answer 200 to
// anonymous callers - the former resolves auth with `required: false` and
// returns an empty list, the latter returns a null session rather than a 401.
// Probing either one reports every caller as authenticated, including callers
// with no credential at all.
//
// `/api/projects` resolves auth through resolveOrganizationId, which tries
// OAuth, then API key, then session, and answers 401 when none of them
// authenticates. That covers `kh_` API keys, which is what the device-login
// flow now issues.
//
// Both the doctor auth check and API-key validation share this constant so the
// two cannot drift onto different endpoints with different auth semantics.
const CredentialProbePath = "/api/projects"
