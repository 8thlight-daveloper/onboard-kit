# Default order of the catalog sections

Status: proposed. A person corrects it, and then the catalog owns it. See
[0660](../decisions/0660-section-order-is-the-list-order.md).

The 59 section templates in the embedded catalog have no explicit order today.
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
- **`anytime` is for open-ended work,** such as help, feedback, and a first pull
  request when a suitable change comes up. It stays last.

## The order

| # | Section | Window | Category | Change |
| -- | -- | -- | -- | -- |
| 1 | `onboarding-welcome` | `week-one` | invitation-and-first-steps |  |
| 2 | `onboarding-owner-meeting` | `week-one` | team-and-coordination |  |
| 3 | `onboarding-buddy-meeting` | `week-one` | team-and-coordination |  |
| 4 | `company-identity` | `week-one` | identity-equipment-and-communication |  |
| 5 | `company-email` | `week-one` | identity-equipment-and-communication |  |
| 6 | `company-calendar` | `week-one` | identity-equipment-and-communication |  |
| 7 | `working-device` | `week-one` | identity-equipment-and-communication |  |
| 8 | `working-hours` | `week-one` | identity-equipment-and-communication |  |
| 9 | `video-meetings` | `week-one` | identity-equipment-and-communication |  |
| 10 | `team-chat` | `week-one` | identity-equipment-and-communication |  |
| 11 | `internal-network` | `week-one` | identity-equipment-and-communication | timing `anytime` to `week-one` |
| 12 | `security-orientation` | `week-one` | identity-equipment-and-communication | timing `anytime` to `week-one` |
| 13 | `working-agreements` | `week-one` | team-and-coordination |  |
| 14 | `regular-meetings` | `week-one` | team-and-coordination |  |
| 15 | `product-orientation` | `week-one` | team-and-coordination |  |
| 16 | `teammate-introduction` | `week-one` | team-and-coordination |  |
| 17 | `source-host-account` | `week-one` | tools-repositories-and-documentation | timing `anytime` to `week-one` |
| 18 | `source-host-membership` | `week-one` | tools-repositories-and-documentation | timing `anytime` to `week-one` |
| 19 | `repository-access` | `week-one` | tools-repositories-and-documentation | timing `anytime` to `week-one` |
| 20 | `ticket-system-access` | `week-one` | tools-repositories-and-documentation | timing `anytime` to `week-one` |
| 21 | `documentation-access` | `week-one` | tools-repositories-and-documentation | timing `anytime` to `week-one` |
| 22 | `ci-access` | `week-one` | tools-repositories-and-documentation | timing `anytime` to `week-one` |
| 23 | `environment-access` | `week-one` | tools-repositories-and-documentation | timing `anytime` to `week-one` |
| 24 | `artifact-access` | `week-one` | tools-repositories-and-documentation | timing `anytime` to `week-one` |
| 25 | `onboarding-repository` | `week-one` | invitation-and-first-steps | timing `anytime` to `week-one` |
| 26 | `onboarding-refresh` | `week-one` | invitation-and-first-steps | timing `anytime` to `week-one` |
| 27 | `repository-clone` | `week-one` | software-work-and-delivery |  |
| 28 | `project-dependencies` | `week-one` | software-work-and-delivery |  |
| 29 | `project-tests` | `week-one` | software-work-and-delivery |  |
| 30 | `application-local` | `week-one` | software-work-and-delivery |  |
| 31 | `codebase-orientation` | `week-one` | software-work-and-delivery |  |
| 32 | `contribution-practices` | `week-one` | software-work-and-delivery | timing `anytime` to `week-one` |
| 33 | `ticket-pairing` | `week-one` | software-work-and-delivery |  |
| 34 | `design-workspace` | `week-one` | design-work |  |
| 35 | `design-system-review` | `week-one` | design-work |  |
| 36 | `design-handoff` | `week-one` | design-work | timing `anytime` to `week-one` |
| 37 | `week-one-check-in` | `week-one` | feedback-and-progress |  |
| 38 | `people-map` | `days-8-to-30` | team-and-coordination |  |
| 39 | `research-review` | `days-8-to-30` | design-work |  |
| 40 | `ticket-ownership` | `days-8-to-30` | software-work-and-delivery |  |
| 41 | `deploy-shadow` | `days-8-to-30` | software-work-and-delivery |  |
| 42 | `design-contribution` | `days-8-to-30` | design-work |  |
| 43 | `day-thirty-review` | `days-8-to-30` | feedback-and-progress |  |
| 44 | `day-ninety-review` | `days-31-to-90` | feedback-and-progress |  |
| 45 | `office-access` | `anytime` | identity-equipment-and-communication |  |
| 46 | `document-reading` | `anytime` | tools-repositories-and-documentation |  |
| 47 | `pr-post` | `anytime` | software-work-and-delivery |  |
| 48 | `pr-review` | `anytime` | software-work-and-delivery |  |
| 49 | `pr-merge` | `anytime` | software-work-and-delivery |  |
| 50 | `deploy-lead` | `anytime` | software-work-and-delivery |  |
| 51 | `incident-orientation` | `anytime` | software-work-and-delivery |  |
| 52 | `hosting-access` | `anytime` | tools-repositories-and-documentation |  |
| 53 | `observability-access` | `anytime` | tools-repositories-and-documentation |  |
| 54 | `secrets-access` | `anytime` | tools-repositories-and-documentation |  |
| 55 | `retrospective-participation` | `anytime` | feedback-and-progress |  |
| 56 | `tool-feedback` | `anytime` | feedback-and-progress |  |
| 57 | `onboarding-correction` | `anytime` | feedback-and-progress |  |
| 58 | `knowledge-transfer` | `anytime` | feedback-and-progress |  |
| 59 | `onboarding-help` | `anytime` | invitation-and-first-steps |  |

## Timing changes

Fourteen sections move from `anytime` to `week-one`, because their goal says they
come before work that happens in the first week: `internal-network`,
`security-orientation`, `source-host-account`, `source-host-membership`,
`repository-access`, `ticket-system-access`, `documentation-access`, `ci-access`,
`environment-access`, `artifact-access`, `onboarding-repository`,
`onboarding-refresh`, `contribution-practices`, and `design-handoff`. Without the
change, every access section would sort after the setup that needs it.

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
| `company-email` | `company-identity` |
| `company-calendar` | `company-email` |
| `video-meetings` | `company-email` |
| `team-chat` | `company-identity` |
| `internal-network` | `working-device` |
| `source-host-membership` | `source-host-account` |
| `repository-access` | `source-host-membership` |
| `ci-access` | `repository-access` |
| `onboarding-repository` | `repository-access` |
| `repository-clone` | `repository-access` |
| `project-dependencies` | `repository-clone` |
| `application-local` | `project-tests` |
| `project-tests` | `project-dependencies` |
| `ticket-pairing` | `ticket-system-access` |
| `ticket-ownership` | `ticket-system-access` |
| `deploy-shadow` | `ci-access` |
| `deploy-lead` | `deploy-shadow` |
| `pr-post` | `contribution-practices` |
| `pr-review` | `repository-access` |
| `pr-merge` | `pr-post` |
| `design-contribution` | `design-handoff` |

## What this needs

The order and the timing changes are catalog content. They change the compiled
catalog and the fixtures that pin its contents, so a person authorizes them. Plans
made before the change see a changed template baseline, which is why this is best
done before the first release.
