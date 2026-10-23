# 0280. First steps help and sources

Status: accepted by agent judgment; implementation pending

Builds on [0090](0090-plan-and-onboarding-record.md),
[0140](0140-person-specific-preparation.md),
[0180](0180-first-onboarding-use.md),
[0240](0240-invitation-archive-envelope.md), and
[0260](0260-prerequisites-and-action-order.md).
Reconciles the [first-steps reference](../reference/first-steps-document.md)
with the staged runtime and the three separately owned artifacts.

## Context

A first-steps guide needs useful setup, access, recovery and support information.
The reference asks for a download address and invitation/start/update commands,
but no supported distribution address or complete invitation runtime exists.
[0250](0250-bootstrap-content-and-projection.md) forbids inventing those details.
The guide must not promise a capability merely because its renderer works.

The plan directory contains possible contacts, not a declaration that every
contact should appear in an invitation. Preparation identifies an onboarding
owner. The standard support template declares general_contacts, but custom plans
have no designated section binding for discovering that question automatically.

## Decision

### Preserve useful help without inventing capabilities

Keep all twelve help topics. Use truthful owner-guided wording until supported
distribution and runtime instructions are independently implemented and checked.
Label missing information rather than guessing a URL, contact, repository,
command or completed activity. Successful rendering is a preview, not invitation
readiness, identity confirmation, an import, or a first-use event.

A separate pure first steps assembler wraps the exact fragment returned by the
real document assembler's AssembleBootstrap operation on the complete original
plan. It does not copy the renderer, slice an ordinary document, prefilter the
plan, inject preparation as accepted answers, or erase bootstrap diagnostics.
Structural refusal yields no partial guide. Empty pre-access content gets a
plain explanation asking the onboarding owner what is needed.

### Use explicit sources

The caller supplies the explicit account and an optional ordered list of typed
contact/role help references selected by the onboarder. Preparation supplies the
onboarding owner; the plan directory supplies contact routes and role holders.
The preparation team may select a help-role holder override without changing
the original plan's answer or bootstrap resolution context. Empty overrides stay
vacant. No first-contact, all-directory, title, unknown-field or default inference.

This is a transient input projection, not a new saved field or an automatic
discovery convention. A future driver must expose explicit selection or show
the missing-help explanation. Readiness requirements and a persistent binding
need their own contract before the invitation can be sent.

Display the four fixed archive names separately from locator-derived record and
preparation destinations. Those destinations are relative to a root chosen by
the later import flow, not extraction instructions or filesystem authorization.

### Keep display and personal information bounded

The personalized guide may show its explicit account, account paths, selected
onboarding owner and selected business help routes. These values help the
recipient recognize the invitation, understand its destinations, and obtain
support. This narrowly replaces the reference's plan-only personal-data rule.
Do not copy recipient name or email, planned dates, arrangements, notes or unknown
preparation fields into standard help. Directory business contact information
remains visible; file ownership is not privacy.

Treat wrapper source values as literal display data: escape Markdown punctuation,
flatten line separators, and fence account/path text safely. Keep the actual
bootstrap fragment unchanged. Missing names and contact routes are labeled;
vacant roles remain visible. Rendering neither executes content nor certifies
its safety, authenticity, readiness or access authorization.

## Rationale

Principles P2 and P4 of [0100](0100-behavior-judgment.md) preserve ownership and
existing work. P3 and P5 favor useful pre-access help with honest limitations.
P7 limits the personalized values and prevents accidental display expansion.
Hard-coding a likely distribution URL or scanning the directory for a plausible
owner would turn a guess into instructions. Dropping help entirely would leave
the newcomer without a recovery route.

## Component boundary and verification

First steps assembler is dependency position 5. It may depend on document
assembler at 4, preparation validator and locator at 1, model, and its own
contract. It reads the original plan context required by bootstrap rendering,
explicit help/account input, validated preparation identity and owner/team, and
display metadata. It writes no input or outside state and reads no record,
clock, filesystem, environment, process or network.

Independent literal goldens must cover standard help, missing sources, identity
and path refusal, fixed archive names versus custom destinations, original-plan
bootstrap selection, gaps, action order, selected help and vacancy, minimal PII,
and literal display escaping. Wrong adapters must distinguish a second renderer,
guessed contacts, dropped diagnostics and unsupported commands. Serialization,
readiness, archive writing, confirmed startup and release delivery remain
separate verification boundaries.
