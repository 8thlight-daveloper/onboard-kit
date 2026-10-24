# Catalog preparation relations

Status: normative authored relation data under
[0410](../decisions/0410-explicit-catalog-instances.md); materialization pending.

The [catalog](onboarding-catalog.md) has section templates, preparation guides
and tool help. These relations surface advisory preparation for explicitly
selected section instances. They create no arrangements, activities, answers,
permissions, prerequisites or outcomes and change no reader manifest fields.

For any nonempty selection, include invitation-prepare once with every selected
permanent instance ID as a reason. Empty selection returns no activity-caused
guidance. The ordinary invitation flow may show its own guidance separately.

For each selected source, also include exactly the additional guides below.
An empty list means no additional guide, not that all shared facts are known.
Return guides in exact guide-ID byte order and each guide's distinct causing
instance IDs in exact byte order. Repeated sources retain each independent
instance reason. Removing one selection must not remove a guide still needed
by another selected instance.

| Section source ID | Additional preparation guide IDs |
| -- | -- |
| application-local | [] |
| artifact-access | resource-access-arrange |
| ci-access | resource-access-arrange |
| codebase-orientation | [] |
| company-calendar | calendar-arrange, meeting-invite |
| company-email | identity-arrange |
| company-identity | identity-arrange |
| contribution-practices | [] |
| day-ninety-review | meeting-invite, support-assign |
| day-thirty-review | meeting-invite, support-assign |
| deploy-lead | deploy-support-assign |
| deploy-shadow | deploy-lead-assign |
| design-contribution | support-assign |
| design-handoff | [] |
| design-system-review | [] |
| design-workspace | resource-access-arrange |
| document-reading | reading-assign |
| documentation-access | resource-access-arrange |
| environment-access | resource-access-arrange |
| hosting-access | resource-access-arrange |
| incident-orientation | support-assign |
| internal-network | resource-access-arrange |
| knowledge-transfer | [] |
| observability-access | resource-access-arrange |
| office-access | office-arrange |
| onboarding-buddy-meeting | meeting-invite, support-assign |
| onboarding-correction | [] |
| onboarding-help | support-assign |
| onboarding-owner-meeting | meeting-invite, support-assign |
| onboarding-refresh | [] |
| onboarding-repository | resource-access-arrange |
| onboarding-welcome | office-arrange, remote-arrange, support-assign |
| people-map | meeting-invite |
| pr-merge | pr-merge-arrange |
| pr-post | pr-reviewer-assign |
| pr-review | pr-review-arrange |
| product-orientation | [] |
| project-dependencies | [] |
| project-tests | [] |
| regular-meetings | meeting-invite |
| repository-access | resource-access-arrange |
| repository-clone | resource-access-arrange |
| research-review | [] |
| retrospective-participation | meeting-invite |
| secrets-access | resource-access-arrange |
| security-orientation | [] |
| source-host-account | resource-access-arrange |
| source-host-membership | resource-access-arrange |
| team-chat | chat-arrange |
| teammate-introduction | meeting-invite |
| ticket-ownership | ticket-assign |
| ticket-pairing | pair-assign |
| ticket-system-access | resource-access-arrange |
| tool-feedback | [] |
| video-meetings | video-arrange |
| week-one-check-in | meeting-invite, support-assign |
| working-agreements | [] |
| working-device | device-arrange |
| working-hours | [] |

The welcome row surfaces office and remote routes as conditional advice. It does
not claim both apply or infer workLocation. Personal mode and assignments require
explicit context in the person flow. Reading guidance applies to actual reading
assignments, not every activity that mentions documentation.

The deploy direction is deliberate: a newcomer shadowing needs an experienced
lead; a newcomer leading needs experienced support. A selected repository-clone
does not automatically select repository-access or add a prerequisite edge.
Shared guidance does not merge separate access or pairing opportunities.

The common guide plus these rows cover all nineteen preparation guides.
They remain advisory Markdown, not new arrangement kinds or completion facts.
Independent fixtures must hand-list expected unions and reasons rather than
derive their expected output from this same table at test time.
