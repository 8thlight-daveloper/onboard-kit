# Onboarding catalog

Status: all eighty entries are compiled into the executable and available offline.

[0150](../decisions/0150-onboarding-library-and-preparation-flow.md) requires the
full catalog in the binary. This reference specifies coverage and the read-only
availability commands. Guided selection and record-backed workflows remain planned.

[0320](../decisions/0320-embedded-catalog-availability.md) divides offline
availability into 59 section templates, 19 onboarder preparation guides and
two non-instantiable tool-help topics. The full authored content is a permanent
input to the [catalog reader contract](../../contracts/catalog/reader/contract.md).
Independent binary scenarios check the complete inventory and exact exported
bytes without supplying catalog files beside the executable.

Sources are the [worked example](examples/acme/README.md), the repository's core
and discipline guides, and the product discussion recorded in 0150. Each row
names a useful outcome. Standard tool wording is embedded once, not copied
into every organization's authored sections.

## Using the embedded catalog

```sh
onboard-kit catalog list
onboard-kit catalog show --id company-email
onboard-kit catalog export --id company-email > company-email.yaml
```

`list` prints a JSON inventory sorted by exact entry ID. `show` includes the
metadata and complete content. `export` prints only the original content:
section YAML for templates, Markdown for preparation guides and tool help.
The shell redirection above creates the file; the command itself writes only
to standard output and standard error. It does not load a plan or local templates.

Use the exact ID from `list`. Unknown IDs and malformed arguments fail without
echoing supplied values. Add `--help` immediately after `catalog` or a recognized
operation for usage. Each command uses the compiled catalog, not local files.
Existing custom-template loading remains available through `new` and `walk`.

Availability does not select activities, fill in organizational facts, generate
personal preparation, or expand repeated activities. The onboarder still supplies
the routes, contacts and policies requested by the section's questions.

## Reading the catalog

- The id identifies a library entry, not a generated repeated instance.
- Targets are planning goals, not deadlines that fail or lock activities.
- Before-start targets describe organizational preparation, not work owed by a newcomer.
- Access is verified by use of the capability, not by a request being sent.
- Credentials never belong in a template, preparation file, or onboarding record.
- Repeated activities use separate sections when outcomes are independent.
- Onboarders supply real routes, commands, contacts, and permission levels.

## Invitation and first steps

| Id | Entry | Target | Result |
| -- | -- | -- | -- |
| `invitation-start` | Start from the invitation ZIP | First interaction | Understand the package, get the tool, confirm identity, and start without company access |
| `onboarding-welcome` | Welcome and expectations | First interaction | Know the team, owner, start arrangements, and initial goals |
| `onboarding-help` | Get help when blocked | Throughout | Have a reachable initial contact and a fallback when the usual route fails |
| `onboarding-repository` | Connect to the onboarding repository | After access | Know where files go and who can read them |
| `onboarding-refresh` | Get the latest plan | After access | Request an update without losing personal work |
| `onboarding-resume` | Resume onboarding | Throughout | Return to the tool, record outcomes, and report differing instructions |

`invitation-start`, `onboarding-resume`, and parts of repository and refresh
guidance include standard tool wording. The [first-steps document](first-steps-document.md)
presents it alongside applicable bootstrap sections. Getting company access
must not depend on already having that access.

## Identity, equipment, and communication

| Id | Entry | Target | Result |
| -- | -- | -- | -- |
| `working-device` | Working device | Day one | An approved device is available and usable |
| `company-identity` | Company identity and SSO | Day one | Required authentication works |
| `company-email` | Company email | Day one | Send and receive mail |
| `company-calendar` | Company calendar | Day one | View the calendar and respond to invitations |
| `working-hours` | Working hours and availability | Day one | Agree and configure hours, time zone overlap, and availability signals |
| `team-chat` | Team chat | Day one | Reach relevant team and support channels |
| `video-meetings` | Video meetings | Day one | Join with required audio and collaboration features |
| `internal-network` | VPN or internal network | Before dependent work | Reach required internal resources |
| `office-access` | Office access and security badge | First on-site visit | Entry and badge work; arrival instructions and contact are clear |
| `security-orientation` | Security and acceptable-use orientation | Before affected work | Understand required practices and support routes |

Chat and video entries use the configured services, including Slack, Webex,
Zoom, or others. Do not assume a provider or automatically create accounts.
Hybrid onboarding can require both office and remote arrangements.

## Team and coordination

| Id | Entry | Target | Result |
| -- | -- | -- | -- |
| `onboarding-owner-meeting` | Meet the onboarding owner or manager | Day one | Understand support, priorities, and the first week's plan |
| `onboarding-buddy-meeting` | Meet the onboarding buddy | Day one | Know the day-to-day first contact |
| `teammate-introduction` | Meet a teammate | First week | Know their responsibilities and when to ask them |
| `regular-meetings` | Join regular team meetings | First week | Know which meetings matter and be able to attend |
| `working-agreements` | Review working agreements | First week | Understand communication, decisions, review, availability, and feedback |
| `product-orientation` | Understand the product and users | First week | Explain the users, product purpose, and current priorities |
| `people-map` | Build a people map | First month | Understand important relationships, responsibilities, and dependencies |

Introductions can repeat for configured people or roles. An established team
teaches its practices and history. A team with no working agreements establishes
them with the people involved rather than pretending they already exist.

## Tools, repositories, and documentation

| Id | Entry | Target | Result |
| -- | -- | -- | -- |
| `source-host-account` | GitHub or other source-host account | Before repository work | The appropriate named account and authentication work |
| `source-host-membership` | Source-host organization membership | Before repository work | Join the required organization or group |
| `repository-access` | Repository access | Before work in that repository | Use the required capability for that specific repository |
| `ticket-system-access` | Ticket system access | Before ticket work | Read and update the relevant project or board as needed |
| `ci-access` | CI access | Before delivery work | Access required pipeline results and actions |
| `documentation-access` | Documentation workspace access | Before assigned reading | Open the assigned material |
| `document-reading` | Read a document | When its purpose requires | Read with a stated purpose and a route for questions |
| `artifact-access` | Package or artifact registry access | Before dependent setup | Obtain required dependencies and artifacts |
| `environment-access` | Development or staging environment access | Before affected work | Use the required environment |
| `hosting-access` | Cloud or hosting access | When the role needs it | Use the required role and permissions |
| `observability-access` | Logs, metrics, and tracing access | Before diagnosis or deploy work | Inspect the relevant service |
| `secrets-access` | Secrets manager access | When needed | Obtain necessary secrets through the approved system |

Repository access and reading assignments repeat from the team's configured
lists. Other resources can repeat when outcomes differ. Access means what the
role needs, not unrestricted access to every repository or production system.

## Software work and delivery

| Id | Entry | Target | Result |
| -- | -- | -- | -- |
| `repository-clone` | Clone a repository | First week | Have a usable local checkout |
| `project-dependencies` | Install project dependencies | First week | Complete the documented setup |
| `project-tests` | Run the test suite | First week | Run tests and identify known failures rather than hide them |
| `application-local` | Run the application locally | First week | Exercise the application |
| `codebase-orientation` | Understand the codebase | First week onward | Locate main components and trace a relevant flow |
| `contribution-practices` | Learn contribution and review practices | Before the first change | Understand branching, review, checks, and the definition of done |
| `ticket-pairing` | Pair on a ticket | First week | Work through a real task with a teammate |
| `ticket-ownership` | Own a ticket | First week or month | Take a suitable task through the team's delivery process |
| `pr-review` | Review a PR | First suitable opportunity | Give a substantive review following team practices |
| `pr-post` | Post a PR | First suitable contribution | Submit a real change with context, checks, and review requests |
| `pr-merge` | Merge a PR | When authorized and ready | Complete an authorized merge after required reviews and checks |
| `deploy-shadow` | Shadow a deploy | First week or month | Observe preparation, deployment, verification, and rollback considerations |
| `deploy-lead` | Lead a deploy with support | When authorized and ready | Lead an appropriate deployment with experienced support available |
| `incident-orientation` | Incident and on-call orientation | Before participation | Understand responsibilities, escalation, and relevant runbooks |

For `pr-merge`, the onboarder must state whether team policy permits the person's
own PR, another person's PR, or both. The learning goal is the authorized merge
process, not authorship. The template asks for that policy rather than choosing
one. The tool neither grants permissions nor performs the merge. Never bypass a
separation-of-duties rule or grant permissions solely to complete onboarding.

The ACME clone-and-test example groups several instructions. Independent
readiness outcomes can use separate sections without changing that historical
example or its acceptance fixture.

## Design work

| Id | Entry | Target | Result |
| -- | -- | -- | -- |
| `design-workspace` | Design workspace access | First week | Access required files and tools |
| `design-system-review` | Review the design system | First week | Find and use the system and main product files |
| `research-review` | Review user research | First weeks | Understand relevant findings and open questions |
| `design-handoff` | Learn design review and handoff | Before the first contribution | Understand review, approval, and collaboration expectations |
| `design-contribution` | First design contribution | First month | Take a suitable contribution through the team's process |

## Feedback and progress

| Id | Entry | Target | Result |
| -- | -- | -- | -- |
| `week-one-check-in` | End-of-week onboarding check-in | End of week one | Discuss blockers, surprises, and next priorities |
| `retrospective-participation` | Participate in a retrospective | First available retrospective | Participate and understand how improvements are followed up |
| `onboarding-correction` | Report incorrect onboarding instructions | Throughout | Give the plan owner corrections without silently editing shared instructions |
| `tool-feedback` | Provide feedback on onboard-kit | After initial use; revisit around day 30 | Send tool friction and suggestions through the agreed route without secrets or unnecessary personal details |
| `day-thirty-review` | Review 30-day goals | Around day 30 | Agree on progress, feedback, and revised priorities |
| `day-ninety-review` | Review 90-day goals | Around day 90 | Agree on outcomes and the next phase |
| `knowledge-transfer` | Knowledge transfer and onboarding ownership | Throughout; confirm before departure | Name knowledge recipients and an owner for shared onboarding material |

## Onboarder preparation

| Id | Entry | Target | Result |
| -- | -- | -- | -- |
| `invitation-prepare` | Prepare the invitation | Before delivery | Review recipient details, relevant activities, missing preparation, and the package preview |
| `support-assign` | Assign onboarding support | Before introductions | Identify manager, onboarding owner, buddy, and initial help contacts as applicable |
| `device-arrange` | Arrange a working device | Before device setup | Arrange delivery and setup support |
| `identity-arrange` | Arrange company identity | Before identity and email verification | Prepare provisioning routes, grantors, and requests |
| `calendar-arrange` | Arrange calendar access | Before calendar verification | Arrange the account and calendar access |
| `chat-arrange` | Arrange chat membership | Before team communication | Arrange required service accounts and channel memberships |
| `video-arrange` | Arrange meeting service access | Before meeting participation | Arrange configured accounts or guest access |
| `office-arrange` | Arrange office access and badge | Before the first on-site visit | Supply the badge route, arrival instructions, and reachable host |
| `remote-arrange` | Arrange remote onboarding | Before remote or hybrid start | Arrange equipment, connection instructions, and a remote first-day contact |
| `meeting-invite` | Send meeting invitations | Before the relevant meeting | Supply invitations and joining instructions for introductions, regular meetings, retrospectives, and check-ins |
| `resource-access-arrange` | Arrange resource access | Before dependent work | Identify the grantor, permissions, and request route; track requests for selected resources |
| `reading-assign` | Assign reading | Before the reading activity | Supply material, purpose, target, and a question contact |
| `pair-assign` | Assign a pairing partner | Before ticket pairing | Identify a partner and suitable ticket |
| `ticket-assign` | Identify a first owned ticket | Before ticket ownership | Identify a suitable task and support contact |
| `pr-review-arrange` | Arrange a PR review opportunity | Before PR review | Identify a suitable change and discussion contact |
| `pr-reviewer-assign` | Assign a reviewer | Before posting a PR | Identify a reviewer and review route |
| `pr-merge-arrange` | Arrange a merge opportunity | Before merging a PR | Confirm a suitable opportunity, authorization, and any support |
| `deploy-lead-assign` | Assign an experienced deploy lead | Before the newcomer shadows | Identify the experienced lead and observation opportunity |
| `deploy-support-assign` | Assign deployment support | Before the newcomer leads | Identify an experienced observer or support contact and suitable deployment |

These are onboarder activities, not newcomer outcomes or proof that access works.
Person-specific arrangements belong in the preparation file from
[0140](../decisions/0140-person-specific-preparation.md); shared routes and policies
stay in the plan.

Selecting an activity surfaces its preparation. Omit irrelevant preparation,
but retain shared arrangements needed by another selected activity. Targets
are early enough to enable the work, not necessarily all before day one.
Unknown partners and opportunities stay visibly unassigned.

## Delivery boundaries

Catalog completeness, authored template content, embedding, repeated instances,
and the three-file preparation workflow are distinct checks. This document
alone does not deliver the library. The [preparation format](preparation-format.md)
and command now provide person-specific ownership. All catalog content is compiled
and available through the read-only commands. Selection relationships, template
bindings, and list-driven expansion remain pending and need separate contracts.
Personal arrangements must not be stored
in shared answers.
