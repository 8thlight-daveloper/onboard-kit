# Manual stage testing

Run these recipes from the repository's Nix dev shell. They build a separate
development binary at output/demo-bin/onboard-kit. They do not replace the
verified release file or edit the checked-in Acme example.

## Start from zero

```sh
just demo-init
just demo-plan
```

Init creates a new plan from Acme's four section templates and plan variables,
with a new identity, no accepted answers and no directory entries. It does not
copy the completed Acme plan. Repeating init or opening a stage preserves saved
work. Use the Contacts tab to add fictional contacts, roles and teams.
From a question's text field, press Esc to reach Overview, then q to quit
without accepting that draft answer.

You choose the account when you prepare: `just demo-prepare ACCOUNT`, where ACCOUNT is one folder name, such as jules. Later stages find it from the saved preparation. A second account needs `just demo-reset preparation` first. Use fictional data only. These files are ignored
by Git, but they are not encrypted and can appear in terminal scrollback.

## Run the stages

| Recipe | Interaction |
| -- | -- |
| `just demo-plan` | Edit the shared plan; walk saves accepted answers immediately |
| `just demo-prepare ACCOUNT` | Edit recipient details and support; review and save explicitly |
| `just demo-invite CONTACT` | Export with the exact pre-access helper contact ID supplied in place of CONTACT |
| `just demo-onboardee` | Review the invitation, confirm newcomer, then begin or cancel |
| `just demo-progress` | Confirm the registered identity, propose one personal answer or outcome, then save or cancel |
| `just demo-results` | Show the current preparation and returned record through read-only JSON |
| `just demo-status` | List exact binary, input and output paths without changing data |

Complete each stage's real prerequisites. A missing input or unanswered required
onboarder question is not filled automatically. The designated invitation helper
must be reachable before company access exists; the recipe cannot verify that.
The contact must be declared in the plan and preparation must identify an owner.

In invitation start, use confirm newcomer, then begin. Record editing uses the
different explicit form confirm "newcomer". It shows activity and question IDs;
use those IDs for changes. Each progress invocation makes at most one change.
See [record editing](session-edit.md), [preparation](preparation-command.md) and
[invitation export](invitation-export.md) for the actual interaction rules.

## Know which files are open

Every stage prints its paths before opening the application. The full-screen
walk does not yet show its plan path. Run just demo-status in another terminal
while that screen is open.

```text
output/manual-test/
  plan/plan.obk.yaml
  preparation/records/ACCOUNT/preparation.obk.yaml
  invitation/invitation.zip
  onboardee/workspace/records/ACCOUNT/record.obk.yaml
  onboardee/home/sessions/<key>/plan.obk.yaml
  onboardee/home/sessions/<key>/session.json
  results/records/ACCOUNT/preparation.obk.yaml
  results/records/ACCOUNT/record.obk.yaml
```

The onboardee uses the imported plan registered under their private home, not
the original organization plan. The onboarder owns the original preparation.
The results recipe copies current preparation and the returned personal
record into a derived view, then reads them with the current organization plan.
It never replaces the original personal record. This simulates returning a
record for review; it is not automatic Git synchronization or a production
transfer protocol. Results are currently JSON, not a graphical dashboard.

This helper uses the standard records/ACCOUNT layout. A custom resultsPath
needs separate explicit wiring; it is not guessed or silently rewritten.

## Reset only what you choose

```sh
just demo-reset invitation
just demo-reset onboardee
just demo-reset all
```

Valid stage names are plan, preparation, invitation, onboardee and results.
Reset moves only that stage's directory into a unique recovery folder under
output/manual-test-history/. It deletes no test payloads. The selected stage
returns to its not-started state; entering it again creates its initial files
when its prerequisites permit. Plan entry creates a new unanswered plan identity.

- Invitation reset retains plan, preparation and personal progress.
- Onboardee reset clears the active imported workspace and registration together;
  the invitation and onboarder inputs remain unchanged.
- Results reset clears only derived copies, never the source personal record.
- All is an explicit convenience, never a prerequisite of another recipe.

Changes to upstream inputs can make an existing invitation or record stale.
Keep those artifacts and diagnose mismatches; no reset cascades into another
stage. Close the selected stage before resetting it. Recipe-owned locks refuse
a reset while that stage is active through these recipes; all acquires every
stage lock before moving any data. An application launched another way does not
hold these helper locks, so close it yourself first.

The reset prints the recovery location. To recover, close the stage, archive
any newer stage with its reset, then move the archived stage directory back to
its original location. Never combine different onboardee workspace and home
generations. Do not remove security attributes or bypass an ownership refusal.

## Filesystem access controls

The application inspects existing operating-system access-control lists. It
does not create an application authorization system, add ACL entries, or grant
company access. Directory roles and contacts are onboarding data, not OS grants.

On macOS, native drivers inspect the extended security ACL using fgetattrlist.
Linux drivers inspect POSIX access and default ACLs. Drivers differ in their
supported profiles: private writers can require no ACL, recovery can retain
exact security metadata, and record replacement preserves supported metadata
or refuses. An absent ACL is not an explicitly empty ACL. Unix owner/group/mode
checks, extended attributes and native file flags are additional controls, not
synonyms for an ACL. These checks are not field encryption.

In particular, catalog editing currently refuses ACL-bearing ancestors. A
standard macOS home directory may have an existing everyone-deny-delete ACL.
The app did not install that rule. See [0490](../decisions/0490-release-runtime-workspace.md).
The demo uses the existing walk route for Acme plan building; it does not
silently relocate files or change ancestor permissions to enable catalog editing.
