---
category: behavior
status: accepted
approval: unrecorded
affects: []
builds-on: ['0010', '0050', '0055', '0060']
---

# 0065. Messages, clipboard, and evidence

Builds on [0010](0010-plan-editor.md), [0050](0050-section-switch.md), [0055](0055-variable-audience.md), and [0060](0060-personal-plan.md). It prepares the action runner and the onboardee runtime.

## Context

A step today can run a command, open a URL or file, and carry checks. Many onboarding steps are really "tell someone": ask IT for access, post in a team channel, email a manager. A template can only describe those in prose, and the tool cannot help the newcomer do them or know that they were done.

Two things came out of discussing the access-request step.

- A section whose steps depend on different facts is an invalid structure. If one step in a section applies only for some people, it belongs in its own section. See Decision 1.
- Which way a message goes varies by organization and by person. Some organizations do not allow bots to post. Some people do not want a tool to post for them.

## Decision

### 1. A section is the unit of "only if"

Every step in a section applies together. A step that applies only sometimes goes in its own small section, with an `includeIf` on a plan variable or a switch the onboarder can turn off. There are no step-level conditions. Text that differs inside a step is a free-text variable the organization fills in. The schema docs say so.

### 2. A step can carry a message

A step may carry a `message`, next to `run` and `open`. A message has:

| Field | Meaning |
| -- | -- |
| `via` | `slack`, `teams`, `email`, or `other` |
| `to` | A `contact` or `role` variable, a channel or address, or text for the person to read |
| `text` | The message, a template that uses `$variables` |
| `subject` | For email only |

Structure is here because the tool acts on it: it copies the text, opens the right place, and in some modes sends it. A step may still carry free text in `notes` for a channel the tool cannot reach.

### 3. Two modes of sending

| Mode | What happens | Needs |
| -- | -- | -- |
| `copy` | The tool puts the text on the clipboard and opens the channel or a link. The person posts it | Nothing. It always works |
| `user` | The tool sends as the person's own account | The person authorizes the tool once |

There is no mode where the organization's own bot sends. See Decision 7.

Three parties have a say, in this order:

- **The onboarder** states which modes the organization allows. An organization that does not want the tool sending anything lists `copy` only. The onboarder may also narrow one message.
- **The onboardee** chooses among the allowed modes and may switch at any time, including choosing `copy` for everything.
- **The tool** states the mode it will use on every send, and offers the switch there.

The data for this is ordinary plan data with audiences from 0055, not a new kind of record. The exact shape is not decided, see below. The shape must exist from the first version even though only `copy` ships first.

### 4. Every send is previewed and confirmed

The tool shows the mode, the recipient, and the exact text, and sends only after an explicit yes. There is no silent or batch send. In `copy` mode it shows what it copied and where it opened.

### 5. A message step closes with evidence

A check can ask for evidence, such as a link, and the step is done only when valid evidence is given. For a Slack message the person pastes the message's link, and the tool checks that it looks like a Slack message link. In `user` mode the tool records the link the service returns, so the person pastes nothing. Evidence is recorded with the step's progress. A person can mark a step skipped with a reason, and the skip is recorded too.

### 6. Clipboard and links are general

Copying text and opening a link are not only for messages. Any step that gives the person text to use can offer them. The tool uses the platform's own program: `pbcopy` on macOS, `wl-copy` on Wayland, `xclip` or `xsel` on X11, and `clip` on Windows. When none exists it shows the text for the person to copy. Starting those programs is a command the tool runs, so it is confirmed like any other, see 0060.

### 7. Credentials never go in a plan, and there is no bot mode

A token stays in the operating system's keychain or the person's own config, set up by the driver. The secrets check keeps rejecting tokens in plans.

An organization's bot token cannot be on every newcomer's machine. It is a shared credential, and anyone with the machine could extract it and post as the bot. A relay the organization runs would avoid that, but it means someone hosts a service for the tool, which is out of scope. So the tool never sends as a bot. If an organization later runs such a relay, that is a new decision.

### 8. Progress is its own record

Step status, evidence, and times are state of the personal plan, kept apart from answers, as 0060 anticipated. The format is not decided here.

### Stages

1. **First:** the `message` field and its data structure, `copy` mode for Slack, Teams, and email, the evidence check, the progress record, and the walkthrough of steps for the onboardee.
2. **Then:** `user` mode with a one-time authorization.

## Why

- Structure only where the tool acts on it, and prose everywhere else.
- `copy` works for every organization and every person, so the first version never blocks anyone. It needs no credential, so nothing secret reaches a newcomer's machine.
- The policy chain matches who owns what: the onboarder owns what the organization allows, and the onboardee owns their own choice.
- Evidence turns "I did it" into something the plan can show.

## Consequences

- **The domain model changes.** A Step gains `message`, a Check gains evidence, and the personal plan gains a progress record. This needs `slow-change`.
- **Contracts**, each its own small increment: the decoder, encoder, loader, and hasher (a message is step content), the resolver (it fills `$variables` in `to` and `text`), the validator (evidence patterns, and a person variable in `to`), and a new component for progress.
- **The walkthrough** has to walk steps, not only questions. Today it asks questions only, so the onboardee runtime (#28) comes first or alongside.
- **Docs.** The schema docs gain the message, the evidence check, and the section rule.

## Not decided

- The shape of the policy data: a few plan variables, or a small settings record.
- The progress record's format and where it sits in the personal plan.
- Which evidence patterns ship for Teams and email, which give no link.
- Whether a message can name more than one recipient.
