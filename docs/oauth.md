# OAuth 2.0 Integration Framework

EkaRouter supports official OAuth 2.0 authorization code flows with PKCE (Proof Key for Code Exchange) for third-party developer platforms (such as GitHub, Google Cloud, and Slack).

## OAuth Workflow

1. **Initiate Flow (`POST /api/accounts/oauth/start`):**
   Generates a cryptographically random `state` and PKCE `code_verifier` / `code_challenge`, storing the hashed state and encrypted verifier in `oauth_states`.
2. **User Consent:**
   Client directs user to the provider's official authorization endpoint with the generated state.
3. **Callback Handling (`POST /api/accounts/oauth/callback`):**
   Verifies state hash, exchanges authorization code for access and refresh tokens, and stores encrypted tokens inside `oauth_connections`.
4. **Token Refresh:**
   The background scheduler checks expiring tokens and automatically performs OAuth token refresh before expiry.
