package reader

import contract "onboard-kit/contracts/catalog/reader"

func inventory() []definition {
	return []definition{
		{
			summary: contract.Summary{
				ID:          "application-local",
				Title:       "Run the application locally",
				Category:    "software-work-and-delivery",
				Goal:        "First week",
				Result:      "Exercise the application",
				Carrier:     "section-template",
				ContentPath: "sections/application-local.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: application-local\n" +
				"title: Run the application locally\n" +
				"summary: Exercise the application\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: local_run_instructions, description: \"What reviewed local startup, shutdown and safe data instructions apply?\", audience: onboarder, required: true}\n" +
				"  - {name: local_exercise, description: \"Which harmless user journey should verify local behavior and who can help?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the local application instructions: $local_run_instructions\n" +
				"  Exercise the approved behavior: $local_exercise\n" +
				"\n" +
				"  Check environment and data sources before starting anything. Use approved\n" +
				"  sample data, not copied production data or personal credentials. Follow a\n" +
				"  representative flow and inspect its result, then shut down as documented.\n" +
				"  A running process alone is not a usable application. Report missing setup\n" +
				"  or differing instructions rather than changing shared services to fit them.\n" +
				"steps:\n" +
				"  - id: exercise-local-flow\n" +
				"    name: Exercise a safe local application flow\n" +
				"    checks:\n" +
				"      - ask: Can you complete the approved local flow and explain how to start and stop the application safely?\n",
		},
		{
			summary: contract.Summary{
				ID:          "artifact-access",
				Title:       "Package or artifact registry access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before dependent setup",
				Result:      "Obtain required dependencies and artifacts",
				Carrier:     "section-template",
				ContentPath: "sections/artifact-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: artifact-access\n" +
				"title: Package or artifact registry access\n" +
				"summary: Obtain required dependencies and artifacts\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: artifact_scope, description: \"Which approved registry and exact dependencies or artifacts are required?\", audience: onboarder, required: true}\n" +
				"  - {name: artifact_access_route, description: \"What approved authentication, request and safe retrieval instructions apply?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Confirm the registry and artifact scope: $artifact_scope\n" +
				"  Follow the approved retrieval instructions: $artifact_access_route\n" +
				"\n" +
				"  Retrieve an intended artifact with the permitted account and verify its\n" +
				"  identity or integrity using the project's documented method. Retrieval\n" +
				"  does not authorize publishing or deleting packages. Keep authentication\n" +
				"  material in the approved system, not a template or record. Do not switch\n" +
				"  to an untrusted mirror or disable verification to bypass missing access.\n" +
				"steps:\n" +
				"  - id: retrieve-approved-artifact\n" +
				"    name: Obtain the intended dependency or artifact\n" +
				"    checks:\n" +
				"      - ask: Can you retrieve and identify the required artifact through the approved registry and authentication route?\n",
		},
		{
			summary: contract.Summary{
				ID:          "calendar-arrange",
				Title:       "Arrange calendar access",
				Category:    "onboarder-preparation",
				Goal:        "Before calendar verification",
				Result:      "Arrange the account and calendar access",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/calendar-arrange.md",
			},
			content: "# Arrange calendar access\n" +
				"\n" +
				"Identify the configured calendar service, account grantor and approved request\n" +
				"process. Shared setup and sharing rules stay in the plan. Put the individual's\n" +
				"request in an access arrangement for an actual declared calendar section\n" +
				"instance, rather than inventing a calendar kind or silently selecting an activity.\n" +
				"\n" +
				"Keep a stable unique arrangement id and name known directory references.\n" +
				"Arranged requires an accountable owner and nonsecret reference. Pending,\n" +
				"requested and deferred can show missing details; supplied unknown contacts or\n" +
				"sections remain invalid. Use a real optional targetDate to coordinate support.\n" +
				"Do not record private calendar contents or personal reasons for availability.\n" +
				"\n" +
				"Coordinate a harmless invitation or support contact so the newcomer can test\n" +
				"viewing and responding after access is available. Sending an invitation and\n" +
				"granting an account are not verification of calendar use. If setup depends on\n" +
				"another service, make the support need clear without inventing a runtime\n" +
				"prerequisite edge. Preparation records arrangements, never newcomer outcomes.\n",
		},
		{
			summary: contract.Summary{
				ID:          "chat-arrange",
				Title:       "Arrange chat membership",
				Category:    "onboarder-preparation",
				Goal:        "Before team communication",
				Result:      "Arrange required service accounts and channel memberships",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/chat-arrange.md",
			},
			content: "# Arrange chat membership\n" +
				"\n" +
				"Confirm the organization's actual chat service, intended account and relevant\n" +
				"workspaces or channels. Keep shared joining instructions and fallback routes\n" +
				"in the plan. Do not assume a provider or create an account automatically.\n" +
				"Identify an initial route that does not depend on already having chat access.\n" +
				"\n" +
				"Use access arrangements for person-specific requests against actual declared\n" +
				"chat section instances. Distinct resources may need separate arrangements with\n" +
				"stable unique ids. Arranged requires a known owner and nonsecret reference;\n" +
				"helpers are included only when assigned. Pending, requested and deferred may\n" +
				"retain visible gaps, but unknown supplied contacts or sections are invalid.\n" +
				"Do not store invitations containing secret tokens or recovery material.\n" +
				"\n" +
				"Confirm who can resolve expired invitations or missing channel membership.\n" +
				"Targets coordinate support and do not fail the newcomer. A sent invitation or\n" +
				"arranged request does not establish that messages can be exchanged; the\n" +
				"newcomer verifies the actual capability. No chat-specific kind or outcome field\n" +
				"is introduced, and neither plan answers nor personal record are changed.\n",
		},
		{
			summary: contract.Summary{
				ID:          "ci-access",
				Title:       "CI access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before delivery work",
				Result:      "Access required pipeline results and actions",
				Carrier:     "section-template",
				ContentPath: "sections/ci-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: ci-access\n" +
				"title: CI access\n" +
				"summary: Access required pipeline results and actions\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: pipeline_scope, description: \"Which pipelines, result views and authorized actions does the role need?\", audience: onboarder, required: true}\n" +
				"  - {name: pipeline_access_route, description: \"Who grants CI access and how should required capabilities be verified safely?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the required pipeline scope: $pipeline_scope\n" +
				"  Follow the access and safe verification route: $pipeline_access_route\n" +
				"\n" +
				"  Open a relevant run and find its checks, logs and outcome. Learn which\n" +
				"  actions can spend resources, publish artifacts or affect environments.\n" +
				"  Viewing a run does not authorize rerunning or changing it. Never expose\n" +
				"  secret log material in onboarding notes. Confirm any action permission\n" +
				"  through the agreed safe procedure rather than triggering a live pipeline.\n" +
				"steps:\n" +
				"  - id: inspect-required-pipeline\n" +
				"    name: Find required CI results and understand action boundaries\n" +
				"    checks:\n" +
				"      - ask: Can you inspect the relevant pipeline results and identify which actions you are authorized to take?\n",
		},
		{
			summary: contract.Summary{
				ID:          "codebase-orientation",
				Title:       "Understand the codebase",
				Category:    "software-work-and-delivery",
				Goal:        "First week onward",
				Result:      "Locate main components and trace a relevant flow",
				Carrier:     "section-template",
				ContentPath: "sections/codebase-orientation.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: codebase-orientation\n" +
				"title: Understand the codebase\n" +
				"summary: Locate main components and trace a relevant flow\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: codebase_map, description: \"Which architecture material and main code locations should be explored?\", audience: onboarder, required: true}\n" +
				"  - {name: tracing_exercise, description: \"Which relevant flow should be traced and who can discuss its design?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Start with the current architecture and locations: $codebase_map\n" +
				"  Trace and discuss this useful flow: $tracing_exercise\n" +
				"\n" +
				"  Follow an input through the important components to its observable\n" +
				"  result. Locate related tests, ownership and failure handling. Compare\n" +
				"  the diagram with the actual code and ask about discrepancies. Aim for\n" +
				"  enough understanding to make a safe first change, not memorization of\n" +
				"  every module. Share corrections through the team's review process.\n" +
				"steps:\n" +
				"  - id: explain-code-flow\n" +
				"    name: Explain a relevant flow through the codebase\n" +
				"    checks:\n" +
				"      - ask: Can you locate the main components and explain a relevant flow, its tests and the owners to consult?\n",
		},
		{
			summary: contract.Summary{
				ID:          "company-calendar",
				Title:       "Company calendar",
				Category:    "identity-equipment-and-communication",
				Goal:        "Day one",
				Result:      "View the calendar and respond to invitations",
				Carrier:     "section-template",
				ContentPath: "sections/company-calendar.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: company-calendar\n" +
				"title: Company calendar\n" +
				"summary: View the calendar and respond to invitations\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: calendar_process, description: \"Which calendar service, setup instructions and sharing rules apply?\", audience: onboarder, required: true}\n" +
				"  - {name: calendar_test, description: \"Which harmless invitation or support contact can be used to verify calendar access?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the calendar setup and sharing instructions: $calendar_process\n" +
				"  Verify with the supplied invitation or contact: $calendar_test\n" +
				"\n" +
				"  Check the displayed time zone before responding. Open an intended event,\n" +
				"  inspect its joining details and send an appropriate response. Confirm\n" +
				"  that your availability is shared only as policy requires. A meeting\n" +
				"  invitation being sent does not prove you can view or answer it.\n" +
				"steps:\n" +
				"  - id: answer-invitation\n" +
				"    name: Open and respond to an intended calendar invitation\n" +
				"    checks:\n" +
				"      - ask: Can you view the intended calendar and respond to an invitation with the correct time zone?\n",
		},
		{
			summary: contract.Summary{
				ID:          "company-email",
				Title:       "Company email",
				Category:    "identity-equipment-and-communication",
				Goal:        "Day one",
				Result:      "Send and receive mail",
				Carrier:     "section-template",
				ContentPath: "sections/company-email.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: company-email\n" +
				"title: Company email\n" +
				"summary: Send and receive mail\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: email_process, description: \"Which approved mail service and setup instructions apply?\", audience: onboarder, required: true}\n" +
				"  - {name: email_test_route, description: \"Which business contact or safe test procedure should verify sending and receiving mail?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Set up the approved mail service using $email_process.\n" +
				"  Use the safe test procedure or contact: $email_test_route\n" +
				"\n" +
				"  Confirm the mailbox is your intended company mailbox. Send a nonsecret\n" +
				"  test message and receive a reply or equivalent approved test. Check how\n" +
				"  to reach support if mail is unavailable, rather than depending on that\n" +
				"  mailbox to obtain its own access. Never forward business mail to a\n" +
				"  personal account as an onboarding workaround.\n" +
				"steps:\n" +
				"  - id: exchange-mail\n" +
				"    name: Verify sending and receiving\n" +
				"    checks:\n" +
				"      - ask: Have you sent and received a nonsecret test message using the intended company mailbox?\n",
		},
		{
			summary: contract.Summary{
				ID:          "company-identity",
				Title:       "Company identity and SSO",
				Category:    "identity-equipment-and-communication",
				Goal:        "Day one",
				Result:      "Required authentication works",
				Carrier:     "section-template",
				ContentPath: "sections/company-identity.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: company-identity\n" +
				"title: Company identity and SSO\n" +
				"summary: Required authentication works\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: identity_process, description: \"What approved account activation and authentication instructions apply?\", audience: onboarder, required: true}\n" +
				"  - {name: identity_support, description: \"Who handles identity provisioning, recovery and authentication problems through an approved route?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Use the approved identity process: $identity_process\n" +
				"  Get provisioning or recovery help through $identity_support.\n" +
				"\n" +
				"  Confirm that the account belongs to you before activation. Set up the\n" +
				"  required authentication factors through the approved system, keeping\n" +
				"  passwords and recovery material out of onboarding files. Test sign-in to\n" +
				"  a permitted service and sign out afterwards where appropriate. A request\n" +
				"  or activation email alone does not prove working authentication.\n" +
				"steps:\n" +
				"  - id: verify-authentication\n" +
				"    name: Verify the required authentication flow\n" +
				"    checks:\n" +
				"      - ask: Can you authenticate with your own intended account and required factors to an approved service?\n",
		},
		{
			summary: contract.Summary{
				ID:          "contribution-practices",
				Title:       "Learn contribution and review practices",
				Category:    "software-work-and-delivery",
				Goal:        "Before the first change",
				Result:      "Understand branching, review, checks, and the definition of done",
				Carrier:     "section-template",
				ContentPath: "sections/contribution-practices.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: contribution-practices\n" +
				"title: Learn contribution and review practices\n" +
				"summary: Understand branching, review, checks, and the definition of done\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: contribution_policy, description: \"Where are the current branching, change, review and completion rules?\", audience: onboarder, required: true}\n" +
				"  - {name: contribution_discussion, description: \"Who can explain authorization, required checks and exceptions for a first contribution?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the contribution rules: $contribution_policy\n" +
				"  Resolve questions through $contribution_discussion.\n" +
				"\n" +
				"  Walk through an example change from task to delivery. Identify required\n" +
				"  reviews, checks, release conditions and who may approve or merge. Ask how\n" +
				"  to handle disagreement and an urgent exception; do not invent one.\n" +
				"  Access to a control is not authorization to use it. Keep the team's\n" +
				"  definition of done visible before beginning the first contribution.\n" +
				"steps:\n" +
				"  - id: explain-contribution-process\n" +
				"    name: Explain how a contribution becomes done\n" +
				"    checks:\n" +
				"      - ask: Can you explain the required branching, review, checks and authorization before a contribution is considered done?\n",
		},
		{
			summary: contract.Summary{
				ID:          "day-ninety-review",
				Title:       "Review 90-day goals",
				Category:    "feedback-and-progress",
				Goal:        "Around day 90",
				Result:      "Agree on outcomes and the next phase",
				Carrier:     "section-template",
				ContentPath: "sections/day-ninety-review.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: day-ninety-review\n" +
				"title: Review 90-day goals\n" +
				"summary: Agree on outcomes and the next phase\n" +
				"timing: days-31-to-90\n" +
				"variables:\n" +
				"  - {name: ninety_day_review, description: \"How is the roughly ninety-day review arranged and which outcomes should be discussed?\", audience: onboarder, required: true}\n" +
				"  - {name: next_phase_process, description: \"How are next-phase priorities, ownership and support agreed?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Arrange the review using $ninety_day_review.\n" +
				"  Agree the next phase through $next_phase_process.\n" +
				"\n" +
				"  Discuss outcomes, confidence, unresolved dependencies and what should\n" +
				"  change in the support model. Keep organizational barriers visible and\n" +
				"  make room for feedback on onboarding itself. Agree next responsibilities\n" +
				"  rather than inferring them from time served. The planning window neither\n" +
				"  fails someone nor grants access, authority or automatic completion.\n" +
				"steps:\n" +
				"  - id: agree-next-phase\n" +
				"    name: Review outcomes and agree the next phase\n" +
				"    checks:\n" +
				"      - ask: Have you reviewed the relevant outcomes and agreed next-phase priorities, responsibilities and support?\n",
		},
		{
			summary: contract.Summary{
				ID:          "day-thirty-review",
				Title:       "Review 30-day goals",
				Category:    "feedback-and-progress",
				Goal:        "Around day 30",
				Result:      "Agree on progress, feedback, and revised priorities",
				Carrier:     "section-template",
				ContentPath: "sections/day-thirty-review.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: day-thirty-review\n" +
				"title: Review 30-day goals\n" +
				"summary: Agree on progress, feedback, and revised priorities\n" +
				"timing: days-8-to-30\n" +
				"variables:\n" +
				"  - {name: thirty_day_review, description: \"How is the roughly thirty-day review arranged and which shared goals guide it?\", audience: onboarder, required: true}\n" +
				"  - {name: review_support_route, description: \"Who can revise priorities and arrange support after the review?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Arrange the review and revisit the shared goals: $thirty_day_review\n" +
				"  Agree changes and support through $review_support_route.\n" +
				"\n" +
				"  Discuss useful outcomes, remaining blockers, feedback and changing work.\n" +
				"  Distinguish verified progress from requests, attempts and unavailable\n" +
				"  opportunities. Revise priorities with the people involved and name any\n" +
				"  support actions. Around day thirty is a planning goal, not a pass/fail\n" +
				"  deadline or a demand for private performance notes in an onboarding file.\n" +
				"steps:\n" +
				"  - id: agree-next-month-priorities\n" +
				"    name: Review progress and revise the next priorities\n" +
				"    checks:\n" +
				"      - ask: Have you discussed progress and feedback and agreed revised priorities and support with the relevant owner?\n",
		},
		{
			summary: contract.Summary{
				ID:          "deploy-lead",
				Title:       "Lead a deploy with support",
				Category:    "software-work-and-delivery",
				Goal:        "When authorized and ready",
				Result:      "Lead an appropriate deployment with experienced support available",
				Carrier:     "section-template",
				ContentPath: "sections/deploy-lead.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: deploy-lead\n" +
				"title: Lead a deploy with support\n" +
				"summary: Lead an appropriate deployment with experienced support available\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: lead_deploy_policy, description: \"What readiness, authorization and supported deployment procedure must a first lead satisfy?\", audience: onboarder, required: true}\n" +
				"  - {name: deploy_stop_rules, description: \"Who may stop or roll back deployment and how is experienced support reached?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the first-lead policy and reviewed procedure: $lead_deploy_policy\n" +
				"  Confirm stop decisions, rollback and support: $deploy_stop_rules\n" +
				"\n" +
				"  Confirm a suitable deployment and available experienced supporter in\n" +
				"  preparation. Check current authorization and readiness before acting.\n" +
				"  Lead only the agreed scope, verify the result and follow up with the\n" +
				"  supporter. Stop or postpone when conditions are unsafe. A target date or\n" +
				"  onboarding goal never authorizes deployment or excuses missing support.\n" +
				"steps:\n" +
				"  - id: lead-supported-deployment\n" +
				"    name: Lead an authorized deployment with experienced support\n" +
				"    checks:\n" +
				"      - ask: Have you led the agreed authorized deployment with support available and verified its outcome under the runbook?\n",
		},
		{
			summary: contract.Summary{
				ID:          "deploy-lead-assign",
				Title:       "Assign an experienced deploy lead",
				Category:    "onboarder-preparation",
				Goal:        "Before the newcomer shadows",
				Result:      "Identify the experienced lead and observation opportunity",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/deploy-lead-assign.md",
			},
			content: "# Assign an experienced deploy lead\n" +
				"\n" +
				"Find an experienced person leading an appropriate deployment for the newcomer\n" +
				"to observe. Keep the deployment runbook and shared safety rules in the plan.\n" +
				"Use deploy-shadow for this personal arrangement and name an actual declared\n" +
				"shadowing section instance; the guide's title does not introduce a new kind.\n" +
				"\n" +
				"An arranged item requires a known accountable owner, at least one known helper\n" +
				"and a nonsecret deployment opportunity reference. Here helpers are experienced\n" +
				"leads whom the newcomer observes. They are not automatically supporters for a\n" +
				"newcomer-led deployment. Keep a stable unique id and optional real planning\n" +
				"target. Missing details may be pending, requested or deferred, but unknown\n" +
				"supplied helpers or section references are invalid.\n" +
				"\n" +
				"Confirm observation access and a route for questions without granting control\n" +
				"authority. Discuss verification, rollback and cancellation. A scheduled or\n" +
				"arranged deployment is not observed work or a verified newcomer outcome.\n" +
				"If conditions change, rearrange the opportunity; the date does not compel it.\n",
		},
		{
			summary: contract.Summary{
				ID:          "deploy-shadow",
				Title:       "Shadow a deploy",
				Category:    "software-work-and-delivery",
				Goal:        "First week or month",
				Result:      "Observe preparation, deployment, verification, and rollback considerations",
				Carrier:     "section-template",
				ContentPath: "sections/deploy-shadow.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: deploy-shadow\n" +
				"title: Shadow a deploy\n" +
				"summary: Observe preparation, deployment, verification, and rollback considerations\n" +
				"timing: days-8-to-30\n" +
				"variables:\n" +
				"  - {name: deployment_runbook, description: \"Which deployment process and runbook should be observed?\", audience: onboarder, required: true}\n" +
				"  - {name: shadow_expectations, description: \"How is a safe observation opportunity arranged and what questions should the observer cover?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the process before observing: $deployment_runbook\n" +
				"  Arrange the opportunity and learning focus: $shadow_expectations\n" +
				"\n" +
				"  Confirm the experienced lead and deployment opportunity in preparation.\n" +
				"  Observe readiness checks, authorization, execution, verification and\n" +
				"  rollback considerations. Ask when a deploy should stop and who decides.\n" +
				"  Observation does not authorize operating production controls. If the\n" +
				"  planned deployment is cancelled or unsuitable, rearrange it instead of\n" +
				"  treating the planning date as a failure.\n" +
				"steps:\n" +
				"  - id: observe-deployment-lifecycle\n" +
				"    name: Observe an experienced lead through a deployment\n" +
				"    checks:\n" +
				"      - ask: Have you observed the deployment lifecycle and discussed readiness, verification, rollback and decision ownership?\n",
		},
		{
			summary: contract.Summary{
				ID:          "deploy-support-assign",
				Title:       "Assign deployment support",
				Category:    "onboarder-preparation",
				Goal:        "Before the newcomer leads",
				Result:      "Identify an experienced observer or support contact and suitable deployment",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/deploy-support-assign.md",
			},
			content: "# Assign deployment support\n" +
				"\n" +
				"Arrange a suitable authorized deployment and experienced support for the\n" +
				"newcomer to lead it. Keep the runbook, readiness and stop rules in the plan.\n" +
				"Use deploy-lead for the person-specific arrangement and an actual declared\n" +
				"lead-deployment section instance. Never infer readiness from a planning date.\n" +
				"\n" +
				"An arranged item needs a known accountable owner, at least one known helper\n" +
				"and a nonsecret opportunity reference. For this kind helpers are experienced\n" +
				"observers or supporters, not the lead being shadowed. Give the item a stable\n" +
				"unique id and optional real targetDate. Pending, requested or deferred keep\n" +
				"missing support visible; unknown supplied contacts and sections are invalid.\n" +
				"\n" +
				"Confirm support availability, authority to act, safe scope, verification and\n" +
				"rollback decisions. Do not grant production rights solely for onboarding.\n" +
				"Arranged is a minimum preparation claim, not permission at action time,\n" +
				"successful deployment or a verified newcomer outcome. Preserve the committed\n" +
				"ownership boundaries: no shared-answer injection or personal-record writes.\n",
		},
		{
			summary: contract.Summary{
				ID:          "design-contribution",
				Title:       "First design contribution",
				Category:    "design-work",
				Goal:        "First month",
				Result:      "Take a suitable contribution through the team's process",
				Carrier:     "section-template",
				ContentPath: "sections/design-contribution.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: design-contribution\n" +
				"title: First design contribution\n" +
				"summary: Take a suitable contribution through the team's process\n" +
				"timing: days-8-to-30\n" +
				"variables:\n" +
				"  - {name: design_contribution_scope, description: \"What makes a first design contribution suitable and how should its scope be agreed?\", audience: onboarder, required: true}\n" +
				"  - {name: design_contribution_process, description: \"Which support, review and completion process must that contribution follow?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Agree a suitable real contribution using $design_contribution_scope.\n" +
				"  Follow the support and completion process: $design_contribution_process\n" +
				"\n" +
				"  Clarify the user need, constraints and expected evidence with the owner.\n" +
				"  Explore and communicate the proposed change, seek the required review\n" +
				"  and follow through on the agreed handoff. Keep research and personal\n" +
				"  data within approved systems. A first-month goal does not justify\n" +
				"  bypassing research, accessibility, approval or collaboration expectations.\n" +
				"steps:\n" +
				"  - id: complete-design-contribution\n" +
				"    name: Complete a suitably scoped design contribution\n" +
				"    checks:\n" +
				"      - ask: Has your agreed contribution met the team's review, evidence and handoff conditions with support available?\n",
		},
		{
			summary: contract.Summary{
				ID:          "design-handoff",
				Title:       "Learn design review and handoff",
				Category:    "design-work",
				Goal:        "Before the first contribution",
				Result:      "Understand review, approval, and collaboration expectations",
				Carrier:     "section-template",
				ContentPath: "sections/design-handoff.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: design-handoff\n" +
				"title: Learn design review and handoff\n" +
				"summary: Understand review, approval, and collaboration expectations\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: design_handoff_process, description: \"What design review, approval and engineering handoff process applies?\", audience: onboarder, required: true}\n" +
				"  - {name: design_handoff_example, description: \"Which reviewed example and discussion contact can explain expected handoff evidence?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the collaboration process: $design_handoff_process\n" +
				"  Walk through the supplied example with $design_handoff_example.\n" +
				"\n" +
				"  Identify how intent, states, accessibility, constraints and unresolved\n" +
				"  questions are communicated. Learn who approves what and how changes\n" +
				"  remain visible after handoff. A shared link is not proof that another\n" +
				"  discipline has the context it needs. Ask about ambiguous responsibilities\n" +
				"  rather than assuming that design sign-off grants delivery authority.\n" +
				"steps:\n" +
				"  - id: explain-design-handoff\n" +
				"    name: Explain a useful reviewed design handoff\n" +
				"    checks:\n" +
				"      - ask: Can you explain the review, approval and collaboration evidence needed for a useful handoff?\n",
		},
		{
			summary: contract.Summary{
				ID:          "design-system-review",
				Title:       "Review the design system",
				Category:    "design-work",
				Goal:        "First week",
				Result:      "Find and use the system and main product files",
				Carrier:     "section-template",
				ContentPath: "sections/design-system-review.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: design-system-review\n" +
				"title: Review the design system\n" +
				"summary: Find and use the system and main product files\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: design_system_material, description: \"Where are the current design system, main product files and usage guidance?\", audience: onboarder, required: true}\n" +
				"  - {name: design_system_discussion, description: \"Which representative task should be explored and who explains gaps or contribution rules?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Explore the current system and product files: $design_system_material\n" +
				"  Discuss a representative use through $design_system_discussion.\n" +
				"\n" +
				"  Locate the relevant components, patterns, accessibility guidance and\n" +
				"  source of truth. Explain how a common interface uses them and how an\n" +
				"  exception is reviewed. Distinguish a published component from a draft.\n" +
				"  If no system exists, ask the team to identify current conventions and\n" +
				"  gaps; do not invent a claim that an established system was reviewed.\n" +
				"steps:\n" +
				"  - id: explain-system-use\n" +
				"    name: Locate and explain a useful design-system pattern\n" +
				"    checks:\n" +
				"      - ask: Can you find the current design references and explain how to use or discuss a relevant pattern and its gaps?\n",
		},
		{
			summary: contract.Summary{
				ID:          "design-workspace",
				Title:       "Design workspace access",
				Category:    "design-work",
				Goal:        "First week",
				Result:      "Access required files and tools",
				Carrier:     "section-template",
				ContentPath: "sections/design-workspace.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: design-workspace\n" +
				"title: Design workspace access\n" +
				"summary: Access required files and tools\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: design_workspace_scope, description: \"Which configured design tools, workspaces and file capabilities does the role require?\", audience: onboarder, required: true}\n" +
				"  - {name: design_access_route, description: \"How is design access requested and safely verified, and who handles problems?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Confirm the required tools, files and capability: $design_workspace_scope\n" +
				"  Follow the access and verification route: $design_access_route\n" +
				"\n" +
				"  Open the intended workspace with your approved identity. Verify the\n" +
				"  permitted capability on an agreed safe file; viewing is not editing\n" +
				"  authority. Learn where current work lives and what may be shared outside\n" +
				"  the team. Do not duplicate restricted assets into a personal workspace\n" +
				"  or request organization-wide rights to resolve a narrow access problem.\n" +
				"steps:\n" +
				"  - id: verify-design-workspace\n" +
				"    name: Verify the intended design files and capability\n" +
				"    checks:\n" +
				"      - ask: Can you use the required design files and tools within the approved permission and sharing scope?\n",
		},
		{
			summary: contract.Summary{
				ID:          "device-arrange",
				Title:       "Arrange a working device",
				Category:    "onboarder-preparation",
				Goal:        "Before device setup",
				Result:      "Arrange delivery and setup support",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/device-arrange.md",
			},
			content: "# Arrange a working device\n" +
				"\n" +
				"Confirm the approved equipment process, delivery expectations and setup support\n" +
				"with the responsible team. Keep shared setup instructions in the plan. Put a\n" +
				"person's equipment opportunity in an equipment arrangement linked to an actual\n" +
				"declared section instance; do not create a catalog-id binding automatically.\n" +
				"Do not put a home address, private shipping details or credentials in preparation.\n" +
				"\n" +
				"Give the arrangement a stable unique id. Name a known accountable owner and a\n" +
				"nonsecret reference to the approved equipment request or handoff before marking\n" +
				"it arranged. Add known support contacts when available. Pending, requested or\n" +
				"deferred may retain missing details visibly, but supplied unknown references\n" +
				"are invalid. A targetDate is an optional real calendar planning date.\n" +
				"\n" +
				"Confirm who resolves delayed delivery or failed setup and how the newcomer\n" +
				"can reach them before company access. Arranged means the onboarder has made\n" +
				"the minimum arrangement, not that a device arrived or works. The newcomer\n" +
				"verifies usability in the corresponding activity; preparation changes no record.\n",
		},
		{
			summary: contract.Summary{
				ID:          "document-reading",
				Title:       "Read a document",
				Category:    "tools-repositories-and-documentation",
				Goal:        "When its purpose requires",
				Result:      "Read with a stated purpose and a route for questions",
				Carrier:     "section-template",
				ContentPath: "sections/document-reading.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: document-reading\n" +
				"title: Read a document\n" +
				"summary: Read with a stated purpose and a route for questions\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: reading_material, description: \"Which current document should be read and for what work purpose?\", audience: onboarder, required: true}\n" +
				"  - {name: reading_questions, description: \"Which questions should guide the reading and who can discuss them?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Read the material for the supplied purpose: $reading_material\n" +
				"  Use these questions and discussion route: $reading_questions\n" +
				"\n" +
				"  Identify the decisions or actions the material informs, not just whether\n" +
				"  you opened it. Note uncertainties and confirm stale or contradictory\n" +
				"  instructions with its owner. Do not paste restricted content into an\n" +
				"  onboarding record. Use separate activities for reading assignments whose\n" +
				"  outcomes need independent discussion or follow-up.\n" +
				"steps:\n" +
				"  - id: discuss-reading\n" +
				"    name: Explain the reading's useful implications\n" +
				"    checks:\n" +
				"      - ask: Can you explain how the document informs your work and raise any unresolved questions through the supplied route?\n",
		},
		{
			summary: contract.Summary{
				ID:          "documentation-access",
				Title:       "Documentation workspace access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before assigned reading",
				Result:      "Open the assigned material",
				Carrier:     "section-template",
				ContentPath: "sections/documentation-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: documentation-access\n" +
				"title: Documentation workspace access\n" +
				"summary: Open the assigned material\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: documentation_scope, description: \"Which documentation workspace and material should the newcomer be able to open?\", audience: onboarder, required: true}\n" +
				"  - {name: documentation_route, description: \"Who grants documentation access and resolves missing or restricted material?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Locate the intended workspace and material: $documentation_scope\n" +
				"  Request or repair access through $documentation_route.\n" +
				"\n" +
				"  Open the assigned content using the intended account, not a copied\n" +
				"  private document. Check how to find the current version and its owner.\n" +
				"  Workspace membership alone does not prove a particular page opens.\n" +
				"  Keep access restrictions intact and ask for an approved alternative\n" +
				"  if assigned material cannot be shared with your role.\n" +
				"steps:\n" +
				"  - id: open-assigned-material\n" +
				"    name: Open the intended documentation\n" +
				"    checks:\n" +
				"      - ask: Can you open the assigned current material and identify its owner or access support route?\n",
		},
		{
			summary: contract.Summary{
				ID:          "environment-access",
				Title:       "Development or staging environment access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before affected work",
				Result:      "Use the required environment",
				Carrier:     "section-template",
				ContentPath: "sections/environment-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: environment-access\n" +
				"title: Development or staging environment access\n" +
				"summary: Use the required environment\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: environment_scope, description: \"Which development or staging environment and permitted tasks are required?\", audience: onboarder, required: true}\n" +
				"  - {name: environment_route, description: \"How is environment access granted and what safe verification and support procedure applies?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Identify the intended environment and task scope: $environment_scope\n" +
				"  Obtain and verify access using $environment_route.\n" +
				"\n" +
				"  Check environment identity before interacting with it. Perform only the\n" +
				"  agreed safe task; do not assume staging data is public or disposable.\n" +
				"  Production permission is separate and is not granted by this activity.\n" +
				"  Report missing access without modifying shared resources or copying\n" +
				"  sensitive data into a local workaround.\n" +
				"steps:\n" +
				"  - id: verify-environment\n" +
				"    name: Use the approved environment safely\n" +
				"    checks:\n" +
				"      - ask: Can you perform the agreed task in the intended environment and distinguish it from production?\n",
		},
		{
			summary: contract.Summary{
				ID:          "hosting-access",
				Title:       "Cloud or hosting access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "When the role needs it",
				Result:      "Use the required role and permissions",
				Carrier:     "section-template",
				ContentPath: "sections/hosting-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: hosting-access\n" +
				"title: Cloud or hosting access\n" +
				"summary: Use the required role and permissions\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: hosting_scope, description: \"Which hosting account, role and resource scope does this work require?\", audience: onboarder, required: true}\n" +
				"  - {name: hosting_route, description: \"What approved role-assumption, access request and safe verification process applies?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Confirm the required hosting scope and role: $hosting_scope\n" +
				"  Follow the access process: $hosting_route\n" +
				"\n" +
				"  Verify account and role context before using a permitted read or other\n" +
				"  approved safe operation. Do not request administrator permissions,\n" +
				"  change billing or create resources merely to prove sign-in. A console\n" +
				"  opening is not proof of authority for every action. Escalate excessive\n" +
				"  or missing permissions to the grantor without retaining secret material.\n" +
				"steps:\n" +
				"  - id: verify-hosting-role\n" +
				"    name: Verify the intended hosting role and permission scope\n" +
				"    checks:\n" +
				"      - ask: Can you use the intended role for the approved task and identify its resource and action limits?\n",
		},
		{
			summary: contract.Summary{
				ID:          "identity-arrange",
				Title:       "Arrange company identity",
				Category:    "onboarder-preparation",
				Goal:        "Before identity and email verification",
				Result:      "Prepare provisioning routes, grantors, and requests",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/identity-arrange.md",
			},
			content: "# Arrange company identity\n" +
				"\n" +
				"Confirm the provisioning owner, approved account policy and request route.\n" +
				"Keep those shared facts in the plan. Preparation may carry a known company\n" +
				"email when useful, but its absence is not itself an invitation blocker.\n" +
				"Never store an initial password, recovery code or personal delivery address.\n" +
				"\n" +
				"Track a person-specific provisioning request with the existing access kind\n" +
				"and an actual declared identity section instance. Use a stable unique id,\n" +
				"known owner and nonsecret request reference before claiming arranged. An access\n" +
				"item need not invent a helper when none is assigned. Missing details can stay\n" +
				"visible while pending, requested or deferred; an unknown supplied contact is\n" +
				"invalid rather than incomplete. Do not invent an identity-arrange kind.\n" +
				"\n" +
				"Ensure activation and support instructions are usable before the account works.\n" +
				"Confirm any target date as a planning goal, not a deadline that fails someone.\n" +
				"Requested or arranged provisioning is not working authentication, confirmed\n" +
				"recipient identity, invitation import or first use. Verification belongs to\n" +
				"the newcomer's actual capability check, not this preparation status.\n",
		},
		{
			summary: contract.Summary{
				ID:          "incident-orientation",
				Title:       "Incident and on-call orientation",
				Category:    "software-work-and-delivery",
				Goal:        "Before participation",
				Result:      "Understand responsibilities, escalation, and relevant runbooks",
				Carrier:     "section-template",
				ContentPath: "sections/incident-orientation.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: incident-orientation\n" +
				"title: Incident and on-call orientation\n" +
				"summary: Understand responsibilities, escalation, and relevant runbooks\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: incident_material, description: \"Which incident roles, runbooks and any on-call expectations apply to this work?\", audience: onboarder, required: true}\n" +
				"  - {name: incident_escalation, description: \"What approved incident reporting, escalation and training route should a newcomer use?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the relevant responsibilities and runbooks: $incident_material\n" +
				"  Locate the reporting and escalation route: $incident_escalation\n" +
				"\n" +
				"  Discuss a safe scenario with an experienced teammate. Identify how to\n" +
				"  raise concern, who coordinates response and where communications belong.\n" +
				"  Distinguish learning from joining a rota or accepting incident authority.\n" +
				"  Do not trigger a real alert, test production recovery or promise on-call\n" +
				"  availability without explicit agreement and the required training.\n" +
				"steps:\n" +
				"  - id: explain-incident-response\n" +
				"    name: Walk through an appropriate incident scenario\n" +
				"    checks:\n" +
				"      - ask: Can you explain your expected role, escalation route and the runbooks to consult before incident participation?\n",
		},
		{
			summary: contract.Summary{
				ID:          "internal-network",
				Title:       "VPN or internal network",
				Category:    "identity-equipment-and-communication",
				Goal:        "Before dependent work",
				Result:      "Reach required internal resources",
				Carrier:     "section-template",
				ContentPath: "sections/internal-network.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: internal-network\n" +
				"title: VPN or internal network\n" +
				"summary: Reach required internal resources\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: network_setup, description: \"What approved VPN or internal-network instructions and resource scope apply?\", audience: onboarder, required: true}\n" +
				"  - {name: network_test, description: \"Which permitted internal resource is a safe connectivity test, and where is connection support?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Set up only the required network access: $network_setup\n" +
				"  Test the permitted resource and use support as described: $network_test\n" +
				"\n" +
				"  Use approved clients and authentication. Distinguish a connection being\n" +
				"  established from authorization to a particular resource. Do not disable\n" +
				"  protective controls, scan other systems or request unrestricted network\n" +
				"  access to complete this activity. Explain the blocked dependent work if\n" +
				"  the approved route does not work.\n" +
				"steps:\n" +
				"  - id: verify-network\n" +
				"    name: Reach an approved internal resource\n" +
				"    checks:\n" +
				"      - ask: Can you reach the specified permitted internal resource through the approved network route?\n",
		},
		{
			summary: contract.Summary{
				ID:          "invitation-prepare",
				Title:       "Prepare the invitation",
				Category:    "onboarder-preparation",
				Goal:        "Before delivery",
				Result:      "Review recipient details, relevant activities, missing preparation, and the package preview",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/invitation-prepare.md",
			},
			content: "# Prepare the invitation\n" +
				"\n" +
				"The onboarder reviews one person's preparation beside the untouched shared\n" +
				"plan. Confirm the explicit account and matching plan identity; do not infer\n" +
				"identity from a directory name or copy another person's preparation. Supply\n" +
				"known business details only. Company email, a planned start date and optional\n" +
				"arrangements may remain absent; missing optional details alone do not block\n" +
				"export. Do not collect personal delivery email, phone or home address.\n" +
				"\n" +
				"Review selected work and its actual declared section instances. Catalog\n" +
				"membership does not select activities, create arrangements or bind recipient\n" +
				"fields to answers. Answer required onboarder questions explicitly or deliberately\n" +
				"defer them under the existing rules; a template default is not an answer.\n" +
				"\n" +
				"Name the preparation's onboarding owner and explicitly designate a known named\n" +
				"business contact with a displayable pre-access help route for export. This is\n" +
				"an organizational assertion, not a network test. Inspect the real first-steps\n" +
				"preview and visible missing information before delivery. Preserve the original\n" +
				"plan, preparation and any newer personal work. Export creates a new archive,\n" +
				"not an import, confirmed identity, first use or verified onboarding outcome.\n",
		},
		{
			summary: contract.Summary{
				ID:          "invitation-start",
				Title:       "Start from the invitation ZIP",
				Category:    "invitation-and-first-steps",
				Goal:        "First interaction",
				Result:      "Understand the package, get the tool, confirm identity, and start without company access",
				Carrier:     "tool-help",
				ContentPath: "help/invitation-start.md",
			},
			content: "# Start from the invitation ZIP\n" +
				"\n" +
				"An invitation contains the organization plan, onboarder-owned preparation,\n" +
				"an identity-only newcomer record seed and first-steps.md. Read its first-steps\n" +
				"document and contact the named onboarding owner if anything is missing. Use\n" +
				"only the supported standalone tool supplied through your organization's\n" +
				"approved route; this catalog invents no download address or installation command.\n" +
				"\n" +
				"Open the invitation with an explicitly selected existing record root:\n" +
				"\n" +
				"```text\n" +
				"onboard-kit start --invitation FILE --root DIR [--home DIR]\n" +
				"```\n" +
				"\n" +
				"Confirm the displayed account is yours, then review every proposed destination\n" +
				"and action. Identity confirmation alone writes nothing. Only choose begin after\n" +
				"the complete review if you want the proposed opening. Cancel before begin leaves\n" +
				"your files unchanged. A ZIP entry name is not permission to extract over local\n" +
				"files. If the account or destinations are wrong, stop and ask the owner.\n" +
				"\n" +
				"Existing matching preparation is retained. Existing records are reconciled with\n" +
				"changes shown in review; their outcomes, notes and first-use history are preserved.\n" +
				"A different local plan needs a separate explicit choice before replacement.\n" +
				"No date decides which copy is newer. Anyone who can read stored records can read\n" +
				"their answers and notes. Private recovery copies may temporarily exist in the\n" +
				"state home and selected artifact parents;\n" +
				"keep credentials and unnecessary personal data out.\n" +
				"\n" +
				"After begin, the tool opens a read-only record-backed session with instructions,\n" +
				"questions, waiting work and available support. It saves no answers or outcomes\n" +
				"and executes no commands or checks. Only a confirmed begin creating a genuinely\n" +
				"new local record records firstOpenedAt. Existing events are preserved; an existing\n" +
				"record without an event stays unknown. This is tool use, not an employment start\n" +
				"date or completion of an activity.\n" +
				"\n" +
				"After leaving that view, the separate session edit command can review and save\n" +
				"one personal answer, follow-plan choice or outcome in the registered context.\n" +
				"Use the explicit absolute root, state home and registered session key described\n" +
				"in Resume onboarding. Opening an invitation does not consent to those changes.\n" +
				"\n" +
				"If opening is interrupted, retain its recovery data and transaction token. Follow\n" +
				"the interrupted-opening recovery instructions in Resume onboarding; recovery\n" +
				"does not need the original ZIP. An invitation export alone sends nothing,\n" +
				"confirms no identity and records no first use. The separate start --org-plan\n" +
				"and walk routes retain their legacy plan-file behavior.\n" +
				"\n" +
				"Use the actual rendered first-steps help and selected business support route\n" +
				"for bootstrap work before company access exists. This static catalog entry\n" +
				"does not render a personal invitation, assert contact availability or duplicate\n" +
				"the first-steps assembler's account-specific guide. Reading a preview or this\n" +
				"entry records no start, accepts no default and verifies no activity.\n",
		},
		{
			summary: contract.Summary{
				ID:          "knowledge-transfer",
				Title:       "Knowledge transfer and onboarding ownership",
				Category:    "feedback-and-progress",
				Goal:        "Throughout; confirm before departure",
				Result:      "Name knowledge recipients and an owner for shared onboarding material",
				Carrier:     "section-template",
				ContentPath: "sections/knowledge-transfer.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: knowledge-transfer\n" +
				"title: Knowledge transfer and onboarding ownership\n" +
				"summary: Name knowledge recipients and an owner for shared onboarding material\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: knowledge_handoff, description: \"Which work knowledge needs transfer, where should it live and how are recipients identified?\", audience: onboarder, required: true}\n" +
				"  - {name: onboarding_ownership, description: \"Who owns shared onboarding material and how are reviewed updates and future handoffs maintained?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Plan the knowledge handoff using $knowledge_handoff.\n" +
				"  Confirm ownership of shared material: $onboarding_ownership\n" +
				"\n" +
				"  Share useful decisions, operational context and open questions with\n" +
				"  the intended recipients while respecting access restrictions. Confirm\n" +
				"  that recipients can find and use the material; sending a link is not\n" +
				"  enough. Keep it current throughout the work and confirm continuity\n" +
				"  before departure. Never transfer credentials or publish personal records\n" +
				"  as a substitute for documented work knowledge.\n" +
				"steps:\n" +
				"  - id: confirm-knowledge-continuity\n" +
				"    name: Confirm recipients and ownership of useful shared knowledge\n" +
				"    checks:\n" +
				"      - ask: Can the intended recipients use the needed knowledge and is an owner identified for shared onboarding material?\n",
		},
		{
			summary: contract.Summary{
				ID:          "meeting-invite",
				Title:       "Send meeting invitations",
				Category:    "onboarder-preparation",
				Goal:        "Before the relevant meeting",
				Result:      "Supply invitations and joining instructions for introductions, regular meetings, retrospectives, and check-ins",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/meeting-invite.md",
			},
			content: "# Send meeting invitations\n" +
				"\n" +
				"Coordinate introductions, regular meetings, retrospectives and check-ins that\n" +
				"are relevant to the selected work. Confirm participants, purpose, time-zone\n" +
				"interpretation and joining instructions without copying private calendar\n" +
				"contents. Keep shared meeting policies in the plan and individual invitations\n" +
				"in preparation; no automatic selection or meeting delivery is implied.\n" +
				"\n" +
				"Use the meeting kind with stable unique arrangement ids and actual declared\n" +
				"section instances. Several arrangements may support one section when useful.\n" +
				"Arranged requires a known owner and nonsecret meeting reference; helpers are\n" +
				"optional for this kind. Pending, requested or deferred retain missing details\n" +
				"visibly. Unknown supplied section or contact references remain invalid.\n" +
				"\n" +
				"Check that joining guidance is usable with the access expected at that time,\n" +
				"and identify who handles missing invitations. Targets are planning goals, not\n" +
				"attendance failures. An invitation sent or an arranged meeting is not proof\n" +
				"of joining, attendance, understanding or a completed conversation. Those\n" +
				"outcomes are separate from onboarder preparation.\n",
		},
		{
			summary: contract.Summary{
				ID:          "observability-access",
				Title:       "Logs, metrics, and tracing access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before diagnosis or deploy work",
				Result:      "Inspect the relevant service",
				Carrier:     "section-template",
				ContentPath: "sections/observability-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: observability-access\n" +
				"title: Logs, metrics, and tracing access\n" +
				"summary: Inspect the relevant service\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: observability_scope, description: \"Which service, approved dashboards and data scope are needed?\", audience: onboarder, required: true}\n" +
				"  - {name: observability_route, description: \"How is access granted, what safe example should be inspected and who handles questions?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Locate the service's approved observability surfaces: $observability_scope\n" +
				"  Use the supplied example and access route: $observability_route\n" +
				"\n" +
				"  Find a relevant metric, log or trace and relate it to service behavior.\n" +
				"  Learn retention and redaction expectations before sharing a sample.\n" +
				"  Observability data may contain personal or secret information; link to\n" +
				"  approved views rather than copying raw data into the record. Read access\n" +
				"  does not authorize changing alerts or querying unrelated systems.\n" +
				"steps:\n" +
				"  - id: inspect-service-signals\n" +
				"    name: Inspect a permitted service signal\n" +
				"    checks:\n" +
				"      - ask: Can you find and interpret a relevant service signal while respecting its data and access restrictions?\n",
		},
		{
			summary: contract.Summary{
				ID:          "office-access",
				Title:       "Office access and security badge",
				Category:    "identity-equipment-and-communication",
				Goal:        "First on-site visit",
				Result:      "Entry and badge work; arrival instructions and contact are clear",
				Carrier:     "section-template",
				ContentPath: "sections/office-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: office-access\n" +
				"title: Office access and security badge\n" +
				"summary: Entry and badge work; arrival instructions and contact are clear\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: arrival_process, description: \"What shared arrival, visitor, badge and physical-security instructions apply?\", audience: onboarder, required: true}\n" +
				"  - {name: arrival_support, description: \"Which business contact and fallback can assist someone who cannot enter?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the approved arrival and badge process: $arrival_process\n" +
				"  If entry fails, use $arrival_support.\n" +
				"\n" +
				"  Confirm your own visit arrangements with the onboarder before travel.\n" +
				"  Test the access needed for your work with an authorized host. Learn\n" +
				"  visitor, lost-badge and emergency procedures. Never borrow another\n" +
				"  person's badge, tailgate or bypass an access control to finish the task.\n" +
				"  Keep access codes and private travel details out of onboarding records.\n" +
				"steps:\n" +
				"  - id: confirm-entry\n" +
				"    name: Verify authorized office entry\n" +
				"    checks:\n" +
				"      - ask: Can you use the intended entry or badge process and identify a reachable host if entry fails?\n",
		},
		{
			summary: contract.Summary{
				ID:          "office-arrange",
				Title:       "Arrange office access and badge",
				Category:    "onboarder-preparation",
				Goal:        "Before the first on-site visit",
				Result:      "Supply the badge route, arrival instructions, and reachable host",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/office-arrange.md",
			},
			content: "# Arrange office access and badge\n" +
				"\n" +
				"Coordinate the approved badge or visitor process, arrival instructions and a\n" +
				"reachable business host before the first on-site visit. Shared security rules\n" +
				"and arrival routes belong in the plan. Keep a person's arrangement in the\n" +
				"office kind, linked to an actual declared office-access section instance.\n" +
				"Do not collect a home address, private itinerary or access code in preparation.\n" +
				"\n" +
				"Use a stable unique id. Before marking arranged, provide a known accountable\n" +
				"owner and a nonsecret reference to the office request or approved instructions.\n" +
				"Add known host contacts when assigned. Pending, requested or deferred can\n" +
				"retain missing information visibly; unknown supplied contacts or sections are\n" +
				"invalid. A targetDate supports coordination rather than enforcing attendance.\n" +
				"\n" +
				"Confirm a fallback if the badge or host is unavailable, including a route that\n" +
				"works before entering. Arranged is not proof that the badge or entry works.\n" +
				"The newcomer verifies authorized access; neither urgency nor onboarding\n" +
				"completion permits borrowed badges or bypassing physical security controls.\n",
		},
		{
			summary: contract.Summary{
				ID:          "onboarding-buddy-meeting",
				Title:       "Meet the onboarding buddy",
				Category:    "team-and-coordination",
				Goal:        "Day one",
				Result:      "Know the day-to-day first contact",
				Carrier:     "section-template",
				ContentPath: "sections/onboarding-buddy-meeting.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: onboarding-buddy-meeting\n" +
				"title: Meet the onboarding buddy\n" +
				"summary: Know the day-to-day first contact\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: buddy_process, description: \"How is a newcomer introduced to their assigned buddy and what support does the buddy provide?\", audience: onboarder, required: true}\n" +
				"  - {name: buddy_backup, description: \"What shared backup route applies when the buddy is unavailable?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Meet your assigned buddy through $buddy_process.\n" +
				"  The backup support route is $buddy_backup.\n" +
				"\n" +
				"  Agree how to ask small day-to-day questions and when a different owner\n" +
				"  should decide. Discuss availability and suitable check-ins without\n" +
				"  assuming the buddy is continuously on call. Keep the personal assignment\n" +
				"  in preparation. If no buddy is assigned, ask the owner to arrange one or\n" +
				"  explicitly agree another support model; do not invent a name.\n" +
				"steps:\n" +
				"  - id: establish-buddy-route\n" +
				"    name: Establish a day-to-day support relationship\n" +
				"    checks:\n" +
				"      - ask: Do you know your agreed day-to-day contact, their support scope and the fallback when they are unavailable?\n",
		},
		{
			summary: contract.Summary{
				ID:          "onboarding-correction",
				Title:       "Report incorrect onboarding instructions",
				Category:    "feedback-and-progress",
				Goal:        "Throughout",
				Result:      "Give the plan owner corrections without silently editing shared instructions",
				Carrier:     "section-template",
				ContentPath: "sections/onboarding-correction.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: onboarding-correction\n" +
				"title: Report incorrect onboarding instructions\n" +
				"summary: Give the plan owner corrections without silently editing shared instructions\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: correction_route, description: \"Where should a newcomer report incorrect shared onboarding instructions?\", audience: onboarder, required: true}\n" +
				"  - {name: correction_review, description: \"Who reviews corrections and how are approved changes shared while preserving personal work?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Send proposed corrections through $correction_route.\n" +
				"  Follow the ownership and review process: $correction_review\n" +
				"\n" +
				"  Identify the affected instruction, the expected behavior and a safe\n" +
				"  explanation of the difference. Keep secrets, personal records and raw\n" +
				"  sensitive logs out of the report. Do not silently rewrite shared facts\n" +
				"  or overwrite local work to match a new copy. If instructions conflict,\n" +
				"  stop the risky action and ask the plan owner which guidance applies.\n" +
				"steps:\n" +
				"  - id: establish-correction-path\n" +
				"    name: Use or identify the reviewed correction process\n" +
				"    checks:\n" +
				"      - ask: Do you know how to give the plan owner a useful correction without exposing private data or silently changing shared instructions?\n",
		},
		{
			summary: contract.Summary{
				ID:          "onboarding-help",
				Title:       "Get help when blocked",
				Category:    "invitation-and-first-steps",
				Goal:        "Throughout",
				Result:      "Have a reachable initial contact and a fallback when the usual route fails",
				Carrier:     "section-template",
				ContentPath: "sections/onboarding-help.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: onboarding-help\n" +
				"title: Get help when blocked\n" +
				"summary: Have a reachable initial contact and a fallback when the usual route fails\n" +
				"timing: anytime\n" +
				"bootstrap: true\n" +
				"variables:\n" +
				"  - {name: general_contacts, description: \"Who can help with an onboarding problem?\", type: contact, multiple: true, audience: onboarder, required: true}\n" +
				"  - {name: preaccess_route, description: \"Which approved business route can the newcomer use before company access?\", audience: onboarder, required: true}\n" +
				"  - {name: fallback_route, description: \"What should the newcomer do if the initial route or usual contact is unavailable?\", audience: onboarder, required: true}\n" +
				"escalation: [$general_contacts]\n" +
				"body: |\n" +
				"  General support: $general_contacts\n" +
				"\n" +
				"  Before company access, use this explicitly supplied route: $preaccess_route\n" +
				"  If it is unavailable, use the fallback: $fallback_route\n" +
				"\n" +
				"  Describe the activity, the nonsecret error and what you already tried.\n" +
				"  Never send passwords, recovery codes, tokens or a complete private record.\n" +
				"  A request being sent is not proof of access. Tell the owner when a blocker\n" +
				"  prevents later work so support can be rearranged without blaming the date.\n" +
				"steps:\n" +
				"  - id: locate-support\n" +
				"    name: Locate the initial and fallback help routes\n" +
				"    checks:\n" +
				"      - ask: Can you identify a route available before company access and a fallback if it fails?\n",
		},
		{
			summary: contract.Summary{
				ID:          "onboarding-owner-meeting",
				Title:       "Meet the onboarding owner or manager",
				Category:    "team-and-coordination",
				Goal:        "Day one",
				Result:      "Understand support, priorities, and the first week's plan",
				Carrier:     "section-template",
				ContentPath: "sections/onboarding-owner-meeting.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: onboarding-owner-meeting\n" +
				"title: Meet the onboarding owner or manager\n" +
				"summary: Understand support, priorities, and the first week's plan\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: owner_meeting_process, description: \"How is the initial owner or manager conversation arranged, and what should it cover?\", audience: onboarder, required: true}\n" +
				"  - {name: priority_context, description: \"What shared priorities and decision context should guide the first week?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Arrange the conversation using $owner_meeting_process.\n" +
				"  Start from the shared priorities: $priority_context\n" +
				"\n" +
				"  Ask your assigned owner to confirm the first week's useful work, how to\n" +
				"  get support and how expectations will be revisited. Discuss dependencies\n" +
				"  and missing arrangements before treating a goal as a commitment. Personal\n" +
				"  manager and owner assignments belong in preparation, not shared answers.\n" +
				"  Leave with a clear next conversation or contact if plans change.\n" +
				"steps:\n" +
				"  - id: agree-initial-plan\n" +
				"    name: Discuss priorities and support with the assigned owner\n" +
				"    checks:\n" +
				"      - ask: Have you discussed initial priorities, available support and how to revisit the first week's plan?\n",
		},
		{
			summary: contract.Summary{
				ID:          "onboarding-refresh",
				Title:       "Get the latest plan",
				Category:    "invitation-and-first-steps",
				Goal:        "After access",
				Result:      "Request an update without losing personal work",
				Carrier:     "section-template",
				ContentPath: "sections/onboarding-refresh.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: onboarding-refresh\n" +
				"title: Get the latest plan\n" +
				"summary: Request an update without losing personal work\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: update_route, description: \"Where should a newcomer request the current organization plan and report a version mismatch?\", audience: onboarder, required: true}\n" +
				"  - {name: preservation_process, description: \"What reviewed procedure preserves newer preparation and personal work when sharing updates?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Ask for the current plan through $update_route.\n" +
				"  Use the reviewed preservation procedure: $preservation_process\n" +
				"\n" +
				"  Compare plan identity and available version information with the owner.\n" +
				"  Preserve your current files before any reviewed replacement. Do not copy\n" +
				"  an older invitation over newer preparation or personal work. A changed\n" +
				"  shared instruction is not permission to erase a personal answer or claim\n" +
				"  a verified outcome. Report conflicts for review. No automatic plan-update\n" +
				"  command or confirmed invitation-import flow is supplied by this section.\n" +
				"steps:\n" +
				"  - id: request-safe-update\n" +
				"    name: Establish a safe route for plan updates\n" +
				"    checks:\n" +
				"      - ask: Do you know how to request the latest plan while preserving and reviewing existing work?\n",
		},
		{
			summary: contract.Summary{
				ID:          "onboarding-repository",
				Title:       "Connect to the onboarding repository",
				Category:    "invitation-and-first-steps",
				Goal:        "After access",
				Result:      "Know where files go and who can read them",
				Carrier:     "section-template",
				ContentPath: "sections/onboarding-repository.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: onboarding-repository\n" +
				"title: Connect to the onboarding repository\n" +
				"summary: Know where files go and who can read them\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: sharing_location, description: \"Where is the approved shared plan location, and how is access requested?\", audience: onboarder, required: true}\n" +
				"  - {name: sharing_policy, description: \"Who may read or update each onboarding artifact, and what approved sharing procedure applies?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  After the relevant access works, locate the shared plan: $sharing_location\n" +
				"  Follow the organization's sharing procedure: $sharing_policy\n" +
				"\n" +
				"  The shared plan, onboarder-owned preparation and newcomer-owned record\n" +
				"  have different owners. Confirm destinations before moving anything.\n" +
				"  Anyone with read access to a stored record can read its answers and notes;\n" +
				"  file ownership alone is not privacy. Keep credentials out of every file.\n" +
				"  Ask the owner about unfamiliar or conflicting files instead of replacing\n" +
				"  them. This activity does not supply an automatic import or Git command.\n" +
				"steps:\n" +
				"  - id: confirm-sharing\n" +
				"    name: Confirm artifact destinations and readership\n" +
				"    checks:\n" +
				"      - ask: Can you locate the approved plan and explain where your artifacts belong and who may read them?\n",
		},
		{
			summary: contract.Summary{
				ID:          "onboarding-resume",
				Title:       "Resume onboarding",
				Category:    "invitation-and-first-steps",
				Goal:        "Throughout",
				Result:      "Return to the tool, record outcomes, and report differing instructions",
				Carrier:     "tool-help",
				ContentPath: "help/onboarding-resume.md",
			},
			content: "# Resume onboarding\n" +
				"\n" +
				"Return through the approved support and sharing procedure for your actual\n" +
				"plan and tool version. Keep your own saved work and ask the onboarding owner\n" +
				"when instructions disagree. Do not replace newer preparation or personal work\n" +
				"with an older invitation just to get back to a familiar screen.\n" +
				"\n" +
				"To recover an interrupted invitation opening, use its reported transaction token\n" +
				"with the same selected record root and state home:\n" +
				"\n" +
				"```text\n" +
				"onboard-kit start --resume-import TOKEN --root DIR [--home DIR]\n" +
				"```\n" +
				"\n" +
				"Recovery does not need the original ZIP. Review the displayed pending actions\n" +
				"before choosing resume. A completed opening awaiting cleanup instead offers\n" +
				"clean; cleanup preserves later owner edits and may finish without displaying a\n" +
				"session when its staged display is gone. Cancel retains recovery state. Keep\n" +
				"recovery data until the tool confirms cleanup; do not guess another destination\n" +
				"or delete unknown files to force progress. Ask the owner when context or stored\n" +
				"work has changed. A reported token is a selector, not proof of valid recovery.\n" +
				"\n" +
				"The token applies only to interrupted opening or its cleanup. It is not a\n" +
				"general command for returning to completed onboarding. The invitation route\n" +
				"opens a read-only record-backed session: existing answers, outcomes and waiting\n" +
				"work can be displayed, but answers and outcomes cannot be edited there and\n" +
				"commands or checks are not executed. Only confirmed begin for a genuinely new\n" +
				"local record records firstOpenedAt; existing events stay unchanged and an\n" +
				"existing missing event stays unknown. Recovery uses its frozen event rather\n" +
				"than sampling first use again.\n" +
				"\n" +
				"The separate start --org-plan route remains the legacy organization-plan flow.\n" +
				"The existing walk command is a legacy plan-file question walkthrough: accepted\n" +
				"answers are saved in that selected plan file. Neither route silently becomes\n" +
				"record-backed answer or outcome editing.\n" +
				"\n" +
				"For a completed invitation session, use its registered session key and explicit\n" +
				"absolute record root and state home:\n" +
				"\n" +
				"```text\n" +
				"onboard-kit session edit --root ABS --home ABS --session KEY\n" +
				"```\n" +
				"\n" +
				"Confirm the displayed account exactly, including its double quotes. Choose one\n" +
				"answer, follow-plan choice or outcome, review the complete proposed effect, then\n" +
				"choose save or cancel. Only the selected record changes; the plan, preparation,\n" +
				"registration, first-use event and historical reference stay intact. Follow-plan\n" +
				"removes one personal override and may reveal an existing shared answer. Defaults\n" +
				"remain suggestions. Outcome verification still requires an explicit statement;\n" +
				"the tool does not run checks. Pending imports and changed source files refuse\n" +
				"rather than bypassing review. Ask the onboarding owner if the registered context\n" +
				"is unknown; do not guess a key or move state to force acceptance.\n" +
				"\n" +
				"In the activity model, verified means the required capability or result was\n" +
				"actually established; failed means an attempted result did not work; deferred\n" +
				"means intentionally postponed. A request sent, an onboarder arrangement or a\n" +
				"date passing is none of those verified outcomes. A displayed instruction or\n" +
				"read-only session does not record that an activity was completed.\n" +
				"\n" +
				"Report blockers, incorrect instructions, missing files, a wrong account or an\n" +
				"unsupported plan version without sharing credentials or complete private records.\n" +
				"Reading this help or rendering a preview records no first use or employment\n" +
				"start date, accepts no unanswered default and completes no activity.\n",
		},
		{
			summary: contract.Summary{
				ID:          "onboarding-welcome",
				Title:       "Welcome and expectations",
				Category:    "invitation-and-first-steps",
				Goal:        "First interaction",
				Result:      "Know the team, owner, start arrangements, and initial goals",
				Carrier:     "section-template",
				ContentPath: "sections/onboarding-welcome.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: onboarding-welcome\n" +
				"title: Welcome and expectations\n" +
				"summary: Know the team, owner, start arrangements, and initial goals\n" +
				"timing: week-one\n" +
				"bootstrap: true\n" +
				"variables:\n" +
				"  - {name: welcome_context, description: \"What should a newcomer know about the team, its purpose and initial goals?\", audience: onboarder, required: true}\n" +
				"  - {name: initial_arrangements, description: \"What shared arrival or remote joining instructions and support route apply before company access?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Welcome to the team. Start with the shared context: $welcome_context\n" +
				"\n" +
				"  Initial joining instructions: $initial_arrangements\n" +
				"\n" +
				"  Ask the onboarding owner to confirm your own arrangements. Individual\n" +
				"  assignments belong in preparation, not shared template answers. If a goal\n" +
				"  or instruction does not fit, discuss it rather than guessing. Planning\n" +
				"  targets guide support; they are not a test of whether you belong here.\n" +
				"steps:\n" +
				"  - id: discuss-expectations\n" +
				"    name: Discuss the first useful work and available support\n" +
				"    checks:\n" +
				"      - ask: Do you know the team's purpose, the first priorities and whom to ask about your arrangements?\n",
		},
		{
			summary: contract.Summary{
				ID:          "pair-assign",
				Title:       "Assign a pairing partner",
				Category:    "onboarder-preparation",
				Goal:        "Before ticket pairing",
				Result:      "Identify a partner and suitable ticket",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/pair-assign.md",
			},
			content: "# Assign a pairing partner\n" +
				"\n" +
				"Arrange a willing partner and suitable real ticket with a clear learning focus.\n" +
				"Keep shared pairing practices in the plan. Store the personal opportunity in\n" +
				"a pairing arrangement for an actual declared ticket-pairing section instance,\n" +
				"not in a shared answer or an invented pair-assign kind.\n" +
				"\n" +
				"Use a stable unique id. Before claiming arranged, supply a known accountable\n" +
				"owner, at least one known helper contact and a nonsecret ticket or opportunity\n" +
				"reference. The owner is accountable for arranging; the helper is the partner.\n" +
				"Pending, requested and deferred may retain unassigned details visibly. A\n" +
				"supplied unknown helper is invalid, not a placeholder for a missing assignment.\n" +
				"\n" +
				"Confirm availability, scope and a support route if the opportunity changes.\n" +
				"The optional real targetDate coordinates people and does not fail the newcomer.\n" +
				"Arranged means a minimum opportunity exists, not that pairing happened or a\n" +
				"ticket was delivered. Preparation edits change neither shared organization\n" +
				"answers nor the newcomer's outcomes or first-use state.\n",
		},
		{
			summary: contract.Summary{
				ID:          "people-map",
				Title:       "Build a people map",
				Category:    "team-and-coordination",
				Goal:        "First month",
				Result:      "Understand important relationships, responsibilities, and dependencies",
				Carrier:     "section-template",
				ContentPath: "sections/people-map.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: people-map\n" +
				"title: Build a people map\n" +
				"summary: Understand important relationships, responsibilities, and dependencies\n" +
				"timing: days-8-to-30\n" +
				"variables:\n" +
				"  - {name: relationship_scope, description: \"Which roles, teams and dependencies matter for this work?\", audience: onboarder, required: true}\n" +
				"  - {name: relationship_reference, description: \"Where is the approved responsibility map and who can correct it?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Focus on the work relationships described in $relationship_scope.\n" +
				"  Use the current reference and correction route: $relationship_reference\n" +
				"\n" +
				"  Map responsibilities, decision owners and handoffs for a real piece of\n" +
				"  work. Confirm uncertain relationships with the people involved. Record\n" +
				"  only useful business context, not private biographies or judgments about\n" +
				"  individuals. Identify where a dependency needs an introduction or clearer\n" +
				"  ownership, and share corrections through the agreed route.\n" +
				"steps:\n" +
				"  - id: explain-key-relationships\n" +
				"    name: Trace responsibilities and work dependencies\n" +
				"    checks:\n" +
				"      - ask: Can you identify the relevant decision owners, collaborators and handoffs for your current work?\n",
		},
		{
			summary: contract.Summary{
				ID:          "pr-merge",
				Title:       "Merge a PR",
				Category:    "software-work-and-delivery",
				Goal:        "When authorized and ready",
				Result:      "Complete an authorized merge after required reviews and checks",
				Carrier:     "section-template",
				ContentPath: "sections/pr-merge.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: pr-merge\n" +
				"title: Merge a PR\n" +
				"summary: Complete an authorized merge after required reviews and checks\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: merge_policy, description: \"Does team policy permit merging your own PR, another person's PR, or only a designated subset, and what approvals and checks are mandatory?\", audience: onboarder, required: true}\n" +
				"  - {name: merge_process, description: \"What approved merge procedure, authorization confirmation and support route apply?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Confirm the explicit team decision before selecting an opportunity:\n" +
				"  $merge_policy\n" +
				"\n" +
				"  Follow the approved procedure and authorization route: $merge_process\n" +
				"\n" +
				"  Your own PR or another person's PR may be suitable only when this policy\n" +
				"  permits it. Confirm that this particular merge is authorized, reviews and\n" +
				"  checks are current, and any separation-of-duties rules are satisfied.\n" +
				"  Recheck the intended branch and merge result. Never bypass protection,\n" +
				"  self-authorize or obtain extra rights merely to complete onboarding.\n" +
				"steps:\n" +
				"  - id: complete-authorized-merge\n" +
				"    name: Complete a policy-permitted and authorized merge\n" +
				"    checks:\n" +
				"      - ask: Did you complete an authorized merge of a policy-permitted PR after all required reviews and checks, without bypassing separation of duties?\n",
		},
		{
			summary: contract.Summary{
				ID:          "pr-merge-arrange",
				Title:       "Arrange a merge opportunity",
				Category:    "onboarder-preparation",
				Goal:        "Before merging a PR",
				Result:      "Confirm a suitable opportunity, authorization, and any support",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/pr-merge-arrange.md",
			},
			content: "# Arrange a merge opportunity\n" +
				"\n" +
				"Ask the team to state whether this newcomer may merge their own PR, another\n" +
				"person's PR or only a defined subset. Keep that policy and mandatory reviews,\n" +
				"checks and separation of duties in the plan. Do not infer authorization from\n" +
				"source-host permissions or grant more access merely to complete onboarding.\n" +
				"\n" +
				"Use pr-merge for an opportunity linked to an actual declared merge section\n" +
				"instance. Arranged requires a known accountable owner, at least one known\n" +
				"helper and a nonsecret PR or opportunity reference. Keep a stable unique id.\n" +
				"Pending, requested or deferred retain missing information visibly; unknown\n" +
				"supplied references are invalid. Targets are optional real planning dates.\n" +
				"\n" +
				"Confirm that the selected opportunity is policy-permitted, that authorization\n" +
				"will be checked at action time and that support is available. Either author's\n" +
				"PR may serve only if team policy permits it. Arranged never means merged,\n" +
				"approved or verified. If a safe authorized opportunity is absent, keep the gap\n" +
				"instead of bypassing protection or manufacturing an outcome.\n",
		},
		{
			summary: contract.Summary{
				ID:          "pr-post",
				Title:       "Post a PR",
				Category:    "software-work-and-delivery",
				Goal:        "First suitable contribution",
				Result:      "Submit a real change with context, checks, and review requests",
				Carrier:     "section-template",
				ContentPath: "sections/pr-post.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: pr-post\n" +
				"title: Post a PR\n" +
				"summary: Submit a real change with context, checks, and review requests\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: posting_practice, description: \"What PR content, branch, check and review-request practices apply?\", audience: onboarder, required: true}\n" +
				"  - {name: posting_support, description: \"How is an appropriate reviewer identified and where should posting problems be raised?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Prepare the change using $posting_practice.\n" +
				"  Identify review support through $posting_support.\n" +
				"\n" +
				"  Submit a real, appropriately scoped change with its purpose, relevant\n" +
				"  evidence and known limitations. Inspect the diff for unintended changes\n" +
				"  or secrets and report actual check results. Request the intended review\n" +
				"  without assuming that posting means approval or readiness to merge.\n" +
				"  Personal reviewer assignments and opportunities belong in preparation.\n" +
				"steps:\n" +
				"  - id: submit-real-change\n" +
				"    name: Submit a clear change for the required review\n" +
				"    checks:\n" +
				"      - ask: Have you posted a real change with useful context, honest check results and the appropriate review request?\n",
		},
		{
			summary: contract.Summary{
				ID:          "pr-review",
				Title:       "Review a PR",
				Category:    "software-work-and-delivery",
				Goal:        "First suitable opportunity",
				Result:      "Give a substantive review following team practices",
				Carrier:     "section-template",
				ContentPath: "sections/pr-review.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: pr-review\n" +
				"title: Review a PR\n" +
				"summary: Give a substantive review following team practices\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: review_practice, description: \"What review scope, conduct and approval rules apply?\", audience: onboarder, required: true}\n" +
				"  - {name: review_support, description: \"How is a suitable review opportunity arranged and who can discuss uncertain findings?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the review practice: $review_practice\n" +
				"  Arrange and discuss the opportunity through $review_support.\n" +
				"\n" +
				"  Confirm the selected PR and discussion contact in preparation. Understand\n" +
				"  its purpose, inspect the change and tests, and give specific useful\n" +
				"  feedback. Separate questions, suggestions and blocking concerns. If you\n" +
				"  lack context or approval authority, ask rather than approving by habit.\n" +
				"  A substantive review can surface uncertainty; it need not manufacture a\n" +
				"  defect or satisfy an arbitrary comment count.\n" +
				"steps:\n" +
				"  - id: provide-substantive-review\n" +
				"    name: Review the intended change and discuss useful feedback\n" +
				"    checks:\n" +
				"      - ask: Have you reviewed the change against team practices and communicated substantive findings or reasoned approval within your authority?\n",
		},
		{
			summary: contract.Summary{
				ID:          "pr-review-arrange",
				Title:       "Arrange a PR review opportunity",
				Category:    "onboarder-preparation",
				Goal:        "Before PR review",
				Result:      "Identify a suitable change and discussion contact",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/pr-review-arrange.md",
			},
			content: "# Arrange a PR review opportunity\n" +
				"\n" +
				"Identify a suitable real change and an experienced discussion contact. Match\n" +
				"the opportunity to the newcomer's context and authority. Shared review rules\n" +
				"stay in the plan; the person-specific opportunity uses pr-review and an actual\n" +
				"declared review section instance. Catalog selection alone creates no instance.\n" +
				"\n" +
				"Before marking arranged, supply a known owner, at least one known helper and\n" +
				"a nonsecret PR reference. Keep a stable unique arrangement id. Missing owner,\n" +
				"opportunity or helper remains visible while pending, requested or deferred;\n" +
				"unknown supplied references remain invalid. A targetDate is optional and a\n" +
				"planning goal. Do not include restricted patch contents or credentials in notes.\n" +
				"\n" +
				"Confirm who can answer questions and whether the newcomer can approve or only\n" +
				"comment. Do not grant extra approval authority for onboarding. Arranged is\n" +
				"not a substantive review, an approval or a newcomer outcome. Preserve honest\n" +
				"gaps if the PR changes, closes or becomes unsuitable before the activity.\n",
		},
		{
			summary: contract.Summary{
				ID:          "pr-reviewer-assign",
				Title:       "Assign a reviewer",
				Category:    "onboarder-preparation",
				Goal:        "Before posting a PR",
				Result:      "Identify a reviewer and review route",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/pr-reviewer-assign.md",
			},
			content: "# Assign a reviewer\n" +
				"\n" +
				"Identify an appropriate available reviewer and the normal review request route\n" +
				"for a first real change. Keep shared branching and review rules in the plan.\n" +
				"Use pr-post for the personal arrangement and link an actual declared posting\n" +
				"section instance. There is no pr-reviewer-assign arrangement kind.\n" +
				"\n" +
				"An arranged item requires a known owner, at least one known helper contact\n" +
				"and a nonsecret change or opportunity reference. The helper is the assigned\n" +
				"reviewer, not automatically the owner. Use a stable unique id; an optional\n" +
				"real targetDate coordinates availability. Pending, requested and deferred can\n" +
				"retain missing assignments visibly, but unknown supplied contacts are invalid.\n" +
				"\n" +
				"Discuss expected context, checks and how to handle reviewer absence. A named\n" +
				"reviewer is not a review request sent, approval granted or change delivered.\n" +
				"Do not invent a reviewer from a guessed variable name or write a personal\n" +
				"arrangement into shared answers. Preparation status changes no newcomer record.\n",
		},
		{
			summary: contract.Summary{
				ID:          "product-orientation",
				Title:       "Understand the product and users",
				Category:    "team-and-coordination",
				Goal:        "First week",
				Result:      "Explain the users, product purpose, and current priorities",
				Carrier:     "section-template",
				ContentPath: "sections/product-orientation.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: product-orientation\n" +
				"title: Understand the product and users\n" +
				"summary: Explain the users, product purpose, and current priorities\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: product_material, description: \"Which product introduction, safe demonstration or user context should be reviewed?\", audience: onboarder, required: true}\n" +
				"  - {name: product_contact, description: \"Who can explain current priorities and answer product questions?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Explore the approved product introduction: $product_material\n" +
				"  Discuss priorities and questions with $product_contact.\n" +
				"\n" +
				"  Explain who uses the product, which problem it addresses and how the\n" +
				"  team learns whether it helps. Trace one ordinary user journey using\n" +
				"  permitted material, not private customer data. Ask how current priorities\n" +
				"  connect to that journey and where assumptions remain uncertain. Share\n" +
				"  misunderstandings early rather than treating a presentation as proof\n" +
				"  that every product decision is understood.\n" +
				"steps:\n" +
				"  - id: explain-product-purpose\n" +
				"    name: Explain the product and its current priorities\n" +
				"    checks:\n" +
				"      - ask: Can you explain the intended users, product purpose and a current priority, including any open questions?\n",
		},
		{
			summary: contract.Summary{
				ID:          "project-dependencies",
				Title:       "Install project dependencies",
				Category:    "software-work-and-delivery",
				Goal:        "First week",
				Result:      "Complete the documented setup",
				Carrier:     "section-template",
				ContentPath: "sections/project-dependencies.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: project-dependencies\n" +
				"title: Install project dependencies\n" +
				"summary: Complete the documented setup\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: dependency_setup, description: \"Which reviewed project dependency setup instructions and versions apply?\", audience: onboarder, required: true}\n" +
				"  - {name: dependency_verification, description: \"What safe check confirms dependency setup and where should failures be discussed?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the setup instructions before using them: $dependency_setup\n" +
				"  Verify with the documented check and support route: $dependency_verification\n" +
				"\n" +
				"  Confirm the correct project and environment. Understand downloads,\n" +
				"  permissions and changes before following a command. Prefer the project's\n" +
				"  documented versions rather than an arbitrary latest release. Do not\n" +
				"  disable integrity checks, add credentials to files or change system-wide\n" +
				"  settings without approval. Record nonsecret setup discrepancies for review.\n" +
				"steps:\n" +
				"  - id: verify-dependency-setup\n" +
				"    name: Complete and verify the documented dependency setup\n" +
				"    checks:\n" +
				"      - ask: Have you completed the documented dependency setup and its verification successfully?\n",
		},
		{
			summary: contract.Summary{
				ID:          "project-tests",
				Title:       "Run the test suite",
				Category:    "software-work-and-delivery",
				Goal:        "First week",
				Result:      "Run tests and identify known failures rather than hide them",
				Carrier:     "section-template",
				ContentPath: "sections/project-tests.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: project-tests\n" +
				"title: Run the test suite\n" +
				"summary: Run tests and identify known failures rather than hide them\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: test_instructions, description: \"Which reviewed test instructions, environment and scope apply?\", audience: onboarder, required: true}\n" +
				"  - {name: test_baseline, description: \"Where are known test failures documented and who can help distinguish setup from product failures?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review and follow the test instructions: $test_instructions\n" +
				"  Compare results and discuss failures through $test_baseline.\n" +
				"\n" +
				"  Confirm whether tests affect external services or persistent data before\n" +
				"  running them yourself. Inspect the actual result rather than assuming\n" +
				"  success from a command starting. Distinguish known baseline failures,\n" +
				"  environment problems and new failures. Do not skip tests, rewrite their\n" +
				"  expectations or hide failing output to complete onboarding. This plain\n" +
				"  instruction template executes no command.\n" +
				"steps:\n" +
				"  - id: understand-test-results\n" +
				"    name: Run the intended tests and interpret the result\n" +
				"    checks:\n" +
				"      - ask: Have you run the intended tests and identified passing results or the specific unresolved and known failures?\n",
		},
		{
			summary: contract.Summary{
				ID:          "reading-assign",
				Title:       "Assign reading",
				Category:    "onboarder-preparation",
				Goal:        "Before the reading activity",
				Result:      "Supply material, purpose, target, and a question contact",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/reading-assign.md",
			},
			content: "# Assign reading\n" +
				"\n" +
				"Choose current material for a stated work purpose. Supply the question or\n" +
				"decision the reading should inform and a business contact for discussion.\n" +
				"Keep shared documentation routes in the plan. A person's assignment uses the\n" +
				"existing reading kind and an actual declared document-reading section instance;\n" +
				"use separate instances when reading outcomes are independent.\n" +
				"\n" +
				"An arranged reading item needs a known owner and a nonsecret material reference.\n" +
				"Helper contacts are optional for this kind, though a discussion route should\n" +
				"be clear. Keep a stable unique id and an optional real planning target. Missing\n" +
				"details may remain pending, requested or deferred; supplied unknown section\n" +
				"or contact references are invalid. Do not copy restricted content into a note.\n" +
				"\n" +
				"Check that the material is accessible for the intended role or arrange an\n" +
				"approved alternative. A link sent or an arranged assignment is not evidence\n" +
				"of reading or understanding. Preserve open questions and stale-content reports\n" +
				"instead of marking the newcomer's outcome from onboarder task status.\n",
		},
		{
			summary: contract.Summary{
				ID:          "regular-meetings",
				Title:       "Join regular team meetings",
				Category:    "team-and-coordination",
				Goal:        "First week",
				Result:      "Know which meetings matter and be able to attend",
				Carrier:     "section-template",
				ContentPath: "sections/regular-meetings.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: regular-meetings\n" +
				"title: Join regular team meetings\n" +
				"summary: Know which meetings matter and be able to attend\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: meeting_purposes, description: \"Which regular meetings matter, what are their purposes and which participation expectations apply?\", audience: onboarder, required: true}\n" +
				"  - {name: meeting_access, description: \"How are invitations, joining details and missing-access support obtained?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the relevant meetings and purposes: $meeting_purposes\n" +
				"  Obtain invitations and joining help through $meeting_access.\n" +
				"\n" +
				"  Confirm the schedule in your time zone and distinguish required,\n" +
				"  optional and role-specific attendance. Join an appropriate meeting and\n" +
				"  learn how decisions or actions are recorded. If invitations are missing\n" +
				"  or overlap creates a problem, ask the meeting owner; being invited alone\n" +
				"  is not proof that joining works or attendance is useful.\n" +
				"steps:\n" +
				"  - id: join-relevant-meeting\n" +
				"    name: Join a relevant team meeting\n" +
				"    checks:\n" +
				"      - ask: Do you understand the relevant meetings and can you join with the required access and context?\n",
		},
		{
			summary: contract.Summary{
				ID:          "remote-arrange",
				Title:       "Arrange remote onboarding",
				Category:    "onboarder-preparation",
				Goal:        "Before remote or hybrid start",
				Result:      "Arrange equipment, connection instructions, and a remote first-day contact",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/remote-arrange.md",
			},
			content: "# Arrange remote onboarding\n" +
				"\n" +
				"Coordinate approved equipment, connection guidance and a remote first-day\n" +
				"business contact. Hybrid work may also need office preparation; neither mode\n" +
				"automatically replaces the other. Shared remote instructions stay in the plan.\n" +
				"Recipient workLocation may be remote, on-site, hybrid or unfilled; timeZone is\n" +
				"preferable to collecting precise location. Do not invent preferences.\n" +
				"\n" +
				"Use the existing remote kind for a person's actual declared remote-related\n" +
				"section instance. An arranged item needs a known owner and nonsecret reference.\n" +
				"Add known helpers when available; do not infer one from directory order.\n" +
				"Pending, requested and deferred may keep gaps visible, but unknown supplied\n" +
				"references are invalid. Use stable unique ids and optional real planning dates.\n" +
				"\n" +
				"Confirm how to get support if company systems or the intended connection do\n" +
				"not work yet. Keep personal phone numbers, home addresses and credentials out\n" +
				"of preparation. Arranged support is not verification of connectivity or device\n" +
				"usability. Updating these arrangements changes no shared answers, newcomer\n" +
				"record or first-use timestamp.\n",
		},
		{
			summary: contract.Summary{
				ID:          "repository-access",
				Title:       "Repository access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before work in that repository",
				Result:      "Use the required capability for that specific repository",
				Carrier:     "section-template",
				ContentPath: "sections/repository-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: repository-access\n" +
				"title: Repository access\n" +
				"summary: Use the required capability for that specific repository\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: repository_scope, description: \"Which repository and exact capability does this activity cover?\", audience: onboarder, required: true}\n" +
				"  - {name: repository_grant_route, description: \"Who grants this capability and what safe verification procedure applies?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Confirm the repository and required capability: $repository_scope\n" +
				"  Request and verify it through $repository_grant_route.\n" +
				"\n" +
				"  Use your intended account to perform a permitted operation that proves\n" +
				"  the required capability. Reading does not prove write or merge authority;\n" +
				"  a request does not prove access. Do not push a test change or grant broad\n" +
				"  permissions without approval. Configure separate section instances for\n" +
				"  repositories with independent outcomes. Report wrong-account access or\n" +
				"  excessive permissions to the grantor.\n" +
				"steps:\n" +
				"  - id: verify-repository-capability\n" +
				"    name: Verify the specific approved repository capability\n" +
				"    checks:\n" +
				"      - ask: Have you used the approved capability in the intended repository without exceeding its permission scope?\n",
		},
		{
			summary: contract.Summary{
				ID:          "repository-clone",
				Title:       "Clone a repository",
				Category:    "software-work-and-delivery",
				Goal:        "First week",
				Result:      "Have a usable local checkout",
				Carrier:     "section-template",
				ContentPath: "sections/repository-clone.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: repository-clone\n" +
				"title: Clone a repository\n" +
				"summary: Have a usable local checkout\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: clone_instructions, description: \"Which repository, approved clone procedure and local destination guidance apply?\", audience: onboarder, required: true}\n" +
				"  - {name: checkout_verification, description: \"How should the intended checkout and branch be verified, and who helps with failures?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the reviewed clone instructions: $clone_instructions\n" +
				"  Verify the result using $checkout_verification.\n" +
				"\n" +
				"  Confirm the repository identity before obtaining it. Choose a destination\n" +
				"  that does not replace existing work. Inspect project instructions before\n" +
				"  executing repository content. Check that the local checkout can be read\n" +
				"  and identifies the intended repository and branch. This section stores\n" +
				"  instructions only; it does not run Git or assume a clone command.\n" +
				"steps:\n" +
				"  - id: verify-local-checkout\n" +
				"    name: Obtain and verify the intended local checkout\n" +
				"    checks:\n" +
				"      - ask: Do you have a usable checkout of the intended repository without overwriting existing local work?\n",
		},
		{
			summary: contract.Summary{
				ID:          "research-review",
				Title:       "Review user research",
				Category:    "design-work",
				Goal:        "First weeks",
				Result:      "Understand relevant findings and open questions",
				Carrier:     "section-template",
				ContentPath: "sections/research-review.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: research-review\n" +
				"title: Review user research\n" +
				"summary: Understand relevant findings and open questions\n" +
				"timing: days-8-to-30\n" +
				"variables:\n" +
				"  - {name: research_material, description: \"Which approved findings and research context are relevant to current work?\", audience: onboarder, required: true}\n" +
				"  - {name: research_discussion, description: \"Who can explain evidence limitations, consent restrictions and open research questions?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the permitted research material: $research_material\n" +
				"  Discuss evidence and limitations through $research_discussion.\n" +
				"\n" +
				"  Distinguish observations, interpretations and hypotheses. Explain which\n" +
				"  decisions a finding supports and where uncertainty remains. Respect\n" +
				"  participant consent and restricted recordings; summarize useful insights\n" +
				"  without copying identifying material into onboarding files. Ask what is\n" +
				"  missing before treating historical findings as current universal truth.\n" +
				"steps:\n" +
				"  - id: explain-research-evidence\n" +
				"    name: Discuss findings, limits and useful next questions\n" +
				"    checks:\n" +
				"      - ask: Can you explain relevant research findings, their limitations and an open question without exposing participant details?\n",
		},
		{
			summary: contract.Summary{
				ID:          "resource-access-arrange",
				Title:       "Arrange resource access",
				Category:    "onboarder-preparation",
				Goal:        "Before dependent work",
				Result:      "Identify the grantor, permissions, and request route; track requests for selected resources",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/resource-access-arrange.md",
			},
			content: "# Arrange resource access\n" +
				"\n" +
				"Identify each required resource, minimum capability, accountable grantor and\n" +
				"approved request route. Shared policies belong in the plan. Person-specific\n" +
				"requests use the existing access kind and actual declared resource section\n" +
				"instances. Do not grant blanket privileges or treat a catalog identity as a\n" +
				"generated repeated instance. Keep independent requests distinct when useful.\n" +
				"\n" +
				"Use stable unique ids. Arranged needs a known accountable owner and a nonsecret\n" +
				"resource or request reference. Helpers are optional for access. Pending,\n" +
				"requested and deferred may show missing details, while supplied unknown\n" +
				"contacts or sections are invalid. Targets are optional real planning dates.\n" +
				"Never put credentials, private tokens or secret request links in preparation.\n" +
				"\n" +
				"Confirm how the newcomer will safely verify the intended capability and who\n" +
				"handles refusal or excessive permissions. A request, membership or arranged\n" +
				"status does not prove use of the resource. No preparation status grants\n" +
				"authorization, verifies access or changes the newcomer-owned record.\n",
		},
		{
			summary: contract.Summary{
				ID:          "retrospective-participation",
				Title:       "Participate in a retrospective",
				Category:    "feedback-and-progress",
				Goal:        "First available retrospective",
				Result:      "Participate and understand how improvements are followed up",
				Carrier:     "section-template",
				ContentPath: "sections/retrospective-participation.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: retrospective-participation\n" +
				"title: Participate in a retrospective\n" +
				"summary: Participate and understand how improvements are followed up\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: retrospective_process, description: \"How is the team's retrospective run and how does a newcomer join?\", audience: onboarder, required: true}\n" +
				"  - {name: retrospective_followup, description: \"What confidentiality, facilitation and improvement follow-up practices apply?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Join an appropriate retrospective using $retrospective_process.\n" +
				"  Respect its discussion and follow-up rules: $retrospective_followup\n" +
				"\n" +
				"  Learn how observations become experiments or actions and who follows\n" +
				"  through. Contribute a useful observation or question without being\n" +
				"  pressured to reveal private concerns publicly. Do not copy confidential\n" +
				"  discussion or judgments about colleagues into onboarding notes. Ask how\n" +
				"  to raise a concern outside the meeting when that is the safer route.\n" +
				"steps:\n" +
				"  - id: participate-and-follow-up\n" +
				"    name: Participate and understand improvement follow-up\n" +
				"    checks:\n" +
				"      - ask: Have you participated appropriately and learned how retrospective improvements are owned and revisited?\n",
		},
		{
			summary: contract.Summary{
				ID:          "secrets-access",
				Title:       "Secrets manager access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "When needed",
				Result:      "Obtain necessary secrets through the approved system",
				Carrier:     "section-template",
				ContentPath: "sections/secrets-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: secrets-access\n" +
				"title: Secrets manager access\n" +
				"summary: Obtain necessary secrets through the approved system\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: secret_scope, description: \"Which approved secret-management system and minimum secret scope does the role need?\", audience: onboarder, required: true}\n" +
				"  - {name: secret_handling, description: \"What approved access, consumption, incident and support procedure applies without revealing secret values?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Confirm the approved system and minimum scope: $secret_scope\n" +
				"  Follow the handling and support procedure: $secret_handling\n" +
				"\n" +
				"  Verify that the required application or approved workflow can consume\n" +
				"  the needed secret without copying it into onboarding material. Do not\n" +
				"  print values, share credentials, store recovery codes or attach a secret\n" +
				"  to a support request. Access approval is not permission to use every\n" +
				"  secret. Report accidental exposure through the security process.\n" +
				"steps:\n" +
				"  - id: verify-approved-secret-workflow\n" +
				"    name: Verify the approved secret-consumption workflow\n" +
				"    checks:\n" +
				"      - ask: Can the approved workflow obtain the necessary secret without exposing its value in onboarding files or messages?\n",
		},
		{
			summary: contract.Summary{
				ID:          "security-orientation",
				Title:       "Security and acceptable-use orientation",
				Category:    "identity-equipment-and-communication",
				Goal:        "Before affected work",
				Result:      "Understand required practices and support routes",
				Carrier:     "section-template",
				ContentPath: "sections/security-orientation.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: security-orientation\n" +
				"title: Security and acceptable-use orientation\n" +
				"summary: Understand required practices and support routes\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: security_material, description: \"Which current security and acceptable-use material must a newcomer review?\", audience: onboarder, required: true}\n" +
				"  - {name: security_support, description: \"How should someone report a security concern or ask about an unclear rule?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review the required material: $security_material\n" +
				"  Questions and concerns go through $security_support.\n" +
				"\n" +
				"  Identify the rules for handling information, devices, authentication and\n" +
				"  external tools before affected work. Ask about unclear cases rather than\n" +
				"  assuming permission. Learn how to report suspicious events without\n" +
				"  spreading secret material. Reading this section does not grant access,\n" +
				"  authorize testing systems or replace required organizational training.\n" +
				"steps:\n" +
				"  - id: explain-security-route\n" +
				"    name: Review the rules and locate the security support path\n" +
				"    checks:\n" +
				"      - ask: Can you explain the practices relevant to your work and the approved route for concerns or uncertain cases?\n",
		},
		{
			summary: contract.Summary{
				ID:          "source-host-account",
				Title:       "GitHub or other source-host account",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before repository work",
				Result:      "The appropriate named account and authentication work",
				Carrier:     "section-template",
				ContentPath: "sections/source-host-account.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: source-host-account\n" +
				"title: GitHub or other source-host account\n" +
				"summary: The appropriate named account and authentication work\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: source_account_policy, description: \"Which source host, account policy and approved authentication setup apply?\", audience: onboarder, required: true}\n" +
				"  - {name: source_account_support, description: \"Who resolves account or authentication problems before repository access is requested?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the configured host and account policy: $source_account_policy\n" +
				"  Ask for account help through $source_account_support.\n" +
				"\n" +
				"  Confirm which named account should be associated with the organization.\n" +
				"  Verify sign-in and required authentication without sharing credentials\n" +
				"  or saving recovery material in the plan. Account creation does not imply\n" +
				"  organization membership or repository permission. Do not create duplicate\n" +
				"  identities or assume a particular source-host provider.\n" +
				"steps:\n" +
				"  - id: verify-source-identity\n" +
				"    name: Verify the intended source-host identity\n" +
				"    checks:\n" +
				"      - ask: Can you authenticate to the configured source host with the account approved for this work?\n",
		},
		{
			summary: contract.Summary{
				ID:          "source-host-membership",
				Title:       "Source-host organization membership",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before repository work",
				Result:      "Join the required organization or group",
				Carrier:     "section-template",
				ContentPath: "sections/source-host-membership.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: source-host-membership\n" +
				"title: Source-host organization membership\n" +
				"summary: Join the required organization or group\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: membership_scope, description: \"Which source-host organization or group and membership level are required?\", audience: onboarder, required: true}\n" +
				"  - {name: membership_route, description: \"Who grants membership and what approved request and invitation process applies?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Request only the membership described in $membership_scope.\n" +
				"  Follow the grantor and invitation process: $membership_route\n" +
				"\n" +
				"  Check that any invitation targets your intended account and the correct\n" +
				"  organization. After approval, verify the membership is visible with the\n" +
				"  expected scope. Membership does not establish access to every repository;\n" +
				"  ask for the specific capability required by assigned work. Report an\n" +
				"  expired invitation or wrong account rather than accepting an ambiguous one.\n" +
				"steps:\n" +
				"  - id: verify-membership\n" +
				"    name: Verify the required organization or group membership\n" +
				"    checks:\n" +
				"      - ask: Can you confirm the intended account belongs to the required organization or group at the approved level?\n",
		},
		{
			summary: contract.Summary{
				ID:          "support-assign",
				Title:       "Assign onboarding support",
				Category:    "onboarder-preparation",
				Goal:        "Before introductions",
				Result:      "Identify manager, onboarding owner, buddy, and initial help contacts as applicable",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/support-assign.md",
			},
			content: "# Assign onboarding support\n" +
				"\n" +
				"Identify the manager, onboarding owner, buddy and initial help contacts as\n" +
				"applicable. Recipient manager, onboardingOwner and buddy reference known contacts\n" +
				"in the plan directory; do not invent names or assume the first directory entry.\n" +
				"Keep shared support routes and policies in the plan and individual assignments\n" +
				"in this person's preparation. Explain any missing assignment visibly.\n" +
				"\n" +
				"When a support arrangement is useful, use the existing support kind and an\n" +
				"actual declared section instance. Do not manufacture an instance or add a\n" +
				"support-assign arrangement kind. The accountable owner and assigned helpers\n" +
				"are distinct roles, even when the same known contact fills both.\n" +
				"\n" +
				"Pending, requested and deferred describe preparation, not newcomer outcomes.\n" +
				"An arranged support item requires an owner; its reference is optional under\n" +
				"the current support rules. Unknown contacts are invalid even while pending.\n" +
				"Confirm a reachable initial route and fallback with the people involved.\n" +
				"An assignment or arranged status is not proof that help is available, access\n" +
				"works or the newcomer has met anyone. Dates are support planning goals only.\n",
		},
		{
			summary: contract.Summary{
				ID:          "team-chat",
				Title:       "Team chat",
				Category:    "identity-equipment-and-communication",
				Goal:        "Day one",
				Result:      "Reach relevant team and support channels",
				Carrier:     "section-template",
				ContentPath: "sections/team-chat.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: team-chat\n" +
				"title: Team chat\n" +
				"summary: Reach relevant team and support channels\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: chat_setup, description: \"Which chat service and approved account or membership instructions apply?\", audience: onboarder, required: true}\n" +
				"  - {name: chat_channels, description: \"Which team and support channels matter, and what are their purposes and fallback routes?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Set up the configured chat service: $chat_setup\n" +
				"  Find the relevant channels and fallback routes: $chat_channels\n" +
				"\n" +
				"  Confirm you joined the intended workspace with the appropriate account.\n" +
				"  Send a nonsecret introduction or test message in a suitable channel and\n" +
				"  confirm someone can respond. Learn response-time expectations and when\n" +
				"  to use a different support route. Do not assume a provider or request\n" +
				"  broad workspace permissions merely to finish onboarding.\n" +
				"steps:\n" +
				"  - id: reach-team\n" +
				"    name: Reach the team and locate support in chat\n" +
				"    checks:\n" +
				"      - ask: Can you send and receive a message in the relevant team channel and locate its support route?\n",
		},
		{
			summary: contract.Summary{
				ID:          "teammate-introduction",
				Title:       "Meet a teammate",
				Category:    "team-and-coordination",
				Goal:        "First week",
				Result:      "Know their responsibilities and when to ask them",
				Carrier:     "section-template",
				ContentPath: "sections/teammate-introduction.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: teammate-introduction\n" +
				"title: Meet a teammate\n" +
				"summary: Know their responsibilities and when to ask them\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: introduction_scope, description: \"Which shared role or responsibility should this introduction explain?\", audience: onboarder, required: true}\n" +
				"  - {name: introduction_route, description: \"How should this introduction be arranged and what questions are useful?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Learn about this part of the team: $introduction_scope\n" +
				"  Arrange the conversation through $introduction_route.\n" +
				"\n" +
				"  Ask what the teammate owns, how their work connects to yours and when\n" +
				"  they prefer to be consulted. Offer a short introduction to your own work\n" +
				"  without collecting personal details. For several independent people or\n" +
				"  roles, ask the plan author for separate configured activities rather than\n" +
				"  treating one conversation as evidence that all introductions happened.\n" +
				"steps:\n" +
				"  - id: understand-responsibility\n" +
				"    name: Discuss responsibilities and collaboration\n" +
				"    checks:\n" +
				"      - ask: Can you explain this teammate's responsibilities and when to involve them?\n",
		},
		{
			summary: contract.Summary{
				ID:          "ticket-assign",
				Title:       "Identify a first owned ticket",
				Category:    "onboarder-preparation",
				Goal:        "Before ticket ownership",
				Result:      "Identify a suitable task and support contact",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/ticket-assign.md",
			},
			content: "# Identify a first owned ticket\n" +
				"\n" +
				"Agree a suitably scoped real task, expected result and support contact. Avoid\n" +
				"selecting a task solely to satisfy an onboarding counter. Keep the shared\n" +
				"delivery process in the plan and the personal opportunity in a ticket\n" +
				"arrangement linked to an actual declared ticket-ownership section instance.\n" +
				"\n" +
				"An arranged ticket item requires a known accountable owner, at least one known\n" +
				"helper and a nonsecret task reference. Give it a stable unique id. Pending,\n" +
				"requested and deferred can show missing details; unknown supplied contacts\n" +
				"or sections are invalid. An optional real targetDate is a planning aid, not\n" +
				"a deadline that authorizes work or fails someone.\n" +
				"\n" +
				"Discuss permission boundaries, dependencies and how to request scope changes.\n" +
				"If no suitable task exists, retain the gap and arrange support rather than\n" +
				"inventing a ticket. Assigned or arranged does not mean accepted, delivered or\n" +
				"verified by the newcomer. Do not update the personal record or populate\n" +
				"answers through guessed recipient-to-template bindings.\n",
		},
		{
			summary: contract.Summary{
				ID:          "ticket-ownership",
				Title:       "Own a ticket",
				Category:    "software-work-and-delivery",
				Goal:        "First week or month",
				Result:      "Take a suitable task through the team's delivery process",
				Carrier:     "section-template",
				ContentPath: "sections/ticket-ownership.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: ticket-ownership\n" +
				"title: Own a ticket\n" +
				"summary: Take a suitable task through the team's delivery process\n" +
				"timing: days-8-to-30\n" +
				"variables:\n" +
				"  - {name: ownership_process, description: \"What task ownership, scope agreement and delivery process should be followed?\", audience: onboarder, required: true}\n" +
				"  - {name: ownership_support, description: \"Which shared support route and completion evidence should guide a first owned task?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the team's ownership process: $ownership_process\n" +
				"  Use the support and completion guidance: $ownership_support\n" +
				"\n" +
				"  Confirm a suitable assigned ticket and support contact in preparation.\n" +
				"  Clarify the outcome and acceptance conditions before changing anything.\n" +
				"  Keep collaborators informed, seek required review and verify the delivered\n" +
				"  result. Ask for help when scope or access changes. Owning a task means\n" +
				"  coordinating its progress, not being expected to solve every problem alone.\n" +
				"steps:\n" +
				"  - id: deliver-owned-task\n" +
				"    name: Take an agreed task through the team's process\n" +
				"    checks:\n" +
				"      - ask: Has the agreed task reached the team's completion conditions with required review and verified delivery evidence?\n",
		},
		{
			summary: contract.Summary{
				ID:          "ticket-pairing",
				Title:       "Pair on a ticket",
				Category:    "software-work-and-delivery",
				Goal:        "First week",
				Result:      "Work through a real task with a teammate",
				Carrier:     "section-template",
				ContentPath: "sections/ticket-pairing.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: ticket-pairing\n" +
				"title: Pair on a ticket\n" +
				"summary: Work through a real task with a teammate\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: pairing_practice, description: \"How should a first pairing session be arranged and what collaboration practices apply?\", audience: onboarder, required: true}\n" +
				"  - {name: pairing_goal, description: \"What learning outcome and suitable task characteristics should guide the session?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Arrange the session using $pairing_practice.\n" +
				"  Use the agreed learning focus: $pairing_goal\n" +
				"\n" +
				"  Confirm the assigned partner and real ticket in preparation; an unassigned\n" +
				"  partner stays missing rather than becoming a shared default. Agree scope,\n" +
				"  roles and permission boundaries together. Work through a useful slice,\n" +
				"  inspect its tests or evidence and discuss what happens next. Pairing does\n" +
				"  not bypass review or authorize a production change.\n" +
				"steps:\n" +
				"  - id: complete-pairing-session\n" +
				"    name: Work through a real task with the assigned partner\n" +
				"    checks:\n" +
				"      - ask: Have you worked through a real task together and discussed the result, learning and next steps?\n",
		},
		{
			summary: contract.Summary{
				ID:          "ticket-system-access",
				Title:       "Ticket system access",
				Category:    "tools-repositories-and-documentation",
				Goal:        "Before ticket work",
				Result:      "Read and update the relevant project or board as needed",
				Carrier:     "section-template",
				ContentPath: "sections/ticket-system-access.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: ticket-system-access\n" +
				"title: Ticket system access\n" +
				"summary: Read and update the relevant project or board as needed\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: ticket_scope, description: \"Which ticket service, project or board and permitted operations are required?\", audience: onboarder, required: true}\n" +
				"  - {name: ticket_access_route, description: \"How is ticket access requested and safely verified, and who handles problems?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Locate the intended service and project: $ticket_scope\n" +
				"  Obtain and verify access using $ticket_access_route.\n" +
				"\n" +
				"  Open a relevant issue and inspect its status and collaboration context.\n" +
				"  If your role needs updates, use an explicitly permitted non-disruptive\n" +
				"  action to confirm that capability. Do not alter a real ticket merely\n" +
				"  as a test without its owner's agreement. Ask about restricted material\n" +
				"  rather than requesting access to every project.\n" +
				"steps:\n" +
				"  - id: verify-ticket-capability\n" +
				"    name: Verify the needed ticket-system operations\n" +
				"    checks:\n" +
				"      - ask: Can you read the relevant board and perform the approved updates your role requires?\n",
		},
		{
			summary: contract.Summary{
				ID:          "tool-feedback",
				Title:       "Provide feedback on onboard-kit",
				Category:    "feedback-and-progress",
				Goal:        "After initial use; revisit around day 30",
				Result:      "Send tool friction and suggestions through the agreed route without secrets or unnecessary personal details",
				Carrier:     "section-template",
				ContentPath: "sections/tool-feedback.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: tool-feedback\n" +
				"title: Provide feedback on onboard-kit\n" +
				"summary: Send tool friction and suggestions through the agreed route without secrets or unnecessary personal details\n" +
				"timing: anytime\n" +
				"variables:\n" +
				"  - {name: tool_feedback_route, description: \"Which agreed internal route receives onboard-kit feedback?\", audience: onboarder, required: true}\n" +
				"  - {name: tool_feedback_context, description: \"What nonsecret context and review process make a feedback report useful?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Send useful feedback through $tool_feedback_route.\n" +
				"  Include only the agreed nonsecret context: $tool_feedback_context\n" +
				"\n" +
				"  Describe the task you were attempting, the friction and the useful\n" +
				"  result you expected. Review any example before sharing it; do not attach\n" +
				"  a real plan, preparation, record or invitation without approved handling.\n" +
				"  Separate tool behavior from organization-specific instructions. Revisit\n" +
				"  feedback around day thirty if helpful, but do not block useful work on\n" +
				"  submitting a report or assume a public issue tracker is the right route.\n" +
				"steps:\n" +
				"  - id: provide-safe-feedback\n" +
				"    name: Identify or use the agreed tool-feedback route\n" +
				"    checks:\n" +
				"      - ask: Can you provide actionable tool feedback through the agreed route without secrets or unnecessary personal details?\n",
		},
		{
			summary: contract.Summary{
				ID:          "video-arrange",
				Title:       "Arrange meeting service access",
				Category:    "onboarder-preparation",
				Goal:        "Before meeting participation",
				Result:      "Arrange configured accounts or guest access",
				Carrier:     "preparation-guidance",
				ContentPath: "preparation/video-arrange.md",
			},
			content: "# Arrange meeting service access\n" +
				"\n" +
				"Identify each configured meeting service and whether an approved account,\n" +
				"client or guest route is appropriate. Keep shared joining and privacy\n" +
				"instructions in the plan. Arrange a safe test without assuming a provider,\n" +
				"paid licence, recording permission or already-working company account.\n" +
				"\n" +
				"Use the existing access kind for a person-specific service arrangement linked\n" +
				"to an actual declared meeting-access section instance. An arranged item needs\n" +
				"a known accountable owner and nonsecret reference; known helper contacts are\n" +
				"optional for this kind. Keep missing details visible under pending, requested\n" +
				"or deferred instead of claiming readiness. Unknown supplied references remain\n" +
				"invalid. Do not add a video-arrange kind or save secret joining tokens.\n" +
				"\n" +
				"Confirm a support route for guest restrictions, audio or collaboration problems.\n" +
				"Any target date is a planning goal. Arranged access is not proof the newcomer\n" +
				"can join with required features, and a sent meeting invitation proves neither\n" +
				"attendance nor usability. Preparation is onboarder-owned and changes no\n" +
				"newcomer record or shared organization answers.\n",
		},
		{
			summary: contract.Summary{
				ID:          "video-meetings",
				Title:       "Video meetings",
				Category:    "identity-equipment-and-communication",
				Goal:        "Day one",
				Result:      "Join with required audio and collaboration features",
				Carrier:     "section-template",
				ContentPath: "sections/video-meetings.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: video-meetings\n" +
				"title: Video meetings\n" +
				"summary: Join with required audio and collaboration features\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: video_setup, description: \"Which configured meeting services, approved clients or guest joining instructions apply?\", audience: onboarder, required: true}\n" +
				"  - {name: meeting_test, description: \"How can a newcomer safely test required meeting features and get joining help?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the service-specific instructions: $video_setup\n" +
				"  Use the approved test and support route: $meeting_test\n" +
				"\n" +
				"  Test joining, sound and the collaboration features your role needs.\n" +
				"  Check privacy and recording expectations before sharing a screen or\n" +
				"  recording. Do not install an unapproved client or assume a paid account\n" +
				"  is required when the organization provides guest access. Report any\n" +
				"  accessibility or connection barrier through the agreed route.\n" +
				"steps:\n" +
				"  - id: test-meeting\n" +
				"    name: Join a safe test meeting\n" +
				"    checks:\n" +
				"      - ask: Can you join with usable audio and the required collaboration features under the team's privacy rules?\n",
		},
		{
			summary: contract.Summary{
				ID:          "week-one-check-in",
				Title:       "End-of-week onboarding check-in",
				Category:    "feedback-and-progress",
				Goal:        "End of week one",
				Result:      "Discuss blockers, surprises, and next priorities",
				Carrier:     "section-template",
				ContentPath: "sections/week-one-check-in.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: week-one-check-in\n" +
				"title: End-of-week onboarding check-in\n" +
				"summary: Discuss blockers, surprises, and next priorities\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: checkin_route, description: \"How is the first-week check-in arranged with the onboarding owner?\", audience: onboarder, required: true}\n" +
				"  - {name: checkin_focus, description: \"What shared first-week goals and support questions should guide the conversation?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Arrange the conversation through $checkin_route.\n" +
				"  Use the agreed focus: $checkin_focus\n" +
				"\n" +
				"  Discuss what helped, what surprised you and what still blocks useful\n" +
				"  work. Separate unavailable organizational support from personal learning\n" +
				"  needs. Agree the next priorities, owners and a follow-up route. The end\n" +
				"  of week one is a planning prompt, not a failure deadline; reschedule when\n" +
				"  needed and preserve the value of the conversation over a completion count.\n" +
				"steps:\n" +
				"  - id: review-first-week\n" +
				"    name: Discuss the first week and agree next priorities\n" +
				"    checks:\n" +
				"      - ask: Have you discussed blockers and surprises with the owner and agreed supported next priorities?\n",
		},
		{
			summary: contract.Summary{
				ID:          "working-agreements",
				Title:       "Review working agreements",
				Category:    "team-and-coordination",
				Goal:        "First week",
				Result:      "Understand communication, decisions, review, availability, and feedback",
				Carrier:     "section-template",
				ContentPath: "sections/working-agreements.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: working-agreements\n" +
				"title: Review working agreements\n" +
				"summary: Understand communication, decisions, review, availability, and feedback\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: agreement_material, description: \"Where are current working agreements, or how will the team establish them if absent?\", audience: onboarder, required: true}\n" +
				"  - {name: agreement_discussion, description: \"Who can explain agreements, resolve uncertainty and discuss improvements?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Review or establish the agreements using $agreement_material.\n" +
				"  Discuss questions through $agreement_discussion.\n" +
				"\n" +
				"  Cover communication, decision ownership, review, availability and\n" +
				"  feedback. Ask for examples of how the agreements work in practice.\n" +
				"  Distinguish written policy from a preference or an assumption. If the\n" +
				"  team has no agreements, work with the people involved to establish them;\n" +
				"  do not report an imaginary policy as reviewed. Capture agreed changes\n" +
				"  through the team's normal process.\n" +
				"steps:\n" +
				"  - id: discuss-agreements\n" +
				"    name: Understand or establish practical working agreements\n" +
				"    checks:\n" +
				"      - ask: Can you describe the agreed collaboration practices and how an unclear or missing agreement will be resolved?\n",
		},
		{
			summary: contract.Summary{
				ID:          "working-device",
				Title:       "Working device",
				Category:    "identity-equipment-and-communication",
				Goal:        "Day one",
				Result:      "An approved device is available and usable",
				Carrier:     "section-template",
				ContentPath: "sections/working-device.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: working-device\n" +
				"title: Working device\n" +
				"summary: An approved device is available and usable\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: device_process, description: \"What approved device setup, management and delivery instructions apply?\", audience: onboarder, required: true}\n" +
				"  - {name: device_support, description: \"What support route handles missing equipment or a device that will not work?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Follow the approved equipment instructions: $device_process\n" +
				"  For missing equipment or setup problems, contact $device_support.\n" +
				"\n" +
				"  Confirm the device is intended for this work before enrolling or changing\n" +
				"  it. Check power, required peripherals, connectivity and an ordinary work\n" +
				"  session. Do not bypass device management or use unapproved personal\n" +
				"  equipment to make progress appear complete. Tell the owner which work is\n" +
				"  blocked if delivery or support is still pending.\n" +
				"steps:\n" +
				"  - id: verify-device\n" +
				"    name: Use the approved working device\n" +
				"    checks:\n" +
				"      - ask: Can you use the approved device and necessary peripherals for an ordinary work session?\n",
		},
		{
			summary: contract.Summary{
				ID:          "working-hours",
				Title:       "Working hours and availability",
				Category:    "identity-equipment-and-communication",
				Goal:        "Day one",
				Result:      "Agree and configure hours, time zone overlap, and availability signals",
				Carrier:     "section-template",
				ContentPath: "sections/working-hours.yaml",
			},
			content: "schemaVersion: 1\n" +
				"id: working-hours\n" +
				"title: Working hours and availability\n" +
				"summary: Agree and configure hours, time zone overlap, and availability signals\n" +
				"timing: week-one\n" +
				"variables:\n" +
				"  - {name: availability_agreement, description: \"What shared overlap expectations and process for agreeing individual hours apply?\", audience: onboarder, required: true}\n" +
				"  - {name: availability_tools, description: \"Which calendar or status settings communicate availability and whom should conflicts be discussed with?\", audience: onboarder, required: true}\n" +
				"body: |\n" +
				"  Discuss the shared working expectations: $availability_agreement\n" +
				"  Configure the agreed availability signals using $availability_tools.\n" +
				"\n" +
				"  Confirm time zones and sustainable overlap with the people involved.\n" +
				"  Individual arrangements are a discussion, not an assumed preference or\n" +
				"  a demand for personal location. Make focus time and planned absence\n" +
				"  understandable without disclosing private reasons. Revisit the agreement\n" +
				"  when it creates avoidable coordination problems.\n" +
				"steps:\n" +
				"  - id: agree-availability\n" +
				"    name: Agree and communicate working availability\n" +
				"    checks:\n" +
				"      - ask: Have the relevant people agreed working overlap and can they interpret your availability signals?\n",
		},
	}
}
