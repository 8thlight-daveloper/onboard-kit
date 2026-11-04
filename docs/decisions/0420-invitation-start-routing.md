---
category: behavior
status: accepted
approval: agent
affects: [command/start, invitation/starter]
builds-on: ['0100', '0110', '0360', '0400']
---

# 0420. Route explicit invitation starts without legacy fallback

Builds on [0100](0100-behavior-judgment.md),
[0110](0110-reject-ambiguous-plans.md),
[0360](0360-confirmed-invitation-start.md), and
[0400](0400-invitation-start-protocol.md).

## Context

The confirmed invitation starter and the retained personal-plan starter share
the start command. Selecting a driver must not reinterpret an option's value as
a request for another mode, hide mixed intent, or start legacy writes after an
invitation error. Both actual parsers remain responsible for their own grammar.

## Decision

Initial -h or --help invokes the invitation starter's exact help, ignoring later
arguments as its contract specifies. No runtime facts, input or files are read.
Otherwise inspect only option positions, splitting an equals option at its
first equals. Never search substrings inside an argument or filename.

Invitation selectors are --invitation and --resume-import, including their
equals forms. Do not add single-dash invitation aliases. Legacy selectors are
--org-plan, --no-walk and --refreshed-on, with their retained single-dash forms.
Shared root/home flags do not select legacy mode.

For this scan, invitation, resume-import, root, home, org-plan and refreshed-on
consume one separate value when not written with equals. Consume that position
even if its contents look like another option. The retained -home spelling also
consumes a value. No-walk is boolean and consumes no separate value. Unknown
options consume no guessed value. The standalone -- ends option recognition;
the selected actual parser decides whether its remaining arguments are valid.
Only spellings recognized by the relevant grammar receive this arity treatment.

Both invitation and legacy selectors mean ambiguity: refuse with the starter's
fixed start.usage JSON, exit2, empty stdout and no runtime/input/file access.
Do not choose the first selector, merge modes or retry with another driver.
Legacy selectors alone route unchanged argv to the existing legacy command.
Invitation selectors alone, or no selector, route unchanged argv to the actual
invitation starter. Empty argv, unknown flags and root/home alone therefore use
the starter's fixed usage behavior. A selected driver's parse/source failure
never authorizes fallback.

## Rationale and compatibility

P2 and P4 favor explicit modes and local effects over a convenience fallback.
P3 favors useful initial help without access; P7 avoids unnecessary personal
file/environment reads.0110 rejects ambiguous intent rather than choosing a
deterministic winner. This adds no answer default, model policy or stored format.

The retained binary accepts single-dash legacy spellings and option-looking
org-plan values, including --org-plan=--invitation and --org-plan followed by
--resume-import as a literal filename. Keep that observed interface; the scan
must not reinterpret those values. Do not infer new aliases from the language's
flag parser. The invitation grammar remains deliberately narrower.

## Verification and effects

Independent whole-binary scenarios pin initial help, no-selector usage, mixed
selectors before effects, selector-looking values, retained single-dash legacy
options and rejection of single-dash invitation aliases. Existing invitation
scenarios separately pin explicit identity/begin, cancellation, retained events
and private transaction outcomes. A first stored event needs a scalar assertion;
an unchanged missing field is not evidence of its creation.

The root router reads argv and delegates streams only. It performs no filesystem,
clock, environment, answer, outcome or record operation. Mixed-mode reporting
uses the fixed starter usage result, not a duplicate semantic driver. Production
implementation and executed scenario evidence remain separate from this decision.
