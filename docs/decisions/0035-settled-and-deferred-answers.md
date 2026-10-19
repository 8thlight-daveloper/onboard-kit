# 0035. Settled and deferred answers

Status: proposed

Relates to [#91](https://github.com/8thlight-daveloper/onboard-kit/issues/91), [#92](https://github.com/8thlight-daveloper/onboard-kit/issues/92). Builds on [0010](0010-plan-editor.md) and [0030](0030-self-contained-plan-file.md).

## Context

A question can be answered, left alone, or put off. Three phrases grew up around that, and none of them says what it does.

- The conditions reference defines a "real answer" as a stored answer that is not `unknown-yet` and not empty. A condition tests only real answers.
- 0010 calls `unknown-yet` "an accepted answer". It satisfies a required question.
- The condition operator `isAnswered` tests for a real answer, so `isAnswered: true` is false for `unknown-yet`. A reader who sees `unknown-yet` stored for a question would call it answered.

So the same stored value counts as an answer for one rule and not for the other, and the names do not tell the two apart. Reviewing the condition evaluator contract showed the problem: a method named `isAnswered` could mean either rule. "Real" says nothing about what it excludes. The model's value kind, Unknown, and the file text `unknown-yet` name what the system lacks, where the person actually made a choice: to answer later.

## Decision

An answer is in one of three states.

| State | What is stored | Example |
| -- | -- | -- |
| Unanswered | Nothing, empty text, or an empty list | no entry, `""`, `[]` |
| Deferred | The value `deferred`. The person chose to answer later | `tracker_grantor: deferred` |
| Settled | Any other value: text with at least one character, a flag, or a list with a non-empty item | `team_state: established`, `flag: false` |

A settled answer is not final. A person can change it. It means the question is decided for now.

### Rules

- **A condition tests settled answers only.** Every test except `isSettled` is false when the answer is deferred or unanswered, before any negation. A deferred answer does not start follow-up questions early. This is the rule the conditions reference already has, under a new name.
- **`isSettled` replaces `isAnswered`.** `isSettled: true` holds when the answer is settled. `isSettled: false` holds when it is deferred or unanswered.
- **A required question is satisfied by a settled or a deferred answer.** Unanswered blocks. There is no separate term for "settled or deferred".
- **A deferred answer becomes a gap** when the document is assembled, as `unknown-yet` does today.
- **The stored form is the plain word `deferred`.** The quoted `"deferred"` is the Text answer with that word, as the encoder already does for the current marker.

### Names

| Old | New |
| -- | -- |
| `unknown-yet`, in plan files, defaults, and the entry the TUI offers | `deferred` |
| The value kind Unknown, `ValueKindUnknown` | The value kind Deferred, `ValueKindDeferred` |
| The operator `isAnswered`, the field `IsAnswered` | `isSettled`, `IsSettled` |
| "Real answer" | Settled answer |
| "Accepted answer" | A settled or deferred answer |
| The JSON form `{"unknown": true}` in the shared test code | `{"deferred": true}` |

The kind is `Deferred` and not `DeferredAnswer`. A kind names the shape of a value, like Text, Flag, and List. A Value is used for defaults and tested values too, so a kind named for answers would claim that every Value is one.

### Versioning

This is a rename of a file token and an operator, which 0030 would call a breaking change. Nothing exists outside this repository: no released tool, no saved plan, and no template that a client holds. So the schema version stays 1. After the first release, a rename like this needs a version bump and a step in the plan migrator.

## Consequences

- The domain model changes. This ADR authorizes the two renames in `internal/model/`.
- The plan decoder and encoder contracts, the shared fixtures, and both built components change. The agent that updates the tests does not patch the code.
- The condition evaluator, section hasher, and section loader contracts take the new words. The evaluator build takes them too.
- The glossary, the schema reference, 0010, 0030, the architecture docs, and the ACME example take the new words.
- No operator tests for a deferred answer. One is added when a template needs it, together with `any` and `none`.
- Every place that said "unknown yet" in prose now says "deferred".

## Alternatives considered

- **Keep `unknown-yet` and the Unknown kind, and only document the two rules.** It names what the system lacks and not what the person chose. Two rules would still share one stored word.
- **Keep `isAnswered` and say in the docs that it means a settled answer.** The name would still read as "has any answer", and two readers would disagree about `unknown-yet`.
- **"Real", "final", or "acceptable" for the settled state.** "Real" says nothing. "Final" implies that the answer cannot change. "Acceptable" already means that an answer passes the plan validator's checks of type, options, and pattern.
- **`ValueKindDeferredAnswer`.** See the reasoning under Names.

## Open questions

- Does the TUI show the deferred entry as "Deferred" or in other words? The label belongs to the walkthrough issue, #6.
