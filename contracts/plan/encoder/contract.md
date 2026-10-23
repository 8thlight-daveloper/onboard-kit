# Plan encoder contract

Status: approved. Component: `plan/encoder`. Issue: [#39](https://github.com/8thlight-daveloper/onboard-kit/issues/39).

The plan encoder turns a Plan into the bytes of a plan file. It is a pure function. It reads no file, writes no file, and does not change its argument. The same Plan always gives the same bytes. See [0030](../../../docs/decisions/0030-self-contained-plan-file.md), the [domain model](../../../docs/reference/domain-model.md), and [0015](../../../docs/decisions/0015-functional-core-and-imperative-shell.md).

## Interface

`Encode(plan) -> (bytes, problems)`.

- When every value in the plan can be written, `problems` is empty and `bytes` is the canonical file.
- When a value cannot be written, `bytes` is empty and `problems` lists each one. A problem has severity `error`.

## The canonical file

The same rules apply everywhere. The goal is that two people, or two versions of the tool, never write different bytes for the same plan, so plan files diff cleanly in version control.

### Document

- UTF-8, LF line endings, exactly one newline at the end, and no trailing spaces on any line.
- No tabs, no `---` or `...` markers, and no comments.
- Block style throughout, indented two spaces. A list under a key is indented two spaces, and each item starts with `- `.
- Never an anchor, alias, tag, or merge key.

### Keys

A key is written in the order the tables below give. A key with a zero value is omitted. Zero means an empty string, `false`, an empty or absent list or map, or a Value of kind none. Three keys are the exception, because the value `false` is a choice they hold. A condition test written `isSettled` is written whenever it is set, including `false`. A section's `enabled` is written whenever the section holds a value, `true` or `false`, and omitted only when the section holds none. A clause's `enabled` is written whenever the clause names a section, including `false`. A map entry, such as an answer, is always written, even when its value is `false`. An answers section with no answers is omitted. Inside an unknown field, an empty map is written `{}` and an empty list is written `[]`, because that data belongs to someone else and is kept as it is. Two exceptions: `schemaVersion` is always written, and `provenance` is omitted only when both of its fields are empty. `origin` is omitted when all four of its fields are empty, and each empty field of an `origin` is omitted on its own.

| Type | Key order |
| -- | -- |
| Plan | `schemaVersion`, `id`, `provenance`, `origin`, `planUrl`, `resultsPath`, `directory`, `variables`, `answers`, `sections`, then the unknown fields |
| Provenance | `templateSet`, `generatedOn` |
| Origin | `planId`, `source`, `contentHash`, `refreshedOn` |
| Directory | `contacts`, `roles`, `teams` |
| Contact | `id`, `name`, `email`, `slack`, `slackId` |
| Role | `id`, `name`, `holders` |
| Team | `id`, `name`, `contact`, `holders` |
| Variable | `name`, `description`, `example`, `type`, `multiple`, `options`, `pattern`, `patternHint`, `required`, `default`, `audience`, `askIf` |
| Condition | `all` |
| Clause | `question`, `answer`, or `section`, `enabled`. A clause that names a section is written with `section` and `enabled` only, never `question` or `answer`. A clause that names no section is written with `question` and `answer` only, never `enabled` |
| AnswerTest | `is`, `isNot`, `isOneOf`, `isNotOneOf`, `isSettled` |
| Section | `id`, `title`, `summary`, `includeIf`, `parent`, `bootstrap`, `timing`, `enabled`, `after`, `requires`, `owner`, `escalation`, `verification`, `body`, `variables`, `steps`, `source`, then the unknown fields |
| Source | `template`, `hash` |
| Step | `id`, `name`, `notes`, `owner`, `timing`, `escalation`, `run`, `open`, `directory`, `env`, `checks` |
| Check | `run`, `ask`, or `env`. Exactly one |
| Answers | `plan`, `sections` |

A plan's `id` is a plain text field. It is written right after `schemaVersion`, by the same rules as a string, and omitted when empty. So a UUID, which starts with a digit, is quoted. The four fields of an `origin` are plain text, written by the same rules. A `refreshedOn` date and a `contentHash` that starts with a digit are quoted, and a `source` such as a URL or a path that starts with a letter or `/` is plain. A plan holds an `origin` in a personal plan only, and the encoder writes whatever the plan holds. See [0075](../../../docs/decisions/0075-refreshing-a-personal-plan.md).

A plan's `planUrl` and `resultsPath` are plain text fields. `planUrl` is written right after `origin`, `resultsPath` right after `planUrl`, and both before `directory`. Each is written by the same rules as a string and omitted when empty, and each is written on its own, so a plan with only one of them writes only that one. They are written as held, with no check that one is a URL or that the other is a template, and nothing is changed on the way out. A URL that starts with a letter, such as `https://git.example/plan.obk.yaml`, is plain, and so is a template that starts with a letter and holds braces, such as `records/{account}/record.obk.yaml`. A value that starts with a character other than a letter, `$`, `/`, or `<` is quoted, so a template such as `{account}/record.obk.yaml` is quoted, which a file needs because a plain `{` starts a flow map, and so are `./records/{account}.yaml`, `~/records.yaml`, and a host that starts with a digit. A value with ` #` or `: ` in it, one that ends in `:`, and a reserved word such as `null` are quoted, and a value with a newline follows Multi-line strings. A `:` not followed by a space, a `#` not preceded by a space, and a `?` do not force quotes. See [0090](../../../docs/decisions/0090-plan-and-onboarding-record.md).

The encoder never writes a `kind`. A plan file has no `kind` key on the way out, and the decoder reads a missing kind as a plan.

A plan's `directory` is written right after `origin` and before `variables`. See [0080](../../../docs/decisions/0080-directory.md).

- It holds `contacts`, `roles`, and `teams`, each a list. Each item is a map written with the keys of its table, in that order. List items stay in the order the plan holds them, repeats and shared ids included. The encoder never sorts a list or removes an item.
- A key with a zero value is omitted, as everywhere. So an empty `contacts`, `roles`, or `teams`, an empty `holders` list on a role, and an empty or absent `holders` map on a team are omitted, and a contact that holds only an `id` writes only that key. The `directory` itself is omitted when it holds nothing, so a plan with no contacts, roles, or teams writes no `directory` key, whether its lists are nil or empty.
- A team's `contact` is the id of a directory contact that reaches the team. It is written after `name` and before `holders`, as a plain text field by the string rules, and omitted when empty. So `platform-team` is plain, and `42`, `2fa-team`, `@oncall`, `true`, and `deferred` are quoted. It is written as held, with no check that it names a contact. It is not a holder, so it is never written under `holders`, and a team that holds only a `contact` writes that one key.
- A team's `holders` is a map of role id to a list of contact ids. Its role ids are written in ascending byte order, like every other map, whatever the map's own order. Each role id is written as a key by the string rules, and each holder is written as a string by the string rules. An entry whose list is empty is the exception to the empty-list rule: it is written `role: []`, because a map entry is always written and the entry says the role has no holder for that team. Dropping it would give the role back to its organization-wide holders.
- Every directory field, holder, and role id is a plain text field written by the string rules in Scalars. That includes the values a person is likely to type: `sam-okafor`, `U01ABC123`, an email that starts with a letter, and `$HOME` are plain, and `@sam`, `#it-help`, an id or a holder that starts with a digit such as `42` or `2fa-team`, an email that starts with a digit, `true`, `on`, `null`, `y`, `n`, `no`, `yes`, and `deferred`, are quoted. A name with a newline follows Multi-line strings. A directory string is never changed on the way out.

An unknown field on a plan whose key is `id`, `origin`, `planUrl`, `resultsPath`, `kind`, or `directory` collides, and on a section it does not. `kind` is in the list although the encoder never writes one, because the decoder reads that key as the file kind and never keeps it as an unknown field, so a file written with one could not be read back as the plan it came from. The collision holds whether or not the plan holds a URL or a path, it is reported once, and the value is not examined. A section's own `planUrl`, `resultsPath`, and `kind` unknown fields are written as any other, and so are a variable or an answer with these names. The keys of an Origin are not keys of a Plan, so `planId` or `source` as an unknown field of a plan is written as any other.

A section's `parent` is a plain text field, the id of another section. It is written right after `includeIf`, by the same rules as `title`, and omitted when empty. An unknown field whose key is `parent` collides on a section and not on the plan. See [0070](../../../docs/decisions/0070-sub-sections.md).

A variable's `example` is a plain text field. It is written right after `description`, by the same rules as `description`, and omitted when empty. A variable that holds only an `example` writes that one key, so it is not an `encode.empty-item`.

A key is written by the same rules as a string, so a key such as `n` is quoted. Keys of a map are written in ascending byte order. This covers the plan's answers by variable name, the sections' answers by section id, a step's `env`, and unknown fields.

Unknown fields are written after the known keys, as siblings of them, in ascending byte order. An unknown field whose key is a known key of that type is a problem.

### Scalars

- An integer is written in plain decimal. A boolean is `true` or `false`, lowercase.
- A string is written plain only when all of these hold. Otherwise it is written in double quotes.
  - It is not empty, and does not start or end with a space.
  - Its first character is an ASCII letter, `$`, `/`, or `<`. A string that starts with any other character, such as `ü`, is quoted, and its characters are still written as they are.
  - It has no control character and no newline.
  - It does not contain `: ` or ` #`, and does not end with `:`.
  - It is not, ignoring case, one of `y`, `n`, `yes`, `no`, `true`, `false`, `on`, `off`, `null`, `inf`, `nan`.
  - It is not the Text value `deferred`.
- Inside double quotes, `\`, `"`, newline, tab, and return are written `\\`, `\"`, `\n`, `\t`, and `\r`. Other control characters are written `\xNN`. Every other character is written as it is.
- A date such as `2026-10-15` starts with a digit, so it is always quoted.

### Multi-line strings

A string with a newline is written as a literal block when all of these hold, and otherwise in double quotes.

- No line has trailing whitespace or a tab.
- No character is a control character, other than the newlines.
- The first line is not empty and does not start with a space.
- The text ends with no newline or with exactly one.

The marker is `|-` when the text has no final newline, and `|` when it has exactly one. A block's lines are indented two spaces more than their key, and an empty line inside the text is an empty line in the file.

### Lists

A list is written as a block list, one item per line, even for short lists of words. Items are written in the order the plan holds them, and repeats are kept. The encoder never sorts a list, reverses it, or removes an item. A list that is directly inside a list is written `- - item`. An empty list is omitted.

### Answers

| Value | Written as |
| -- | -- |
| Text | A string, by the string rules |
| Flag | `true` or `false` |
| List | A block list of strings, or `[]` when the list is empty, wherever a Value is written |
| Deferred | The plain word `deferred` |

A Value of kind none inside answers is a problem: `encode.empty-answer`.

### Unknown field values

The unknown fields of a plan or section hold maps, lists, strings, booleans, integers, decimal numbers, and null. They are written by the same rules, with null as `null`.

In Go, an integer is an `int` or an `int64`, and a decimal number is a `float64`. A decimal is written in plain notation, with `.0` added when it has no fractional part, so `2.0` stays a decimal. NaN, infinity, every other number kind, and every other kind of value are problems: `encode.unsupported-value`. The path of a value inside a list ends in the item's position, such as `x[0]`.

## Problems

| Code | When |
| -- | -- |
| `encode.empty-answer` | A Value of kind none is an answer, or is an item of `isOneOf` or `isNotOneOf` |
| `encode.empty-item` | A variable, section, step, contact, role, team, or condition clause in a list would write no keys. A clause that names no section and holds no question or answer writes no keys, whatever its `enabled` holds. Dropping it would shift the list, so it is a problem. A team whose `holders` map is empty and whose `contact` is empty writes no keys. A team with a `contact`, or whose map has an entry with no holder, does write a key, so it is not empty. The path is such as `directory.contacts[1]`. |
| `encode.ambiguous-check` | A Check has none of `run`, `ask`, `env`, or more than one |
| `encode.unknown-field-collision` | An unknown field has the key of a known field, so `enabled` or `parent` on a section collides, and the same keys on the plan do not. On the plan, `id`, `origin`, `planUrl`, `resultsPath`, `kind`, and `directory` collide, and the same keys on a section do not, so a section's own `directory` unknown field is written as any other, and a step's `directory` stays its working directory. It is reported once, and its value is not examined |
| `encode.unsupported-value` | An unknown field holds a value that cannot be written |

A problem names the place it was found in `detail.path`. The path is the place's address in the plan, written with dots for fields and a number in brackets for a list item, such as `sections[1].steps[0].checks[0]`, `answers.plan.organization`, or `answers.sections.s.name`. An unknown field at the plan level is its key alone, such as `answers`. Problems come in the order their places occur in the canonical file.

## Bootstrap

Under [0250](../../../docs/decisions/0250-bootstrap-content-and-projection.md),
write bootstrap: true after parent and before timing when Section.Bootstrap is
true; omit false. Preserve every other field and all ordinary list ordering.
Do not infer the marker from timing, answers, defaults, Enabled or hierarchy.
Treat bootstrap as a known section key for unknown-field collisions, including
when the held marker is false, and do not inspect a colliding value. At the
plan root it remains an ordinary unknown field, not a new known root key.
The [bootstrap fixtures](testdata/bootstrap.json) and [tests](bootstrap_test.go)
pin canonical bytes and collision paths independently of the product decoder.

The existing known-field collision rule applies even when a typed field is
zero and omitted. [Known-section-key tests](section_collision_test.go) exercise
every reserved section key with unsupported and scalar unknown values. Each
must return only encode.unknown-field-collision at sections[0].<key> and no
bytes, without inspecting the colliding value. Section ID is deliberately
nonempty so an empty-item guard cannot hide the collision requirement.

## Requires

Under [0260](../../../docs/decisions/0260-prerequisites-and-action-order.md),
write Requires as a string block list after After and before Owner, preserving
exact strings, order and repeats. Omit nil and empty Requires; do not infer it
from After or validate its references. A section holding only nonempty Requires
is not an empty item. Treat requires as a known section key even when its held
list is nil or empty. A colliding unknown field gives only
encode.unknown-field-collision at sections[n].requires, without traversing its
value. At the plan root and with a different case it remains unknown data.
Independent [fixtures](testdata/requires.json) and [tests](requires_test.go)
pin canonical bytes, reuse across inputs and collisions without decoder calls.

## Properties

These hold for every plan the contract tests use.

- **Deterministic.** Two calls with the same plan give identical bytes.
- **Inputs unchanged.** The plan is the same after the call as before.
- **Well formed.** The output ends in exactly one newline, has no `---`, and has no line with trailing whitespace.

## Not covered

- Reading plan files, which is the decoder.
- Writing the bytes to a file, which a driver does.
- Checking that a plan is valid. An encoder writes any plan it can, and the plan validator judges it. That includes whether a `parent` names a section of the plan and whether the parents form a cycle: the encoder writes the text as held, and the plan validator's structure check judges it.
- Decimal numbers written in exponent form, and decimals too large or small for plain notation.

## Decisions to know

These are the choices the directory change made where the docs were silent.

1. **hard.** A team's `holders` entry with no holder is written `role: []`, and is not omitted. The task said an empty list is omitted. The reason for the exception is the decoder contract: `[]` is how a file says a role is vacant for a team, and omitting it would give the role back to its organization-wide holders. A Go `nil` list and an empty list both write `[]`.
2. **hard.** Strings follow the existing quoting rule with no directory special case. So `U01ABC123`, emails, `sam-okafor`, `$HOME`, and `<noreply@acme.example>` are plain, and `@sam`, `#it-help`, `42`, `2fa-team`, `007`, `1password@acme.example`, `true`, `on`, `null`, and `deferred` are quoted, because the rule quotes anything that does not start with a letter, `$`, `/`, or `<`. The first-letter rule quotes some strings YAML would read as plain text, such as `Ünal@acme.example`. A file therefore holds more quotes than a hand-written one.
3. **cheap.** A contact, role, or team that would write no keys is `encode.empty-item` at its path, such as `directory.teams[0]`. The task did not name it. Writing `- {}` or dropping the item are the other choices, and dropping shifts the list.
4. **cheap.** An unknown field named `directory` on the plan collides, and on a section it does not, as for `origin`.
5. **cheap.** A team's `holders` with only an empty map is omitted, so a team with only that is an empty item. A team with only an entry that has no holder writes that entry.
6. **cheap.** A team's `contact` is written after `name` and before `holders`, the order the shared decisions gave, and an empty one is omitted like any empty text. A team with only a `contact` is not an empty item.

These are the choices the plan link and kind change made where the docs were silent.

1. **hard.** An unknown field named `kind` on a plan collides, although the encoder never writes a kind. The task named only `planUrl` and `resultsPath`. The decoder drops `kind: plan` and refuses any other kind, so a plan could never come back from a file that holds an unknown `kind`. Writing one would make a file the decoder refuses or silently changes. A section's `kind` does not collide.
2. **hard.** The two links follow the existing string rule with no special case. A URL is plain when it starts with a letter, a template such as `records/{account}/record.obk.yaml` is plain, and a template that starts with `{`, `.`, `~`, or a digit is quoted. A file therefore holds a quoted template in the common case that an account placeholder comes first.
3. **cheap.** The order is `origin`, `planUrl`, `resultsPath`, `directory`, as the task gave it, and each key is omitted on its own when empty.
4. **cheap.** The collision holds whether or not the plan holds a link, as for `id` and `origin`, and a plan's own link is not what is reported.
5. **cheap.** A link with a newline follows the multi-line rule like any string. The task did not ask for it, and it is unlikely in practice.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| Omit Requires, write it in the wrong key position, normalize its list, copy After into it, or miss a zero-held known-key collision | [Requires fixtures](testdata/requires.json), [tests](requires_test.go) and [known section keys](section_collision_test.go) |
| Forget a reserved section key when its typed field is zero, or traverse its colliding value | [all known section keys](section_collision_test.go) |
| Drop true Bootstrap or emit bootstrap false instead of omitting it | [bootstrap fixtures](testdata/bootstrap.json) |
| Map keys in iteration order, not sorted | The golden cases that have maps, and the determinism test |
| `yes`, `no`, `on`, and other reserved words written plain | `strings`, `key-order`, `unknown-fields` |
| Lists in flow style, `[a, b]` | `key-order`, `acme` |
| Zero fields written, such as `required: false` | `omit-zero`, `multiline`, and others |
| A list sorted, reversed, or stripped of repeats, in any array field | `list-order` |
| The argument changed during the call | The inputs-unchanged test |
| No final newline | Every golden case, and the well-formed test |
| `True` and `False` for booleans | `answers`, `conditions` |
| Unknown fields dropped | `unknown-fields`, and the collision cases |
| Keys in alphabetical order, not canonical order | Every case with more than one key |
| A date written plain | `minimal`, `key-order`, `strings` |
| Every multi-line string as a literal block | `multiline`, the well-formed test |
| A string with two final newlines as `\|+` | `multiline` |
| An empty answer written as `null`, no problem | `empty-answer` |
| The Text `deferred` written plain | `answers`, `strings` |
| A section's `enabled` never written | `section-switch`, `enabled-unknown`, the switch test |
| `enabled: false` omitted, as a zero value | `section-switch`, the switch test |
| `enabled: true` omitted, as the default | `section-switch`, the switch test |
| `enabled: true` written for a section that holds no value | `section-switch`, and every case with a section |
| A section's `enabled` at the end, after `source`, or after `id`, or after `after` | `section-switch` |
| A section clause's `enabled: false` omitted | `section-clause`, the section clause test |
| A section clause written `enabled` before `section` | `section-clause` |
| A section clause that also writes `question` and `answer` | The section clause test |
| A section clause dropped, or written only when it has no question | `section-clause`, the section clause test |
| A question clause that writes `enabled: true` | The question clause test |
| A clause with only `enabled` written, not reported | The switch-value test |
| `enabled` as an unknown field of a section not reported | `enabled-collision` |
| `enabled` as an unknown field of the plan reported | `enabled-unknown` |
| `example` never written | `example` |
| `example` written before `description`, or after `type` or `askIf` | `example` |
| `example` written when empty, as `example: ""` | `example` |
| `example` written plain when it needs quotes, such as `yes` or text with `: ` | `example` |
| A multi-line `example` written in double quotes, not as a literal block | `example` |
| A variable with only an `example` reported as `encode.empty-item` or dropped | `example` |
| `example` written for a section's variables only, or only for the plan's | `example` |
| `parent` never written | `sub-section` |
| `parent` written after `timing`, `enabled`, or `source`, or before `includeIf` | `sub-section` |
| `parent` written when empty, as `parent: ""` | `sub-section` |
| `parent` changed on the way out, such as trimmed or lowercased | `sub-section` |
| `parent` written plain when it needs quotes, such as `on`, `Team: core`, or a leading space | `sub-section` |
| `parent` dropped for a sub-section of a sub-section, or when the parent is not a section of the plan | `sub-section` |
| `parent` as an unknown field of a section not reported | `parent-collision` |
| `parent` as an unknown field of the plan reported | `parent-unknown` |
| `id` never written | `id-only`, `id-origin` |
| `id` written after `provenance`, or at the end | `id-origin` |
| `id` written plain when it needs quotes, such as a UUID that starts with a digit, or `null` | `id-origin`, `origin-quoting` |
| `id` written when empty, as `id: ""` | every case with no `id` |
| `origin` never written | `id-origin`, `origin-planid-refreshed`, `origin-source-hash`, `origin-quoting` |
| `origin` written before `provenance`, or after `variables` | `id-origin`, `origin-source-hash` |
| `origin` keys in alphabetical order, not canonical order | `id-origin`, `origin-quoting`, `origin-source-hash` |
| An empty `origin` written as `origin: {}`, or an empty key written as `source: ""` | every case with no `origin`, and `origin-planid-refreshed` |
| One origin key never written, or all four written only when `planId` is set | `id-origin`, `origin-planid-refreshed`, `origin-source-hash` |
| `origin` dropped when the plan also holds an `id` | `id-origin`, `origin-quoting` |
| `refreshedOn` or `contentHash` written plain when it starts with a digit | `id-origin`, `origin-quoting`, `origin-planid-refreshed` |
| `id` or `origin` as an unknown field of the plan not reported | `id-origin-collision` |
| The keys of an Origin reported as collisions on the plan, or `origin` reported on a section | `origin-unknown` |
| `directory` never written, or written after `variables`, `answers`, or `sections`, or before `id`, `provenance`, or `origin` | `directory-order`, `directory-shared`, `directory-problem-order` |
| `contacts`, `roles`, and `teams` written in another order, or a field of an item written out of order, such as `slackId` before `slack` or `holders` before `name` | `directory-order`, `directory-shared`, `directory-quoting` |
| A contact, role, team, or holder list sorted, reversed, deduplicated, or stripped of an item that shares an id | `directory-order` |
| A list sorted in place, so the plan the caller holds changes | `directory-order`, the inputs-unchanged test |
| A team's role ids in map order, or sorted without regard to case | `directory-order`, `directory-quoting`, `directory-vacant-role`, the determinism test |
| A team's entry with no holder dropped, written as `null`, or written as an empty block | `directory-vacant-role`, `directory-team-contact`, the vacant role test |
| A team's `contact` never written, or written under `holders` or as a holder | `directory-team-contact`, `directory-team-contact-only` |
| A team's `contact` written before `name`, or after `holders`, or first among the team's keys | `directory-team-contact` |
| A team's empty `contact` written as `contact: ""` or `contact: null` | `directory-team-contact` |
| A team's `contact` written plain when it needs quotes, such as `42`, `2fa-team`, `true`, `on`, `deferred`, `@oncall`, a date, or text with `: ` or ` #`, or in quotes when it is plain, such as `platform-team`, `$HOME`, or an email that starts with a letter | `directory-team-contact-quoting`, `directory-team-contact` |
| A team's `contact` with a newline written in quotes, or as the wrong block | `directory-team-contact-quoting` |
| A team's `contact` lowercased, trimmed, deduplicated across teams, or dropped when it names no contact | `directory-team-contact` |
| A team that holds only a `contact` taken for an empty item, or dropped, or its directory omitted | `directory-team-contact-only`, `directory-team-contact-empty-items` |
| A team with an empty `contact` and nothing else not reported as an empty item, or a team with an empty `contact` and a vacant entry reported | `directory-team-contact-empty-items` |
| An empty `holders`, an empty `contacts`, or a directory that holds nothing written, as `holders: {}`, `contacts: []`, or `directory: {}` | `directory-omit-zero`, `directory-only-teams`, `directory-blank`, the empty directory test |
| A zero field written, such as `slack: ""` | `directory-order`, `directory-omit-zero` |
| A field dropped, such as `slackId` | `directory-order`, `directory-shared`, `directory-quoting` |
| `@sam`, `#it-help`, `U01ABC123`-like ids, an id that starts with a digit, a number-like id, `true`, `on`, `null`, or `deferred` written plain, or text that is plain written in quotes | `directory-quoting`, `directory-shared` |
| A role id in a team's `holders` written without the key quoting, such as `2fa`, `@it`, `n`, `on`, or `deferred` | `directory-quoting` |
| A name with a newline written in quotes, or as the wrong block | `directory-quoting` |
| An empty contact, role, or team dropped without a problem, or written as `{}` | `directory-empty-items`, `directory-problem-order` |
| A team with an empty `holders` map taken for a team with content, or a team with only a vacant entry reported | `directory-empty-items` |
| A plan's `directory` unknown field written, or not reported | `directory-collision` |
| A section's `directory` unknown field reported, or a step's `directory` changed | `directory-section-unknown` |
| `planUrl` or `resultsPath` never written, or written only when the other is set | `plan-links`, `plan-links-url-only`, `plan-links-path-only` |
| The links written before `origin`, after `directory`, or in the order `resultsPath`, `planUrl` | `plan-links`, `plan-links-url-only`, `plan-links-path-only` |
| An empty link written as `planUrl: ""` or `resultsPath: null` | every case with no link, and `plan-links-blank` |
| A link always plain, or always quoted | `plan-links-brace-first`, `plan-links-punctuation`, `plan-links-reserved-words`, `plan-links`, the plan link string test |
| A template that starts with `{` written plain | `plan-links-brace-first`, the plan link string test |
| A link that holds `: ` or ` #` written plain, or one with `:` or `#` and no space written in quotes | `plan-links-punctuation`, `plan-links-plain-punctuation`, the plan link string test |
| A link lowercased, trimmed, or otherwise changed | the plan link string test, `plan-links` |
| A link with a newline written in quotes, or as the wrong block | the plan link block test |
| A `kind` written, such as `kind: plan` | every golden case, and the no-kind test |
| `planUrl`, `resultsPath`, or `kind` as an unknown field of the plan accepted, or written | `plan-links-collision`, `plan-links-collision-with-values`, the reserved key test |
| One of the three keys reported and not the others | the reserved key test |
| The collision value examined | the reserved key test |
| `planUrl`, `resultsPath`, or `kind` as an unknown field of a section reported, or dropped | `plan-links-section-unknown` |
| A variable or an answer named `planUrl`, `resultsPath`, or `kind` treated as a collision or changed | `plan-links-answers-named-alike` |
