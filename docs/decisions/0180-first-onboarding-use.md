---
category: behavior
status: accepted
approval: unrecorded
affects: [firststeps/assembler, outcome/recorder, preparation/decoder, preparation/editor, preparation/encoder, preparation/reviewer, preparation/validator, section/walker]
builds-on: ['0150', '0160']
partly-supersedes: ['0170']
---

# 0180. First onboarding use, not actual start

Supersedes [0170](0170-planned-and-actual-start-dates.md)'s actual-start-date
question. Keeps its onboarder-owned plannedStartDate in preparation. Builds on
[0150](0150-onboarding-library-and-preparation-flow.md) and
[0160](0160-personal-data-and-preparation-approval.md).

## Context

An actual-start-date question adds work for the newcomer. For understanding when
someone first used onboarding, the system can record the event it observes
instead. Opening the tool does not establish when someone started their job.
The person approved dropping the question in favor of a first-use timestamp.

## Why this choice

We first considered one start date suggested by the onboarder and accepted or
changed by the onboardee. That makes the two people's ownership unclear and
loses the distinction between the expectation and what happened. We then
accepted a planned date plus an explicitly answered actual date in 0170.
That could support a comparison, but it adds a question the newcomer must
understand, answer, and maintain without a demonstrated onboarding need.

The narrower useful fact is when the person first begins using this tool.
The system can record that event without asking them to restate it. This
reduces avoidable newcomer work under principle P1 and keeps the next useful
action clear under P3. The existing identity confirmation defines the event;
we do not add another question or treat an unaccepted default as an answer.

We keep plannedStartDate because the organization needs to prepare equipment,
access, and support ahead of an expected start, and that plan can change.
We name the observed fact firstOpenedAt because someone can begin work before
opening the tool, or open onboarding before their first working day. Calling
it actual start would create a misleading fact and misleading future metrics.

The two fields therefore serve different purposes: organizational planning
and observed tool engagement. Neither replaces the other. No current feature
needs the true employment-start date, so we do not collect it speculatively.
If a concrete need appears later, it requires an explicit product and privacy
decision rather than relabeling first use.

First use may be the more useful onboarding signal in its own right: it shows
engagement with the invitation and tool rather than duplicating an employment
fact the company already maintains. If a real start differed from the plan and
that distinction matters, the manager or another authorized company contact can
check the company's source. Do not make every newcomer answer another question
to cover that occasional investigation.

The expectation that planned and actual starts usually agree is an assumption,
not a verified invariant. This decision does not depend on their agreement:
first use remains first use even when employment dates differ. A future metric
can help improve invitation timing and support, but must not infer attendance,
individual performance, or blame from a tool-opening timestamp.

This is the close-enough principle, recorded as P8 in
[0100](0100-behavior-judgment.md): choose sufficient accuracy with less user
effort, not speculative precision. The first-use timestamp is precise about
the event observed and only a possible proxy for engagement around the planned
start. It is not an approximate employment-start answer. If a future decision
truly depends on actual employment start, first use is not sufficient for it.

## Decision

- Preparation holds the onboarder's optional plannedStartDate, a calendar date.
- The newcomer's record will hold an optional firstOpenedAt, a UTC RFC 3339
  timestamp supplied by the driver. The core never reads a clock.
- Record it once when the onboardee confirms the invitation identity and begins
  onboarding. Do not record it for help commands, archive inspection, the
  onboarder preparing a package, or an unconfirmed or rejected identity.
- Preserve the existing timestamp on later visits, retries, imports, and changes
  to preparation. Opening another person's record must not record their first use.
- There is no actual-start-date question and no inferred actualStartDate.

This timestamp is an observed tool event, not an accepted default answer, an
employment start date, a readiness judgment, or completion of an activity.
The flow explains that it records first use in the local onboarding record;
it introduces no background transmission or separate telemetry service.
Sharing that file follows the existing visibility rules.

Changing the planned date does not change firstOpenedAt. A difference between
the planned date and the calendar date of first use does not fail onboarding.
Do not manufacture a first-use timestamp from an existing startDate, file
modification time, Git history, plan generation date, or invitation creation.
For an old record without a recorded event, earlier first use is unknown.
Define legacy-record handling before that flow is changed; do not present an
upgrade-time timestamp as proven historical first use.

## Reporting and privacy

Both fields are personal data under 0160. Keep the names honest in displays and
future metrics. Planned date versus first use measures tool engagement relative
to a plan, not the delay before someone actually began work. Client clock
timestamps are not authoritative evidence of attendance or employment.

Future reports must declare which date or event anchors each target and handle
unknowns without a silent fallback. Do not add metrics or original-plan history
as part of this change. Targets remain planning goals, not enforcement.

## Verification and delivery

- Confirmed first onboarding use records the injected timestamp once.
- Rejected or unconfirmed identity and non-onboarding commands write no event.
- Revisits and retries with another supplied time preserve the original value.
- Editing plannedStartDate changes no existing firstOpenedAt or newcomer outcome.
- Existing startDate alone does not create firstOpenedAt or an actual-start answer.
- The onboarding question list contains no actual-start-date question.

These are requirements, not verified runtime behavior. The preparation validator
checks planned date validity only. Record serialization, the first-use event,
and driver writes need separate approved contract and binary acceptance
increments under [0025](0025-contracts-and-tests.md).
