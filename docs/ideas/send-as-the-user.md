# Send messages as the user

Status: researched, 2026-10-29.

## Idea

Let a step send its message through the chat service for the person, instead of
only copying the text. [0065](../decisions/0065-messages-and-evidence.md)
already designs two modes. `copy` puts the text on the clipboard and opens the
channel. `user` sends as the person's own account after a one-time
authorization. It rules out a bot mode. This note covers how `user` mode could
get and keep a Slack token.

## What the Slack documentation says

- **Sign in without a client secret.** PKCE, a way for a public client to prove
  its identity without a secret, has been generally available since March 2026.
  The tool sends a hash of a one-time secret in the authorization URL and then
  the secret itself in place of `client_secret`. A `localhost` redirect counts
  as a desktop redirect when the app opts in. The guide's sample manifest
  requests the user scope `chat:write` with `pkce_enabled: true`, and its
  example URL includes `user_scope=chat:write`.
  - Enabling PKCE on an app cannot be undone without contacting Slack.
  - Refresh tokens for such apps expire after 30 days.
  - The documentation does not say whether a workspace administrator must
    approve the app. Assume some will.
- **Bring your own token.** After a person installs their own app to a
  workspace, the OAuth & Permissions page shows an access token under OAuth
  Tokens. The page does not say whether it is a bot or user token, and it does
  not cover creating an app from a manifest. Test this by hand before writing
  steps.
- **Sending.** Posting needs `chat:write` for user and bot tokens alike. Public
  channels accept a name or an ID. Private channels need the ID, and the person
  must be a member. Prefer IDs.
- **Permalinks.** The documentation disagrees about the scopes
  `chat.getPermalink` needs. Test with a real token before relying on the link
  as evidence.

Not verified: the exact screens for creating an app from a manifest, whether a
published app can be installed by other workspaces without review, and the
permalink scope.

## Shape

- Two ways to get a token, one storage and send path.
  - **Sign in (preferred).** The tool opens the browser, listens briefly on
    `127.0.0.1`, completes PKCE, and stores the refresh token. It needs a
    published Slack app owned by the project.
  - **Bring your own token.** The onboarder generates a manifest, and the
    newcomer installs their own app. It works where the organization will not
    approve a shared app.
- Before saving, the tool calls `auth.test` and shows "This token belongs to
  Pat Lee in the Acme workspace. Save it?"
- Storage uses a credential helper, as git does. The setting holds a command, as
  an argument list with no shell, that prints the token. Examples: a password
  manager's read command, the macOS Keychain's `security` tool, or `secret-tool`
  on Linux. A file path also works, which covers agenix and sops-nix. A 0600
  file in the state home is the plain default. Environment variables are
  avoided because they land in shell history and child processes. The helper
  setting lives in the state home and never in a plan.
- The token never goes in a plan, a log, or an error message. Every send is
  previewed and confirmed. The onboarder still lists the allowed modes, and
  `copy` always works.
- This is the first network call in the tool, and it goes only to Slack. The
  effects table in [0020](../decisions/0020-binary-structure-and-regenerative-components.md)
  needs a line for the Slack API and one for the loopback listener. Tests use a
  fake Slack server.
- A catalog section `message-sending-setup` in a "Messaging" group after
  `team-chat` would carry the manifest, the steps, and the token check. It
  should wait until a send mode exists.

## Open questions

- Can 8th Light own a Slack app and its developer account? If not, the
  bring-your-own route comes first.
- Teams and email need the same design. 0065 lists them as other `via` values.

## Sources

- [Using PKCE](https://docs.slack.dev/authentication/using-pkce/)
- [PKCE general availability](https://docs.slack.dev/changelog/2026/03/30/pkce/)
- [Creating an app from app settings](https://docs.slack.dev/app-management/quickstart-app-settings/)
- [chat.postMessage](https://docs.slack.dev/reference/methods/chat.postMessage/)
- [chat.getPermalink](https://docs.slack.dev/reference/methods/chat.getPermalink/)
- [App manifest reference](https://docs.slack.dev/reference/app-manifest/)
