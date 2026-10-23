# 0100. Behavior judgment

Status: accepted for experimental use

[0130](0130-survivor-risk-acceptance.md) makes a scoped exception for agent-only
acceptance of the remaining mutation-survivor batch.

Builds on [0025](0025-contracts-and-tests.md),
[0045](0045-defaults-are-suggestions.md),
[0055](0055-variable-audience.md), and
[0090](0090-plan-and-onboarding-record.md).

## Context

Mutation survivors and contract reviews expose choices about product behavior.
Asking a person to settle each one slows work, but letting an agent follow the
current code makes an implementation accident into a requirement. We need
explicit product principles and a bounded way for an agent to apply them.

The behavior judge is a development workflow. It adds no model call or runtime
dependency to the shipped tool. Its decisions become deterministic tests, not
model judgments during onboarding or assembly.

## Product principles

The goal is to help an organization prepare a newcomer to do useful work, with
as little avoidable effort and uncertainty for the newcomer as possible.

### P1. Prepare what the onboarder reliably knows

The onboarder fills in what they can reliably supply for an invitation. The
onboardee should not rediscover organizational facts or re-enter information
already supplied. An onboarder who knows the person's team may supply it, as
0055 allows.

This does not permit guessing personal facts, exposing private information,
making preferences on someone's behalf, or claiming that a step was completed.
A person supplying a known answer is not the tool accepting a template default.

### P2. Respect ownership and deliberate choices

Keep organizational facts under onboarder control. Respect saved personal
answers and intentional deferrals. A default is a suggestion, not an answer;
only a person's explicit acceptance makes it an answer. Never silently use a
default, for any variable type, audience, or scope.
The ownership and audience rules of 0055 and 0090 still govern where answers
are stored and who may change them.

[0120](0120-defaults-require-acceptance.md) makes the suggestion rule explicit
for every variable type, audience, and scope, including directory references.

### P3. Make the next useful action clear

Ask relevant questions. When work cannot proceed, explain what is missing and
who can resolve it. Prefer successful onboarding to fewer clicks or a screen
that merely looks complete. A diagnostic should help a person understand the
problem, not merely contain some text.

### P4. Preserve work and keep changes local

Do not discard answers or rewrite unrelated data as a side effect of another
action. Prefer explicit, recoverable changes. When two options serve the person
equally well, prefer fewer side effects. This does not impose one absent/empty
collection rule on components whose contracts make different promises.

### P5. Keep first steps usable before access exists

Onboarding must not require the access it is helping someone obtain. Preserve
the invitation's bootstrap path, the standalone offline-capable tool, and
deterministic assembly. See [0005](0005-self-contained-install.md) and 0090.

### P6. Use dates to plan support, not punish progress

Dates are goals. A passed target must not fail an activity, block the newcomer,
or imply blame. Identify missing preparation, who can supply it, and whom it
blocks. Preserve real prerequisites and authorization independently of dates.
See [0150](0150-onboarding-library-and-preparation-flow.md).

### P7. Treat personal data as PII

Treat information about an identifiable person as PII for product design,
including company email, account ids, and linked dates, assignments, and progress.
Collect and retain only what serves a stated onboarding purpose, and explain
storage and readership. File ownership does not provide privacy. Exclude personal
email accounts, phone numbers, home addresses, and precise location from the
normal flow unless an explicit exception is justified. Prefer time zone to
location. Needed business email may be visible in shared files; do not imply
it is hidden without an implemented control. Avoid unnecessary personal values
in diagnostics or logs. See [0160](0160-personal-data-and-preparation-approval.md),
including the recorded option for field-level encryption in a later version.

### P8. Prefer sufficient accuracy with less user effort

Collect information precise enough for the real decision, not the most precise
information imaginable. When an observed event or clearly labeled proxy serves
the purpose, prefer it to another question or manual entry. Make the limitation
visible and ask for the exact fact when the difference would change a meaningful
decision. Do not collect extra personal data for hypothetical future metrics.

Close enough is not permission to guess an answer, silently accept a default,
hide ambiguity, claim completion, or relax safety or authorization. Keep a
proxy's honest name rather than presenting it as a fact it does not establish.
[0180](0180-first-onboarding-use.md) applies this to first use versus actual start.

## Decision

Use the [behavior judge skill](../../.agents/skills/behavior-judge/SKILL.md) for
open product-behavior questions. The principles live here; the skill carries
the procedure. Neither an implementation nor a mutation score is product
authority.

These principles are not points in a score. Privacy, correctness, ownership,
and explicit commitments constrain convenience. Within those constraints,
reduce avoidable newcomer work and uncertainty, then prefer a smaller and more
reversible change. A general principle does not silently supersede a specific
accepted decision or resolve a contract conflict.

### Delegated product judgment

An agent may choose a new behavior without another product discussion when:

- It stays within the person's assigned task.
- These principles clearly favor one option without an unresolved tradeoff.
- It is local and practically reversible, with no new privacy, security,
  ownership, destructive, stored-format, or hash-compatibility consequences.
- It does not contradict an accepted decision, contract, or explicit direction.
- It has an observable acceptance criterion and a valid independent test oracle.

This authority selects a product direction. It does not grant permission to
edit protected artifacts, approve tests, or weaken a requirement. A recorded
choice must distinguish product judgment from evidence and implementation.

### Escalation

Ask the person when principles conflict, required facts are missing, a valid
test cannot be established, the consequences exceed the delegation, or an
existing commitment must change. Include a concrete example, recommendation,
alternatives, and the effect on the people using the tool.

Do not escalate arithmetic or library claims as product choices. Investigate
them as engineering questions. Conversely, a plausible test does not justify
inventing a preferred behavior without product authority.

### Evidence and approval

Record the decision, its governing principles, assumptions, observable scenario,
independent expected result, and a wrong behavior the proposed test rejects.
Use a component contract test or a whole-binary acceptance scenario according
to the boundary being judged. Separate these states: decided, test planned,
test verified, implemented.

The human approval and independent authorship rules in 0025 remain unchanged.
A person still approves protected contract and test changes. A person still
adds accepted-survivor entries after checking the evidence. A skill does not
approve itself, promote its own authority, or mark a trial judgment as human
approval. Contract conflicts are reported, not repaired without authorization.

## Examples

- **Known team:** the invitation supplies a valid team answer. Do not require
  the newcomer to enter it again. A template default alone is not such an
  answer. This follows P1, P2, and 0055.
- **Allocation estimate:** changing a map's initial capacity without changing
  returned data is an engineering equivalence question, not a product choice.
  Apply and reach the mutation before recommending an exception.
- **Diagnostics:** punctuation need not be fixed. Removing words so that a
  message no longer explains found versus expected values is a candidate
  semantic test gap under P3, not automatic permission to tighten the contract.
- **Audience ambiguity:** duplicate declarations with different audiences can
  change answer ownership. Do not invent a winner under P1; report the existing
  rule and escalate a proposed policy change under P2.

## Consequences

The judge can settle bounded behavior choices and present routine decisions in
batches. It must not merely repeat the current code or find a principle that
sounds like an excuse for each survivor. Working review records stay in
`output/`; decisions that contracts or acceptance tests depend on belong in
ADRs under the existing repository rules.

This experiment does not approve outstanding survivor exceptions, resolve
known contract conflicts, or change the staged rollout of earlier decisions.
Evaluate the judge on concrete outcomes and escalation behavior before
expanding its authority. A passing skill-format check is not proof that its
product judgments are sound.
