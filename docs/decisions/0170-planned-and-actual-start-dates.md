# 0170. Planned and actual start dates

Status: superseded in part by [0180](0180-first-onboarding-use.md)

0180 removes the actual-start question and records first onboarding use instead.
The planned-date field remains. The two-date implementation below did not ship.

Builds on [0120](0120-defaults-require-acceptance.md),
[0140](0140-person-specific-preparation.md),
[0150](0150-onboarding-library-and-preparation-flow.md), and
[0160](0160-personal-data-and-preparation-approval.md).
Records the person's approval of R2's two-date design. Supersedes the single
start-date interpretation of 0090 and 0095 for the new invitation flow.

## Context

A planned start can change for many reasons. Treating the onboarder's planned
date as the newcomer's actual start would turn an organizational expectation
into an unconfirmed personal fact. Keeping the two facts distinct can also
support later process analysis without guessing what happened.

## Decision

The preparation recipient has an optional plannedStartDate, owned by the
onboarder. The onboarding record will have an optional actualStartDate, owned
by the onboardee. Both are calendar dates, not timestamps or activity outcomes.

The onboardee is asked when they actually started. Offer today as a suggestion;
they may explicitly accept it or enter another date. The driver supplies the
calendar date to the core, which never reads a clock. Opening, saving, importing,
or revisiting onboarding does not accept the suggestion. Today is not evidence
of when a person started, and a blank actual date stays unknown.

Do not silently substitute the planned date for the actual date or overwrite
one when the other changes. Editing an actual date is an explicit onboardee
action. A later change to the plan does not erase their answer. The two dates
may differ; that is useful information, not an ambiguity to resolve or failure
to punish. Date format validation is independent of completion deadlines.

Do not silently relabel an existing record.startDate as actualStartDate. Its
source may have been an invitation seed rather than a newcomer confirmation.
Keep existing data until approved serializer, compatibility, and driver changes
can preserve it and request explicit actual-date confirmation. Adding a model
type alone does not implement that transition.

## Reporting and privacy

Dates and any comparison tied to a person remain PII under 0160. This decision
does not authorize tracking, collecting explanations for a delay, or adding
metrics now. A future comparison must identify which planned date it used;
the latest edited plan is not necessarily the original invitation expectation.

Before changing report calculations, specify whether each target is anchored
to the planned or actual date and how an unknown date is displayed. Do not
silently use the other date as a fallback. Before-start organizational
preparation and later newcomer support can have different planning needs.
Targets remain goals under 0150 regardless of their anchor.

## Verification and delivery

- A missing actual date remains missing when today is offered but not accepted.
- Explicitly accepting today saves that supplied date. A different valid date
  saves the person's entered answer instead.
- A later visit with a different today keeps the saved actual date.
- Changing plannedStartDate leaves actualStartDate unchanged, and vice versa.
- A valid difference between planned and actual dates does not fail onboarding.
- Importing a legacy startDate does not manufacture confirmed actual start.

These are acceptance requirements, not verified runtime behavior. The preparation
validator checks only the supplied planned date's calendar validity; it cannot
prove actual-date confirmation, file ownership, or report calculations. Those
need separate approved contract and binary acceptance increments under
[0025](0025-contracts-and-tests.md).
