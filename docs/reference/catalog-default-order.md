# Default order of the catalog sections

Status: proposed. A person corrects it, and then the catalog owns it. See
[0660](../decisions/0660-section-order-is-the-list-order.md).

The 59 section templates in the embedded catalog have no explicit order today. Each
section below links to its file in the catalog's content folder.
None declares `requires` or a parent, and 29 have the timing `anytime`, which
sorts after every dated window. Inside a window the order is the alphabet of the
ids. This page proposes a logical order, and a few changes to timing and to
`requires` that make that order true. A new plan starts in this order, and the
onboarder can move sections within a window.

## How the order was built

- **Windows first,** in the order of the timing set: `week-one`, `days-8-to-30`,
  `days-31-to-90`, then `anytime`. Nothing in the catalog uses `before-day-one`.
- **Then the journey.** Welcome and first meetings, day-one setup, team and ways of
  working, tool access, setup and learning the code, design work, and the first
  check-in. Inside each, a section follows what it depends on: an account before
  its membership, membership before repository access, access before a clone.
- **The onboarding repository comes first among repositories.** It is the plan's own
  repository, so it follows source host membership and comes before access to any code
  repository. A newcomer gets updates to their plan from it from then on.
- **`anytime` is for open-ended work,** such as help, feedback, and a first pull
  request when a suitable change comes up. It stays last.

## The order

| # | Section | Title | Window | Scope | Target | Change |
| -- | -- | -- | -- | -- | -- | -- |
| 1 | [`onboarding-welcome`](../../contracts/catalog/reader/content/sections/onboarding-welcome.yaml) | Welcome and expectations | `week-one` | `team` | First interaction |  |
| 2 | [`onboarding-owner-meeting`](../../contracts/catalog/reader/content/sections/onboarding-owner-meeting.yaml) | Meet the onboarding owner or manager | `week-one` | `team` | Day one |  |
| 3 | [`onboarding-buddy-meeting`](../../contracts/catalog/reader/content/sections/onboarding-buddy-meeting.yaml) | Meet the onboarding buddy | `week-one` | `team` | Day one |  |
| 4 | [`company-identity`](../../contracts/catalog/reader/content/sections/company-identity.yaml) | Company identity and SSO | `week-one` | `organization` | Day one |  |
| 5 | [`company-email`](../../contracts/catalog/reader/content/sections/company-email.yaml) | Company email | `week-one` | `organization` | Day one |  |
| 6 | [`company-calendar`](../../contracts/catalog/reader/content/sections/company-calendar.yaml) | Company calendar | `week-one` | `organization` | Day one |  |
| 7 | [`working-device`](../../contracts/catalog/reader/content/sections/working-device.yaml) | Working device | `week-one` | `organization` | Day one |  |
| 8 | [`working-hours`](../../contracts/catalog/reader/content/sections/working-hours.yaml) | Working hours and availability | `week-one` | `organization` | Day one |  |
| 9 | [`video-meetings`](../../contracts/catalog/reader/content/sections/video-meetings.yaml) | Video meetings | `week-one` | `organization` | Day one |  |
| 10 | [`team-chat`](../../contracts/catalog/reader/content/sections/team-chat.yaml) | Team chat | `week-one` | `team` | Day one |  |
| 11 | [`internal-network`](../../contracts/catalog/reader/content/sections/internal-network.yaml) | VPN or internal network | `week-one` | `organization` | Before dependent work | timing `anytime` to `week-one` |
| 12 | [`security-orientation`](../../contracts/catalog/reader/content/sections/security-orientation.yaml) | Security and acceptable-use orientation | `week-one` | `organization` | Before affected work | timing `anytime` to `week-one` |
| 13 | [`working-agreements`](../../contracts/catalog/reader/content/sections/working-agreements.yaml) | Review working agreements | `week-one` | `team` | First week |  |
| 14 | [`regular-meetings`](../../contracts/catalog/reader/content/sections/regular-meetings.yaml) | Join regular team meetings | `week-one` | `team` | First week |  |
| 15 | [`product-orientation`](../../contracts/catalog/reader/content/sections/product-orientation.yaml) | Understand the product and users | `week-one` | `team` | First week |  |
| 16 | [`teammate-introduction`](../../contracts/catalog/reader/content/sections/teammate-introduction.yaml) | Meet a teammate | `week-one` | `team` | First week |  |
| 17 | [`source-host-account`](../../contracts/catalog/reader/content/sections/source-host-account.yaml) | GitHub or other source-host account | `week-one` | `organization` | Before repository work | timing `anytime` to `week-one` |
| 18 | [`source-host-membership`](../../contracts/catalog/reader/content/sections/source-host-membership.yaml) | Source-host organization membership | `week-one` | `organization` | Before repository work | timing `anytime` to `week-one` |
| 19 | [`onboarding-repository`](../../contracts/catalog/reader/content/sections/onboarding-repository.yaml) | Connect to the onboarding repository | `week-one` | `organization` | After access | timing `anytime` to `week-one` |
| 20 | [`onboarding-refresh`](../../contracts/catalog/reader/content/sections/onboarding-refresh.yaml) | Get the latest plan | `week-one` | `organization` | After access | timing `anytime` to `week-one` |
| 21 | [`repository-access`](../../contracts/catalog/reader/content/sections/repository-access.yaml) | Repository access | `week-one` | `team` | Before work in that repository | timing `anytime` to `week-one` |
| 22 | [`ticket-system-access`](../../contracts/catalog/reader/content/sections/ticket-system-access.yaml) | Ticket system access | `week-one` | `team` | Before ticket work | timing `anytime` to `week-one` |
| 23 | [`documentation-access`](../../contracts/catalog/reader/content/sections/documentation-access.yaml) | Documentation workspace access | `week-one` | `organization` | Before assigned reading | timing `anytime` to `week-one` |
| 24 | [`ci-access`](../../contracts/catalog/reader/content/sections/ci-access.yaml) | CI access | `week-one` | `team` | Before delivery work | timing `anytime` to `week-one` |
| 25 | [`environment-access`](../../contracts/catalog/reader/content/sections/environment-access.yaml) | Development or staging environment access | `week-one` | `team` | Before affected work | timing `anytime` to `week-one` |
| 26 | [`artifact-access`](../../contracts/catalog/reader/content/sections/artifact-access.yaml) | Package or artifact registry access | `week-one` | `team` | Before dependent setup | timing `anytime` to `week-one` |
| 27 | [`repository-clone`](../../contracts/catalog/reader/content/sections/repository-clone.yaml) | Clone a repository | `week-one` | `team` | First week |  |
| 28 | [`project-dependencies`](../../contracts/catalog/reader/content/sections/project-dependencies.yaml) | Install project dependencies | `week-one` | `team` | First week |  |
| 29 | [`project-tests`](../../contracts/catalog/reader/content/sections/project-tests.yaml) | Run the test suite | `week-one` | `team` | First week |  |
| 30 | [`application-local`](../../contracts/catalog/reader/content/sections/application-local.yaml) | Run the application locally | `week-one` | `team` | First week |  |
| 31 | [`codebase-orientation`](../../contracts/catalog/reader/content/sections/codebase-orientation.yaml) | Understand the codebase | `week-one` | `team` | First week onward |  |
| 32 | [`contribution-practices`](../../contracts/catalog/reader/content/sections/contribution-practices.yaml) | Learn contribution and review practices | `week-one` | `team` | Before the first change | timing `anytime` to `week-one` |
| 33 | [`ticket-pairing`](../../contracts/catalog/reader/content/sections/ticket-pairing.yaml) | Pair on a ticket | `week-one` | `team` | First week |  |
| 34 | [`design-workspace`](../../contracts/catalog/reader/content/sections/design-workspace.yaml) | Design workspace access | `week-one` | `team` | First week |  |
| 35 | [`design-system-review`](../../contracts/catalog/reader/content/sections/design-system-review.yaml) | Review the design system | `week-one` | `team` | First week |  |
| 36 | [`design-handoff`](../../contracts/catalog/reader/content/sections/design-handoff.yaml) | Learn design review and handoff | `week-one` | `team` | Before the first contribution | timing `anytime` to `week-one` |
| 37 | [`week-one-check-in`](../../contracts/catalog/reader/content/sections/week-one-check-in.yaml) | End-of-week onboarding check-in | `week-one` | `team` | End of week one |  |
| 38 | [`people-map`](../../contracts/catalog/reader/content/sections/people-map.yaml) | Build a people map | `days-8-to-30` | `team` | First month |  |
| 39 | [`research-review`](../../contracts/catalog/reader/content/sections/research-review.yaml) | Review user research | `days-8-to-30` | `team` | First weeks |  |
| 40 | [`ticket-ownership`](../../contracts/catalog/reader/content/sections/ticket-ownership.yaml) | Own a ticket | `days-8-to-30` | `team` | First week or month |  |
| 41 | [`deploy-shadow`](../../contracts/catalog/reader/content/sections/deploy-shadow.yaml) | Shadow a deploy | `days-8-to-30` | `team` | First week or month |  |
| 42 | [`design-contribution`](../../contracts/catalog/reader/content/sections/design-contribution.yaml) | First design contribution | `days-8-to-30` | `team` | First month |  |
| 43 | [`day-thirty-review`](../../contracts/catalog/reader/content/sections/day-thirty-review.yaml) | Review 30-day goals | `days-8-to-30` | `team` | Around day 30 |  |
| 44 | [`day-ninety-review`](../../contracts/catalog/reader/content/sections/day-ninety-review.yaml) | Review 90-day goals | `days-31-to-90` | `team` | Around day 90 |  |
| 45 | [`office-access`](../../contracts/catalog/reader/content/sections/office-access.yaml) | Office access and security badge | `anytime` | `organization` | First on-site visit |  |
| 46 | [`document-reading`](../../contracts/catalog/reader/content/sections/document-reading.yaml) | Read a document | `anytime` | `team` | When its purpose requires |  |
| 47 | [`pr-post`](../../contracts/catalog/reader/content/sections/pr-post.yaml) | Post a PR | `anytime` | `team` | First suitable contribution |  |
| 48 | [`pr-review`](../../contracts/catalog/reader/content/sections/pr-review.yaml) | Review a PR | `anytime` | `team` | First suitable opportunity |  |
| 49 | [`pr-merge`](../../contracts/catalog/reader/content/sections/pr-merge.yaml) | Merge a PR | `anytime` | `team` | When authorized and ready |  |
| 50 | [`deploy-lead`](../../contracts/catalog/reader/content/sections/deploy-lead.yaml) | Lead a deploy with support | `anytime` | `team` | When authorized and ready |  |
| 51 | [`incident-orientation`](../../contracts/catalog/reader/content/sections/incident-orientation.yaml) | Incident and on-call orientation | `anytime` | `team` | Before participation |  |
| 52 | [`hosting-access`](../../contracts/catalog/reader/content/sections/hosting-access.yaml) | Cloud or hosting access | `anytime` | `team` | When the role needs it |  |
| 53 | [`observability-access`](../../contracts/catalog/reader/content/sections/observability-access.yaml) | Logs, metrics, and tracing access | `anytime` | `team` | Before diagnosis or deploy work |  |
| 54 | [`secrets-access`](../../contracts/catalog/reader/content/sections/secrets-access.yaml) | Secrets manager access | `anytime` | `team` | When needed |  |
| 55 | [`retrospective-participation`](../../contracts/catalog/reader/content/sections/retrospective-participation.yaml) | Participate in a retrospective | `anytime` | `team` | First available retrospective |  |
| 56 | [`tool-feedback`](../../contracts/catalog/reader/content/sections/tool-feedback.yaml) | Provide feedback on onboard-kit | `anytime` | `organization` | After initial use; revisit around day 30 |  |
| 57 | [`onboarding-correction`](../../contracts/catalog/reader/content/sections/onboarding-correction.yaml) | Report incorrect onboarding instructions | `anytime` | `organization` | Throughout |  |
| 58 | [`knowledge-transfer`](../../contracts/catalog/reader/content/sections/knowledge-transfer.yaml) | Knowledge transfer and onboarding ownership | `anytime` | `team` | Throughout; confirm before departure |  |
| 59 | [`onboarding-help`](../../contracts/catalog/reader/content/sections/onboarding-help.yaml) | Get help when blocked | `anytime` | `organization` | Throughout |  |

## Scope

Each section is marked for the organization plan or for a team plan, as
[0670](../decisions/0670-layered-plans-and-where-a-plan-lives.md) decides. The 17
organization sections are the HR and IT group of 0530 (company email, identity, calendar,
video meetings, device, working hours, internal network, security orientation, office
access), the accounts and shared documentation every team uses (source host account and
membership, documentation access), and how onboarding itself works (help, correction,
tool feedback, the plan repository, and refreshing the plan). The other 42 are the team's:
meetings and ways of working, repository and tool access that differs by team, learning
the code, contributing, deploying, design work, and the reviews. In a composed plan, the organization's sections come first within each window,
in their order here, followed by the team's in theirs, so the order a person sees is not
the interleaved order of the table. A team section may require an organization section.
A team plan cannot change an organization section. This is a proposal, and `team` is the default when a section
could be either.

## Timing changes

Fourteen sections move from `anytime` to `week-one`, because their goal says they
come before work that happens in the first week: `internal-network`,
`security-orientation`, `source-host-account`, `source-host-membership`,
`repository-access`, `ticket-system-access`, `documentation-access`, `ci-access`,
`environment-access`, `artifact-access`, `onboarding-repository`,
`onboarding-refresh`, `contribution-practices`, and `design-handoff`. Without the
change, every access section would sort after the setup that needs it.

## Edges decided in review

Three edges were dropped or changed after a review: `ci-access` no longer requires
`repository-access`, because CI may be a hosted service tied to the repository or a
separate system with its own accounts. `onboarding-repository` requires
`source-host-membership`, the account and membership on the forge, and not access to a
specific code repository. `pr-merge` no longer requires `pr-post`.

## Proposed `requires` edges

An edge gates: the dependent waits for a verified outcome, and switching off a
prerequisite switches off its dependents. Keep an edge only where that is true. Order
alone needs no edge, because list position says it. These make the order a
prerequisite order where one exists. The action sequence
already puts a prerequisite before its dependent, so they also fix the first-steps
order. They are proposals: each is a statement about onboarding, and a person
decides which to keep.

| Section | Requires |
| -- | -- |
| [`company-email`](../../contracts/catalog/reader/content/sections/company-email.yaml) | [`company-identity`](../../contracts/catalog/reader/content/sections/company-identity.yaml) |
| [`company-calendar`](../../contracts/catalog/reader/content/sections/company-calendar.yaml) | [`company-email`](../../contracts/catalog/reader/content/sections/company-email.yaml) |
| [`video-meetings`](../../contracts/catalog/reader/content/sections/video-meetings.yaml) | [`company-email`](../../contracts/catalog/reader/content/sections/company-email.yaml) |
| [`team-chat`](../../contracts/catalog/reader/content/sections/team-chat.yaml) | [`company-identity`](../../contracts/catalog/reader/content/sections/company-identity.yaml) |
| [`internal-network`](../../contracts/catalog/reader/content/sections/internal-network.yaml) | [`working-device`](../../contracts/catalog/reader/content/sections/working-device.yaml) |
| [`source-host-membership`](../../contracts/catalog/reader/content/sections/source-host-membership.yaml) | [`source-host-account`](../../contracts/catalog/reader/content/sections/source-host-account.yaml) |
| [`repository-access`](../../contracts/catalog/reader/content/sections/repository-access.yaml) | [`source-host-membership`](../../contracts/catalog/reader/content/sections/source-host-membership.yaml) |
| [`onboarding-repository`](../../contracts/catalog/reader/content/sections/onboarding-repository.yaml) | [`source-host-membership`](../../contracts/catalog/reader/content/sections/source-host-membership.yaml) |
| [`onboarding-refresh`](../../contracts/catalog/reader/content/sections/onboarding-refresh.yaml) | [`onboarding-repository`](../../contracts/catalog/reader/content/sections/onboarding-repository.yaml) |
| [`repository-clone`](../../contracts/catalog/reader/content/sections/repository-clone.yaml) | [`repository-access`](../../contracts/catalog/reader/content/sections/repository-access.yaml) |
| [`project-dependencies`](../../contracts/catalog/reader/content/sections/project-dependencies.yaml) | [`repository-clone`](../../contracts/catalog/reader/content/sections/repository-clone.yaml) |
| [`application-local`](../../contracts/catalog/reader/content/sections/application-local.yaml) | [`project-tests`](../../contracts/catalog/reader/content/sections/project-tests.yaml) |
| [`project-tests`](../../contracts/catalog/reader/content/sections/project-tests.yaml) | [`project-dependencies`](../../contracts/catalog/reader/content/sections/project-dependencies.yaml) |
| [`ticket-pairing`](../../contracts/catalog/reader/content/sections/ticket-pairing.yaml) | [`ticket-system-access`](../../contracts/catalog/reader/content/sections/ticket-system-access.yaml) |
| [`ticket-ownership`](../../contracts/catalog/reader/content/sections/ticket-ownership.yaml) | [`ticket-system-access`](../../contracts/catalog/reader/content/sections/ticket-system-access.yaml) |
| [`deploy-shadow`](../../contracts/catalog/reader/content/sections/deploy-shadow.yaml) | [`ci-access`](../../contracts/catalog/reader/content/sections/ci-access.yaml) |
| [`deploy-lead`](../../contracts/catalog/reader/content/sections/deploy-lead.yaml) | [`deploy-shadow`](../../contracts/catalog/reader/content/sections/deploy-shadow.yaml) |
| [`pr-post`](../../contracts/catalog/reader/content/sections/pr-post.yaml) | [`contribution-practices`](../../contracts/catalog/reader/content/sections/contribution-practices.yaml) |
| [`pr-review`](../../contracts/catalog/reader/content/sections/pr-review.yaml) | [`repository-access`](../../contracts/catalog/reader/content/sections/repository-access.yaml) |
| [`design-contribution`](../../contracts/catalog/reader/content/sections/design-contribution.yaml) | [`design-handoff`](../../contracts/catalog/reader/content/sections/design-handoff.yaml) |

## Proposed addition

**`first-change-merged`, "Get your first change merged."** The catalog has `pr-post` for
posting a change and `pr-merge` for completing an authorized merge, which may be another
person's change or one's own when the policy permits. It has no section for the milestone
of one's own change landing. This one would be `team` scope, window `days-8-to-30`, placed
before the thirty-day review, and it requires `pr-post`. That edge is true: a change cannot
be merged before it is posted. `pr-merge` then keeps its meaning of merging when authorized,
and its summary would say a teammate's change. The two ids differ by one letter, which is
easy to misread, so the new one carries the longer name. This adds a 60th section, which
changes the catalog's pinned count and its fixtures.

## What this needs

The order and the timing changes are catalog content. They change the compiled
catalog and the fixtures that pin its contents, so a person authorizes them. Plans
made before the change see a changed template baseline, which is why this is best
done before the first release.
