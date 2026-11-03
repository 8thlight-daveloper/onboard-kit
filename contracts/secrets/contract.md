# Secrets check contract

Status: draft by an agent. Not reviewed by a person. Component: `secrets`.

The secrets check decides whether a value looks like a secret, and says what kind it looked like. It is a pure function. It reads no file, keeps no state, and does not change its argument. The plan validator calls it when an answer is submitted and when a plan is validated, so a secret typed in or added by hand is caught. See [0010](../../docs/decisions/0010-plan-editor.md), [0030](../../docs/decisions/0030-self-contained-plan-file.md), the [domain model](../../docs/reference/domain-model.md), and [0015](../../docs/decisions/0015-functional-core-and-imperative-shell.md). Obvious secrets such as tokens are rejected with a message.

The fixtures are made up. Every secret in `testdata/` is a plain shape filled with repeated or invented characters, such as a prefix followed by `FAKE` many times. None is a real credential, and none may become one. Each prefix of a well-known token in a fixture is followed by the character `¤`, which the tests remove, so no file holds a whole token and a secret scanner has nothing to match.

## Interface

`Check(value) -> problems`.

- `value` is a Value, as stored. The caller passes one value. For a step's `env` map the caller passes each value on its own, as Text.
- `problems` is a list of Problem. It is empty when the value does not look like a secret.

The check does not know which question or section the value belongs to. It leaves `question` and `section` empty, and the caller fills them in.

## What is checked

| Value | Checked |
| -- | -- |
| Text | The whole text |
| List | Each item on its own, as Text |
| Flag, Deferred, a Value of no kind | Never. No problems |

Only the kind of the Value decides which field is read. The fields of other kinds are ignored. Empty text, text of only spaces, and the Text `deferred` are ordinary text, and none of them looks like a secret.

## What looks like a secret

A text is tested for four shapes, in this order. The first one that matches decides, so one text gives at most one problem.

| Order | Kind | Code |
| -- | -- | -- |
| 1 | Private key block | `secret.private-key` |
| 2 | Password in a URL | `secret.url-credential` |
| 3 | Well-known token prefix | `secret.token-prefix` |
| 4 | Long random-looking string | `secret.random-string` |

### Private key block

The text contains a match for `-----BEGIN (?:[A-Z]+ )*PRIVATE KEY( BLOCK)?-----`, in RE2 syntax. The header alone is enough, so a text that holds only the header is a match. A public key, a certificate, the words "private key" in a sentence, a header in lower case, and an `END` line alone do not match.

### Password in a URL

The text contains a scheme, `://`, and then a user part, a colon, and a password, followed by `@`, with no space, `/`, `?`, or `#` between `://` and the `@`. In RE2 syntax: `[A-Za-z][A-Za-z0-9+.-]*://[^\s/?#@:]*:([^\s/?#@]+)@`, and the group is the password.

- Every match in the text is tested, not only the first, so a template in one URL does not hide a password in another. A match counts unless the password is exactly one `$name` variable reference, `\$[a-z][a-z0-9_]*`. Templates write `https://$user:$token@host`, and a variable reference is not a secret. A password that only starts with a variable, or a reference with an upper case letter, counts.
- The user part may be empty.
- A user with no password does not count: `ssh://git@github.com/org/repo.git` and `https://deploy@host/repo.git`. An empty password does not count either.
- A Git SSH address such as `git@github.com:org/repo.git` has no `://`, so it never counts.
- A port is not a password: `https://host:8443/repo.git` has no `@` before the first `/`, so it does not count. An `@` after the path, query, or fragment starts does not count.

### Well-known token prefix

The text contains a prefix from the table, followed by a run of body characters at least as long as the minimum. A longer run counts. The prefix is case sensitive, and it counts only when it is at the start of the text or the character before it is not an ASCII letter or digit. A body is the longest run of its characters that follows the prefix. A first character outside the set leaves a body of length 0, whatever follows it. Only the listed characters belong to a body. A character that sorts after `_` in ASCII, such as a backtick, `{`, `|`, `}`, or `~`, and any non-ASCII character, ends it. A prefix after a letter or digit is not a token, but the text may still report `secret.random-string`: `0ghp_` followed by 36 upper case letters does.

| Token | Prefix | Body characters | Minimum body |
| -- | -- | -- | -- |
| GitHub token | `ghp_`, `gho_`, `ghu_`, `ghs_`, `ghr_` | `A-Za-z0-9` | 36 |
| GitHub fine-grained token | `github_pat_` | `A-Za-z0-9_` | 22 |
| AWS access key id | `AKIA`, `ASIA` | `A-Z0-9` | 16 |
| Slack token | `xoxa-`, `xoxb-`, `xoxp-`, `xoxr-`, `xoxs-` | `A-Za-z0-9-` | 10 |
| Stripe live key | `sk_live_`, `rk_live_` | `A-Za-z0-9` | 16 |
| Google API key | `AIza` | `A-Za-z0-9_-` | 35 |
| GitLab token | `glpat-` | `A-Za-z0-9_-` | 20 |
| npm token | `npm_` | `A-Za-z0-9` | 36 |
| Secret key | `sk-` | `A-Za-z0-9_-` | 40 |

A JSON web token also counts: `eyJ` and at least 10 characters of `A-Za-z0-9_-`, a dot, `eyJ` and at least 10 more, a dot, and at least 10 characters of `A-Za-z0-9_-`. The boundary rule above applies to the first `eyJ` like it does to every prefix: it counts only at the start of the text or after a character that is not an ASCII letter or digit. So `xeyJ` and `7eyJ` do not count, and `x.eyJ` does.

Stripe test keys do not count.

### Long random-looking string

The text is split into words at white space, which means the characters that Unicode calls space, including tab and line breaks. For each word:

1. Remove from both ends every character in `"`, `'`, a backtick, `,`, `;`, `:`, `.`, `(`, `)`, `[`, `]`, `{`, `}`, `<`, and `>`.
2. The rest counts when it is at least 32 characters, every one of them in `A-Za-z0-9_+=-`, and it holds at least one ASCII upper case letter, one ASCII lower case letter, and one digit.

This is a shape test and not a measure of randomness. A string of repeated characters that has all three kinds counts. A word with a `/`, `\`, `.`, `:`, or any character outside the set in its middle never counts, so URLs, paths, and a dotted version are skipped whole.

What it skips on purpose: a commit hash, a SHA-256 digest, and an upper case hash, which have no mixed case; a UUID; a long lower case slug; and a Google document link, which holds `/`. What it still catches as a false alarm: a long identifier that mixes cases and digits with no spaces, such as `OnboardingChecklistForNewEngineers2026Edition`, and a public SSH key line. A person who hits one can write the answer another way.

## The problem

Every problem has:

| Field | Value |
| -- | -- |
| `code` | One of the four codes above |
| `severity` | `error`, so the answer is not applied |
| `message` | Plain text for a person, not empty. It says what kind of secret the value looked like and tells the person to keep it out of the plan. It is the same text for the same code |
| `question`, `section` | Empty. The caller fills them in |
| `detail` | Empty for a Text value. For a List item, `item` holds the zero-based position of the item, as text, such as `1` |

A problem never repeats the secret or any part of it, in the message, the detail, or the code.

A List gives one problem for each item that looks like a secret, in item order. An item that holds two kinds gives one problem, by the order above.

## Properties

- **Pure.** The Value is the same after the call as before.
- **Same value, same problems.** Two calls give equal results. A call does not depend on any call before it, forward or in reverse order.
- **Kind decides.** A Flag, a Deferred, and a Value of no kind give no problems, whatever the other fields hold.

## Not covered

- The wording of the message. Only its rules are tested: not empty, the same for one code, and no part of the secret.
- A password in words, such as `password: hunter2`, and a short or ordinary word used as a password. No shape marks them.
- A token whose prefix is not in the table, and a Stripe test key.
- A random string that holds a `/`, such as some base64 keys, or one inside a URL. Words with a `/` are skipped.
- A value that is a pointer to a secret, such as the name of an environment variable. An `env` check holds only a name.
- Which values the plan validator passes, and how it reports. The caller picks them. The component sees one Value.
- Text that is not valid UTF-8.

## Decisions to know

`hard` choices change what a person sees or what the callers must do. They come first.

1. **Hard. Every finding is an `error`, with no way to override it, and the shapes lean toward catching.** A false alarm blocks a consultant from entering an answer, and a miss puts a secret in a plan file. [Issue 7](https://github.com/8thlight/onboard-kit/issues/7) asks for secrets to be "rejected with a message", so a finding stops the answer. The known false alarms are a long mixed-case identifier and a public SSH key. The known misses are listed under Not covered. Raised by: "Obvious secrets such as tokens are rejected with a message" in [issue 7](https://github.com/8thlight/onboard-kit/issues/7), and "An `error` stops an answer from being applied" in the domain model.
2. **Hard. Four shapes count: a private key header, a password in a URL, a well-known token prefix with a body, and a long word that mixes upper case, lower case, and digits.** A plain word used as a password, a `password:` label, and a Stripe test key do not. The long-word rule is a shape test with a length of 32, not an entropy measure, so a builder has no threshold to tune and a fixture can say exactly where the line is. Words that hold a `/` are skipped, which keeps document links and paths quiet and lets some base64 keys through. Raised by: "decides whether a value looks like a token" in the components table, and "the secrets check also looks inside a URL for an embedded credential" in the variables reference.
3. **Hard. The result is a list of Problem with the four codes, not a yes or no.** It is one problem for a Text and one per bad item for a List, with `item` in the detail. The message never repeats the secret, because the plan file and the interface would then show it. This is what the validator and the editor pass on, so the codes are the names they will report. Raised by: "A report about a plan or one value" in the domain model, and "Problem: value looks like a credential" in 0010.
4. **Hard. The function takes one Value and leaves `question` and `section` for the caller.** The check has no plan and no id, so it cannot say where the value sits. The validator fills them in, as it does for its own problems. Raised by: `CheckValue(plan, id, value)` in 0010, which holds the id and the plan, not the secrets check.
5. **Cheap. A deferred, empty, Flag, or no-kind value is never a problem.** The Text `deferred` is checked as text and gives none. Raised by: "The literal `deferred`, accepted for every type" in the domain model.
6. **Cheap. A password that is one `$name` reference is not a credential.** `https://$user:$token@host` is a template, and the value is filled in later. A password that only starts with a variable still counts. Raised by: "Values may use `$name` variables. The secrets check runs on them like any other value" in the steps reference.
7. **Cheap. A URL with a user and no password does not count.** `ssh://git@github.com/org/repo` is a normal clone address. A token used as the user, as in `https://ghp_...@host`, is caught by the token prefix instead. Raised by: "A clone address such as `git@host:org/repo.git`" in the variables reference.
8. **Cheap. The first matching kind wins, in the order private key, URL password, token prefix, random string.** One value gives one problem, so a person does not see two messages for one secret. Nothing in the docs sets an order.
9. **Cheap. The token table.** The prefixes and minimum lengths are the common public formats at the time of writing. Adding or dropping a row touches one fixture.

## Wrong implementations tried

Each row names a plausible wrong implementation and the groups of cases that fail it. How the contract was checked, and with what scores, is recorded in the issue, not here.

| Wrong implementation | Caught by |
| -- | -- |
| Private key needs the body or the `END` line, accepts only one header, ignores case, or counts a public key | `private-keys` |
| URL rule: a user with no password counts, an empty password counts, a port counts, the URL must start the text, only `http` and `https` count, the scheme has no digit or plus | `url-credentials` |
| A `$name` password counts, or any password that starts with `$` does not | `variable-references` |
| Only the first one or two URLs in a text are tested, so a template ahead of a real password hides it | `variable-references` |
| Token boundary ignored, or too wide, or only white space | `token-boundaries` |
| Token body minimum too small, or a body that must end the run | `token-boundaries`, `token-prefixes` |
| A token family left out or cut short: GitHub kinds, AWS, Slack, Stripe, Google, GitLab, npm, `sk-`, web tokens | `token-prefixes` |
| Stripe test keys count | `token-boundaries` |
| Random string: minimum 24 or 40, no digit, upper, or lower case needed, a hex string counts | `random-strings`, `paths-and-ids` |
| Random string: `/` or `.` allowed in a word, `-` not allowed, words split only on spaces, the whole value read as one word, only the first word read | `random-strings`, `paths-and-ids` |
| Random string: one trimmed character left out, such as a quote, a backtick, a bracket, a semicolon, a colon | `random-strings` |
| Order wrong: token before URL, random before token, private key last | `kind-order` |
| A range in a body or in the random-string set stops one short: `Z`, `z`, or `9` left out, or `A`, `a`, or `0` left out | `token-body-boundary-characters`, `random-string-characters` |
| A range runs one too far: `@`, `[`, a backtick, `{`, `/`, or `:` accepted in a body or in a random string | `token-body-boundary-characters`, `random-string-characters` |
| The AWS body takes letters only, or digits do not count toward its 16 | `aws-body-digits` |
| A family takes the wrong extra characters: no `-` in a Slack body, no `_` in a `github_pat_` body, no `_` or `-` in a Google, GitLab, or `sk-` body, or an `_`, `-`, or `+` where the table gives none | `token-body-special-characters` |
| A body accepts a character that sorts after `_` in ASCII, or a non-ASCII character, in a `github_pat_`, Google, GitLab, or `sk-` body | `token-body-characters-above-underscore` |
| The first body character is not read, so a prefix, one character outside the family's set, and a full-length body counts | `token-first-body-character` |
| A minimum body is one too large or one too small: GitHub 36, fine-grained 22, GitLab 20, npm 36, `sk-` 40, Slack 10, Stripe 16 | `token-minimum-lengths` |
| Random string: the first or last character is not read, so a word with one bad edge character counts, or one good edge character is the only upper case letter, lower case letter, or digit and is missed | `random-string-characters` |
| Random string: `-`, `_`, `+`, or `=` is not in the set, or does not count as a character of the 32 | `random-string-characters` |
| A prefix after a digit hides the whole text from the random string rule | `token-boundary-and-random-string` |
| Web token ignores the boundary rule, or a part has a minimum of 9 or 11 | `web-token-boundaries` |
| Message or detail repeats the value | Every case that has a secret |
| Severity `warning`, an empty message, or `question` filled in | Every case with a problem |
| A Text also gets an `item`, a List item missing, or counted from one | `list-items` and every case with a problem |
| A List stops at the first problem, joins its items, or reads only the first item | `list-items` |
| Deferred or Flag read as text, or the text field read for a List, or the list field read for a Text | The kind test, `values-that-are-not-text` |
| Blank text counted as a secret | `values-that-are-not-text` |
| Words such as "password" or an email address counted | `ordinary-values`, `url-credentials` |
| One code for every kind, or a code renamed | Every case with a problem |
| The Value changed | The inputs-unchanged test |
| A cache that remembers an earlier call | The no-memory test, or a crash under the race detector |
