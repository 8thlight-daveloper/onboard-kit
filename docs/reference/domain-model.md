# Domain model

Status: proposed

The types that several components pass to each other. They live in one protected folder, import nothing from the project, and change only by ADR, because the stored types are the plan file format in memory. See [0020](../decisions/0020-binary-structure-and-regenerative-components.md) and [0030](../decisions/0030-self-contained-plan-file.md).

This page is language-neutral. A type here is a record with named fields. A field marked optional may be absent. The code layout for the current language is in [code layout](../stack/code-layout.md).

## Stored and computed

| Kind | Types | Lives in |
| -- | -- | -- |
| Stored | Plan, Provenance, Variable, Answers, Value, Section, Source, Variable, Condition, Clause, AnswerTest, Step, Check | The plan file |
| Computed | Question, Suggestion, Blocker, Problem | Returned by the finder, validator, and editor. Never saved |
| Output | Document | Returned by the document assembler. Transitional, see [0025](../decisions/0025-contracts-and-tests.md) |

Stored types are what the decoder reads and the encoder writes. Computed types are what drivers show and what acceptance scenarios assert on, so their serialized shape is part of the `stdio` protocol in [0010](../decisions/0010-plan-editor.md).

```mermaid
classDiagram
    class Plan {
        schemaVersion
        id
        provenance
        origin
        directory
        variables
        answers
        sections
        unknownFields
    }
    class Provenance {
        templateSet
        generatedOn
    }
    class Origin {
        planId
        source
        contentHash
        refreshedOn
    }
    class Directory {
        contacts
        roles
        teams
    }
    class Answers {
        plan
        sections
    }
    class Section {
        id
        title
        summary
        owner
        timing
        enabled
        verification
        body
        includeIf
        parent
        after
        escalation
        variables
        steps
        source
        unknownFields
    }
    class Source {
        template
        hash
    }
    class Variable {
        name
        description
        type
        options
        pattern
        patternHint
        required
        multiple
        default
        audience
        askIf
    }
    class Condition {
        all
    }
    class Clause {
        question
        answer
        section
        enabled
    }
    class Step {
        id
        name
        notes
        owner
        timing
        run
        open
        directory
        env
        escalation
        checks
    }
    class Check {
        run
        ask
        env
    }

    Plan "1" --> "1" Provenance
    Plan "0..1" --> "1" Origin
    Plan "0..1" --> "1" Directory
    Plan "1" --> "*" Variable : plan variables
    Plan "1" --> "1" Answers
    Plan "1" --> "*" Section
    Section "0..1" --> "1" Source
    Section "1" --> "*" Variable
    Section "1" --> "*" Step
    Section "0..1" --> "1" Condition : includeIf
    Variable "0..1" --> "1" Condition : askIf
    Condition "1" --> "*" Clause
    Step "1" --> "*" Check
```

## Plan

| Field | Type | Notes |
| -- | -- | -- |
| `schemaVersion` | Integer | See [0030](../decisions/0030-self-contained-plan-file.md#versioning) |
| `id` | Text, optional | An org plan's identity, a UUID the driver makes when the plan is first created. See [0075](../decisions/0075-refreshing-a-personal-plan.md) |
| `provenance` | Provenance |  |
| `origin` | Origin, optional | Present in a personal plan only: the org plan it was refreshed from. To be retired by [0090](../decisions/0090-plan-and-onboarding-record.md) |
| `planUrl` | Text, optional | Where the plan's latest copy lives, a path or an HTTPS address. Set when the plan is created. See [0085](../decisions/0085-walking-sections-and-recording-outcomes.md) |
| `resultsPath` | Text, optional | A template for where each newcomer's onboarding record is kept, such as `records/{account}/record.obk.yaml`. `{account}` is its only variable |
| `directory` | Directory, optional | Contacts, roles, and teams, in the org plan. See [0080](../decisions/0080-directory.md) |
| `variables` | List of Variable | Plan variables: the axis variables, document details, and settings |
| `answers` | Answers | Every answer, by scope |
| `sections` | List of Section | Every template, as a section |
| `unknownFields` | Map of unknown fields | Fields the tool does not know. Kept on load and written back, so a plan from a newer tool survives an older one |

A Plan is a value. No function changes one, and an edit returns a new Plan. A Section in a plan and a section template are the same type. A template is a Section with no `source`. The file's `schemaVersion` belongs to the decoder and the migrator, not to the Section.

## Provenance

| Field | Type | Notes |
| -- | -- | -- |
| `templateSet` | Text | The build's commit sha or tag |
| `generatedOn` | Date | Text in ISO form, such as `2026-10-15`. A driver passes it in, because the core never reads a clock |

## Onboarding record

The newcomer's own file. See [0090](../decisions/0090-plan-and-onboarding-record.md). It holds only what the newcomer produced, and it is read and written apart from the plan.

| Field | Type | Notes |
| -- | -- | -- |
| `schemaVersion` | Integer | As for a plan |
| `kind` | Text | Always `record`. A decoder refuses a file of the wrong kind. It is part of the file and not of the Record type |
| `account` | Text | The permanent key. It is in the path of the file |
| `startDate` | Date | The day the newcomer starts. Due dates count from it |
| `plan` | PlanReference | Which plan the record was last reconciled against |
| `answers` | Answers | Only the answers the newcomer gave, to onboardee questions |
| `outcomes` | List of Outcome | One entry for each section the newcomer tried, sorted by `section` |
| `unknownFields` | Map of unknown fields | Kept on load and written back, as for a plan |

A PlanReference has `id`, `hash`, and `url`: the plan's id, the hash of that plan, and where its latest copy lives.

An Outcome has `section` (a section id), `outcome` (`verified`, `failed`, or `deferred`), `on` (a UTC time in RFC 3339 form, passed in by a driver), `sectionHash` (the section's content hash when the outcome was recorded), and optional `note` and `differently`. A section with no outcome has not been tried.

## Origin

| Field | Type | Notes |
| -- | -- | -- |
| `planId` | Text | The `id` of the org plan. A different id means a different organization |
| `source` | Text | A path or an HTTPS URL the org plan came from |
| `contentHash` | Text | The hash of the org plan's file when last refreshed from. A driver computes it |
| `refreshedOn` | Date | Passed in by a driver, because the core never reads a clock |

## Directory

Ids are permanent. An answer to a `contact`, `role`, or `team` variable stores an id and never a name.

| Type | Field | Type | Notes |
| -- | -- | -- | -- |
| Contact | `id`, `name` | Text | A person or a shared contact such as a help channel |
| Contact | `email`, `slack`, `slackId` | Text, optional | `slack` is for display, and `slackId` is for links |
| Role | `id`, `name` | Text | A named job |
| Role | `holders` | List of contact ids, optional | The organization-wide holders, in order |
| Team | `id`, `name` | Text | Teams are a flat list |
| Team | `contact` | Text, optional | The id of a contact for reaching the team, such as its channel or on-call alias |
| Team | `holders` | Map of role id to a list of contact ids, optional | Who holds the role for this team, in order. It overrides the role's own holders |

## Answers and Value

`answers` holds the answers, by scope.

| Field | Type | Notes |
| -- | -- | -- |
| `plan` | Map of name to Value | Answers to plan variables |
| `sections` | Map of section id to a map of name to Value | Each section's own answers |

A Value is one of:

| Variant | Holds | Written as |
| -- | -- | -- |
| Text | Text | A string |
| Flag | Boolean | `true` or `false`, lowercase |
| List | List of text | A list. Used by `command` variables and `multiple` variables |
| Deferred | Nothing | The literal `deferred`, accepted for every type. The person chose to answer later |

A Value is a settled answer unless it is Deferred or empty. See [conditions](section-template-schema/conditions.md).

## Section

| Field | Type | Notes |
| -- | -- | -- |
| `id` | Text | A lowercase slug. Never `plan` |
| `title`, `summary` | Text | The only required fields besides `id`. Blank text counts as missing, see [0040](../decisions/0040-blank-text-and-variable-name-length.md) |
| `owner` | Text, optional | May contain `$name` |
| `timing` | Timing, optional | Defaults to `anytime` |
| `enabled` | Boolean, optional | True by default. Only a plan section has it, and a person turns a section off. See [0050](../decisions/0050-section-switch.md) |
| `verification`, `body` | Text, optional | Markdown prose |
| `parent` | Text, optional | The id of another section. A section with a parent is a sub-section: it is active only if its parent is. See [0070](../decisions/0070-sub-sections.md) |
| `includeIf` | Condition, optional | Names plan variables and the questions of its ancestors |
| `after` | List of section ids, optional |  |
| `escalation` | List of variable names, optional | Who to ask when the section is blocked. Names of `contact` or `role` variables the section declares. See [escalation](section-template-schema/escalation.md) |
| `variables` | List of Variable, optional |  |
| `steps` | List of Step, optional |  |
| `source` | Source, optional | Absent on a template and on a section added by hand |
| `unknownFields` | Map of unknown fields | Kept in a plan, rejected in a template |

## Source

| Field | Type | Notes |
| -- | -- | -- |
| `template` | Text | The template id |
| `hash` | Text | The content hash from the section hasher, computed over everything except `source` |

## Variable

| Field | Type | Notes |
| -- | -- | -- |
| `name` | Text | `[a-z][a-z0-9_]{2,}`, not reserved. See [0040](../decisions/0040-blank-text-and-variable-name-length.md) |
| `description` | Text |  |
| `example` | Text, optional | An illustration of a good answer. Never an answer or a default |
| `type` | VariableType | `string` by default |
| `options` | List of text, optional | For a `choice` |
| `pattern`, `patternHint` | Text, optional | RE2 syntax, for a `string` |
| `required` | Boolean, optional |  |
| `multiple` | Boolean, optional | Takes several answers, held as a list. Only `contact` and `role` allow it |
| `default` | Value, optional | Must satisfy the type, options, and pattern |
| `audience` | Audience, optional | `onboarder` by default |
| `askIf` | Condition, optional |  |

## Condition, Clause, and AnswerTest

| Type | Field | Notes |
| -- | -- | -- |
| Condition | `all` | List of Clause. Every clause must hold |
| Clause | `question` | A question id. A bare name means the same section |
| Clause | `answer` | An AnswerTest. A clause with `question` has `answer` |
| Clause | `section` | A section id. A clause with `section` tests that section's switch, and has `enabled` and no `question` |
| Clause | `enabled` | Boolean. The clause holds when the named section's switch equals it |
| AnswerTest | `is`, `isNot` | One Value |
| AnswerTest | `isOneOf`, `isNotOneOf` | A list of two or more Values |
| AnswerTest | `isSettled` | Boolean |

Several tests in one AnswerTest must all hold. A section uses a Condition only through `includeIf`, and a variable only through `askIf`.

## Step and Check

| Field | Type | Notes |
| -- | -- | -- |
| `id`, `name` | Text |  |
| `notes` | Text, optional | Markdown prose |
| `owner` | Text, optional |  |
| `timing` | Timing, optional | Inherited from the section |
| `run` | List of text, optional | Each item may contain `$name`, or one `$name` that expands to a list |
| `open` | Text, optional | A URL or file |
| `directory` | Text, optional |  |
| `env` | Map of name to text, optional | Names are uppercase letters, digits, and underscores |
| `escalation` | List of variable names, optional | Who to ask when the step is blocked. Inherits the section's when absent |
| `checks` | List of Check, optional |  |

A Check has exactly one of `run` (a list of text), `ask` (text, a yes or no question), or `env` (a variable name). The meaning of each is in [steps](section-template-schema/steps.md).

## Enumerations

| Name | Values |
| -- | -- |
| VariableType | `string`, `boolean`, `choice`, `command`, `contact`, `role`, `team` |
| Audience | `onboarder`, `onboardee` |
| Timing | `before-day-one`, `week-one`, `days-8-to-30`, `days-31-to-90`, `anytime`. Each has a fixed day window, see [ordering](section-template-schema/ordering.md#timing) |
| Severity | `error`, `warning` |

## Computed types

### Question

What the finder returns for one active question.

| Field | Type | Notes |
| -- | -- | -- |
| `id` | Text | `<scope>.<name>`, such as `access.repo_url` |
| `scope` | Text | `plan` or a section id |
| `name` | Text |  |
| `description` | Text |  |
| `example` | Text, optional |  |
| `type` | VariableType |  |
| `choices` | List of Choice, optional | For `contact`, `role`, and `team` questions: each directory entry as a `value` (the id) and a `label` (the name), in directory order. `options` stays for `choice` questions |
| `options` | List of text, optional |  |
| `required` | Boolean |  |
| `multiple` | Boolean |  |
| `audience` | Audience |  |
| `answer` | Value, optional | The stored answer, if any |
| `declaredDefault` | Value, optional | The variable's `default`, shown to the person. Not a stored answer |
| `group` | Text | The key shared by questions with the same name, type, options, and `multiple` |
| `suggestions` | List of Suggestion |  |

### Suggestion

A group is the set of questions that share a `group` key, so a user interface can ask them once. It has no type of its own. A Suggestion is a Value offered for a question, with where it came from: a declared default, or an answer already given in the plan or a section.

| Field | Type | Notes |
| -- | -- | -- |
| `value` | Value |  |
| `source` | `default` or `given` |  |

### Reconciliation

What the record reconciler returns beside the record.

| Field | Type | Notes |
| -- | -- | -- |
| `newQuestions` | List of text | Question ids the new plan declares and the previous plan did not, sorted |
| `removedQuestions` | List of text | Question ids the previous plan declared and the new plan does not, sorted. Their answers stay as hidden answers |
| `answersToReview` | List of text | Question ids whose answer the new plan no longer accepts. The answer is removed, sorted |
| `newSections` | List of text | Section ids that are new, sorted |
| `removedSections` | List of text | Section ids that are gone, sorted |
| `changedOutcomes` | List of text | Section ids whose outcome was recorded against content that has since changed, sorted |

### Cleared

What the editor returns when it removes a directory entry.

| Field | Type | Notes |
| -- | -- | -- |
| `answers` | List of text | The question ids whose answers were removed or changed, sorted |
| `holders` | List of text | The holder places changed, such as `role:ops` or `team:platform/ops`, sorted |

### Blocker

| Field | Type | Notes |
| -- | -- | -- |
| `question` | Text | The id of an active, required planning question with no answer |

### Problem

A report about a plan or one value. It is a normal result, not a failure. A warning is advice, such as a role with only one person.

| Field | Type | Notes |
| -- | -- | -- |
| `code` | Text | A stable machine name, defined by the contract of the component that reports it |
| `severity` | Severity | An `error` stops an answer from being applied and blocks assembly. A `warning` is shown and never blocks |
| `message` | Text | For people |
| `question` | Text, optional | The question id, when one applies |
| `section` | Text, optional | The section id, when one applies |
| `detail` | Map of text to text, optional | Extra facts, such as the two sections in a timing conflict |

## Open questions

- Which problem codes exist? The contracts of the validators and the editor define them, and this page lists them once they exist.
