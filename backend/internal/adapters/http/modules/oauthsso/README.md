# OAuth SSO Module

The OAuth SSO module wires the domain repositories into Ory Fosite. The `adapters.FositeStorage` type is the single entry point that Fosite interacts with and is responsible for delegating persistence to the module repositories:

- **SessionRepository** stores high-level request sessions such as authorization codes and PKCE exchanges. Extending this repository lets you adjust how those flows are stored or invalidated.
- **TokenRepository** persists opaque tokens such as access or refresh signatures and handles revocation. Any changes to token life-cycle or how request IDs map to stored signatures belong here.

When you need to support new grant types or change storage semantics, update the repositories first and the adapter will automatically start using the new behavior because every Fosite storage callback simply forwards to these abstractions.
