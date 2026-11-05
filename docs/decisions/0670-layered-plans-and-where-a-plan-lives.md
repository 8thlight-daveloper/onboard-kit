---
category: architecture
status: accepted
approval: person
affects: [catalog/materializer, command/invite, invitation/encoder, invitation/reviewer, invitation/starter, plan/decoder, plan/encoder, plan/refresher, plan/validator, record/reconciler, tui]
builds-on: ['0030', '0090', '0095', '0120', '0300', '0580', '0610', '0630', '0660']
---

# 0670. Layered plans, and where a plan lives

Builds on [0030](0030-self-contained-plan-file.md),
[0090](0090-plan-and-onboarding-record.md),
[0095](0095-reports.md),
[0120](0120-defaults-require-acceptance.md),
[0300](0300-explicit-invitation-export.md),
[0580](0580-security-scope-at-this-stage.md),
[0610](0610-people-work-in-screens-and-flags-serve-scripts.md),
[0630](0630-the-preparation-tab-saves-what-was-reviewed.md), and
[0660](0660-section-order-is-the-list-order.md). It settles how an organization's
plan and its teams' plans relate, how they reach a person's machine, and where a
person's own record is kept.

## Context

A plan built from the embedded catalog asks no organization question and no team
question. Its questions are section questions, and every answer belongs to one plan.
That fits one team. It does not fit an organization that wants something from
everyone. Take a new compliance system that every person must use. With one plan per
team, the item is copied into each plan and drifts. With one plan for the whole
organization, everyone edits one file.

The tool is offline and one file (0005, 0030). A plan reaches a newcomer as an
invitation. Two sections in the catalog, `onboarding-repository` and
`onboarding-refresh`, describe a shared plan location and an update route, and both
say the tool supplies no automatic command. Teams already keep plans in git.

## Decision

**Sections have a scope.** Each catalog section is organization-targeted, such as
company email or security orientation, or team-targeted, such as team chat or working
agreements. The catalog marks each one. The Answer screen groups questions by scope,
organization first.

**There are organization plans and team plans, in one format, and every organization
has both.** The first question of an organization plan is the organization's name. The
first question of a team plan is the team's name. Asking for a name first is the same in
both, and both come before the targeted questions. Both are required onboarder questions,
so an invitation cannot be exported until they are answered. A team plan says which plan
it extends, and takes the organization's name from it, so it never asks again. An
organization with one team has an organization plan and one team plan. There is no merged
shape, so nothing has to be split when a second team arrives. Plans that already exist are
not changed by this.

**A team plan changes nothing in the organization plan.** It adds its own sections and
orders them. It cannot switch off, reorder, or edit an organization section. Whether
something is mandatory is the existing `required` flag on its question, and a team that
cannot answer a required question defers it explicitly, as 0035 allows. There is no new
mark for required or offered sections. A team that needs different wording writes its own
section. Nobody has needed to change the organization's sections from a team, and this
waits for someone to.

**The first plan of an organization is two plans in one flow.** A new organization has no
organization plan yet, so the front door's "Build an onboarding plan" sets up the
organization first, then offers "Add your first team", and writes both files. Adding
another team later starts from the team question and asks where the organization plan is.

**`extends` is a relative path and a plan id.** Example: a path `../org/plan.obk.yaml`
and the id `acme`. The tool resolves the path from the team plan's folder. If it is
not there, the tool looks for a known plan with that id in the plan list. If neither
works, it asks the person to find the file, as the front door does. A team plan
records the hash of the organization plan it last composed against. Opening it
compares the two hashes with no network, and a difference offers a review through the
plan refresher, with nothing changed until the person accepts, as 0120 requires.

**Composition has a fixed order.** Within one timing window, the organization plan's
sections come first and then the team plan's, each in its own list order. 0660 then
applies to the composed list.

**Composition happens at export and again at refresh.** The invitation carries the
composed plan, so a newcomer needs no other file. It records the id and hash of each
plan it came from. A later refresh composes from the plan repository, with the same
code.

**The onboarder says where the plan is kept.** A question in the first sections of a
plan, "Where is this plan kept?", takes a repository address, the path of the plan
file inside it, and an optional branch. It becomes the structured form of
`sharing_location` in `onboarding-repository`. Export copies it into the invitation.

**The plan repository has a standard place on a person's machine.** The tool
reads and writes nothing there until the person asks. The place is
`<plans folder>/<host>/<path of the address>`, so one clone serves every plan in a
repository. The plans folder is the platform data folder: `$XDG_DATA_HOME/obk/plans`,
or `~/.local/share/obk/plans`, on Linux, and `~/Library/Application Support/obk/plans`
on macOS. A setting can change it, and a setting can say a repository was cloned
somewhere else.

**The tool checks that the plan is there, and never runs git.** It reads only the
plan file at the standard place and compares its plan id. If the file is there and
matches, the tool offers any update. If it is missing, the tool says the repository is
not cloned, shows the command, and offers "Remind me later". The plan works without it.
If something else is there, the tool says so and touches nothing.

**Until the action runner exists, the tool shows the command.** The line is
`git clone ADDRESS PLACE`, and a later pull is `git -C PLACE pull`. The person runs it.
When the runner exists, the same prompt becomes a confirmation to run it. The address
comes from an invitation, so the tool allows only forms such as `https://`, `ssh://`,
and `git@host:path`. It refuses an address that starts with `-` and any exotic
transport, quotes every argument, and shows the address on its own line.

**A person's record lives in the tool's own data folder.** The path is
`<data folder>/records/<plan id>/<account>/record.obk.yaml`. It works before any clone
exists, and a clone of a plan repository never has to avoid it. The plan repository
stays a mirror that the tool only reads, so a pull cannot conflict with anything the
tool wrote. The record keeps a single owner, as 0090 says. A person shares progress
with "Share my progress". It shows the exact file, says who can read the repository,
and asks for confirmation. Then it copies the record to `records/<account>/record.obk.yaml`
in the plan repository, the layout the first-steps guide already names, with an
atomic write. The person commits and pushes it, and the tool shows the command.
Nothing writes there silently. The tool shares the whole record, with no option to leave parts out. The file holds
the person's answers, outcomes, notes, and the "differently" feedback that says where
the plan was wrong, and that feedback is what keeps a plan current. The
onboarder's preparation sits beside it, in `records/<account>/preparation.obk.yaml`.
The commit history is the history of progress, and the commit author is the
evidence of who shared it. The onboarder sees progress through the report of 0095,
which reads `records/*` in the repository after a pull.

**Writes stay atomic.** Records use the store pattern of the other writers: a private
temporary file in the same folder, a check that the bytes have not changed since the
load, an atomic publish, and recovery of a leftover file. A stale save reloads and
asks again, as 0630 decides.

**About lists the plan repository.** For each one it shows the address, the standard
place, whether the plan file is there, the plan id and hash it holds against the
invitation's, and the update settings. It also lists the records folder.

**The plan repository shows when it was last fetched.** About, the front door's list
of plans, and the plan's own header show the time of the last fetch, such as "Updated
3 days ago". The tool takes it from the modification time of the repository's fetch
record, which is a metadata read and never reads inside `.git`. About lists that read.
The time shows as words and a number, never colour alone. Two settings,
`updates.stale_after_days` and `updates.critical_after_days`, decide when the line
changes. After a week, the default of the first, it reads "Updated 9 days ago, stale"
and takes the warning style. After a month, the default of the second, it reads
"Updated 41 days ago, overdue" and takes the strongest style, so the person knows to
update by hand. The words carry the state, and colour only adds to them.
Where no fetch
record exists, the tool shows when the plan file last changed instead. With the
runner and the settings on, fetches run on their own and the time stays fresh.

**Direction, not built here.**

- A settings file in the state home. `updates.check` is `off` by default or `fetch`,
  which reads the network and shows the notice. `updates.apply` is `ask`, where a
  fetch only reports and a pull needs its own confirmation, or `auto`, where the pull
  runs at once.
- An auto-pull never runs a new or changed action. A pull can change a plan, and a
  changed plan can add commands. The trust a person gave an action is bound to its
  content, and a changed action asks again.
- A step with a command shows it as a proposal that a person confirms. This is the
  action runner of the v2 epic, with its own record and a security review, and it
  needs the newcomer's own screens first.
- **Begin from a repository.** A person who has no invitation, such as an existing
  employee asked to use a new plan, can open the front door's "Begin from a
  repository", or run the tool with a repository address. The tool shows the clone
  command for the standard place, and once the plan is there it starts a record
  under a name the person gives. An update to a plan that a person already has comes
  with their next pull and launch, so it needs no new invitation.
- A built-in newcomer fact for the operating system, offered as a suggestion from the
  tool's own platform and accepted by the person, so a template can choose
  instructions for a machine.

## Consequences

- The plan format gains `extends` and a recorded hash. The decoder, encoder,
  validator, and refresher change. A composer is a new component with its own
  contract, and it is a core component that reads no file.
- The invitation envelope gains provenance and the plan source. The invitation
  encoder, reviewer, and starter change, and so does the invite command.
- Records move from a chosen root to the data folder. The starter and the session
  components change where they read and write.
- The catalog marks the scope of each section, and `onboarding-repository` and
  `onboarding-refresh` change. These are catalog content and fixtures that a person
  authorizes, and they are best changed before the first release.
- URL handling is new input from a file that came from outside, so the safety rules
  above need a test that tries `--upload-pack` and similar forms.
- Contract changes need a person's authorization.

## Not decided

1. **Whether a team may switch off an organization section.** The answer today is no. It
   stays no until a team has a real need, and then a mark on the organization section
   would be the place to say which sections may be dropped.
