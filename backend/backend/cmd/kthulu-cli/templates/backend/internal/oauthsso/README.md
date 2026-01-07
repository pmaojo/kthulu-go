# OAuth SSO Module (Template)

The template ships the same storage wiring as the runtime module. `adapters.FositeStorage` adapts the Fosite callbacks onto the repositories exposed by your service:

- **SessionRepository** is responsible for authorization codes, PKCE requests, and other session-oriented flows.
- **TokenRepository** handles access/refresh token persistence and revocation using the identifiers passed in by Fosite.

To extend how OAuth data is stored, focus on these repositories. The adapter only forwards calls, so once the repositories understand a new persistence rule, every Fosite flow automatically benefits.
