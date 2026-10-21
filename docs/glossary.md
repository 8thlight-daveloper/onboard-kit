# Glossary

Plain definitions of the terms used in the architecture docs and decisions. Add a term here before using it in a doc.

## The content

| Term | Meaning |
| -- | -- |
| Section template | The authored form of a section: one YAML file, written once. It has the same shape as a section, without `source`. It says when it applies, which questions it needs answered, and what text it produces. A new plan copies every standard section template in as a section. Often shortened to template. |
| Variable | A named, typed value with a scope, the plan or one section. A person fills it in once. It is written `$name` where text uses it, so it is also called a placeholder. It does not have to appear in any text: axis variables never do. The tool asks one question for each variable. |
| Answers | What the consultant says in response to the questions: team state, discipline, plan and section variables. |
| Section | The content unit in a plan. It has the same shape as a template, plus a record of which template it came from. Its variables stay until the document is assembled. A plan holds every standard template as a section, and only the active ones appear in the document. |
| Plan | The finished, standalone description of one organization's onboarding: sections, answers, question definitions, and a record of where it came from. It is saved as `<name>.obk.yaml` and carries everything needed to reopen it without the templates. |
| Plan copy | The local copy of the plan on a newcomer's machine, kept in their per-user state folder so the walk works offline. The tool replaces it on a refresh and never edits it for the newcomer's answers. See [0090](decisions/0090-plan-and-onboarding-record.md). |
| Onboarding record | The newcomer's own file, often called the record. It holds their `account`, a reference to the plan, the answers they gave, and the `outcomes` of the sections they tried. It has no sections and no variables. It is kept at the plan's `resultsPath`, often in a shared repository. See [0090](decisions/0090-plan-and-onboarding-record.md). |
| Account | The name that identifies a newcomer, such as `dnawara`. It is the permanent key in the path of the onboarding record. A mistake is fixed by moving the file. See [0085](decisions/0085-walking-sections-and-recording-outcomes.md). |
| Outcome | The result of one section for one newcomer: `verified`, `failed`, or `deferred`, with a time, the section's hash, and optional `note` and `differently` text. A section with no outcome has not been tried. See [0085](decisions/0085-walking-sections-and-recording-outcomes.md). |
| Differently | The optional text in an outcome that says what the newcomer had to do differently from the instructions. It is feedback for the onboarder, and it does not change the plan. |
| Report | A derived, read-only view over one or more onboarding records, computed on demand and stored nowhere. It reads the records and the plan. It is not the assembled document and not a screen. Its design is open, see [0095](decisions/0095-reports.md). |
| Start date | The day a newcomer starts, kept in their onboarding record as `startDate`. The onboarder gives it when making the invitation, and it defaults to today. A section's due date is the start date plus the last day of its timing window. See [0095](decisions/0095-reports.md). |
| Output style | The target format of the document, such as Markdown for GitHub. |
| Gap | A value not known yet. It shows in the document as a visible to-do with an owner and timing. |
| Provenance | A record of where a plan came from: the template set version and when it was made. |
| Default plan | The plan a new walkthrough starts from: the axis variables and every standard template as a section, with no answers. |
| Active section | A section whose `includeIf` condition holds, or that has none, and whose `enabled` setting is true. Only active sections appear in the document. |
| Axis variables | The plan variables that decide which sections are included, such as `team_state` and `discipline`. `discipline` is the kind of work the person will do, answered by the onboardee. They are data in the default plan, like every other variable. |

## The tool

| Term | Meaning |
| -- | -- |
| Tool | The `onboard-kit` binary, the one file consultants install. Everything below this heading is a part of it. |
| Scope | Where a variable and its answer attach: the plan, or one section. A question's id is `<scope>.<name>`, such as `access.repo_url`. |
| Audience | Who answers a question: the `onboarder`, who builds and maintains the plan, or the `onboardee`, the person being onboarded. |
| Question group | Questions with the same name, type, and options, which a user interface can ask once and apply to every member. |
| Question condition | An optional `askIf` on a variable. The question is asked only when other answers match. When it is false the question is skipped, keeps any stored answer, and does not block. See [conditions](reference/section-template-schema/conditions.md). |
| Section condition | An optional `includeIf` on a section or template, with the same syntax as `askIf`. The section is in the plan's document only when it holds, and it may name only plan variables. See [conditions](reference/section-template-schema/conditions.md#section-conditions). |
| Directory | The org plan's lists of contacts, roles, and teams. Answers to `contact`, `role`, and `team` questions are ids from it. See [0080](decisions/0080-directory.md). |
| Contact | A person or a shared contact, such as a help channel, in the directory. |
| Role | A named job in the directory. Its holders are contacts, and a team can name its own holders. A role with no holder is vacant. |
| Escalation | An ordered list of contact or role variables on a section or step, naming who to ask when it is blocked. The general path is its own section, "If you are stuck". See [escalation](reference/section-template-schema/escalation.md). |
| Requires | A section's list of the sections it needs first, by id. A section comes after what it requires, and its verification waits until each of them is `verified`. A section that requires an inactive section is inactive too. It replaces `after`. See [0085](decisions/0085-walking-sections-and-recording-outcomes.md). |
| Bootstrap section | A section marked `bootstrap: true`. A newcomer can do it with only a public copy of the tool and the plan file they were sent, such as getting access to the organization's repository. The walk shows these first. It may not require a section that is not a bootstrap section. |
| Plan location | The plan's `planUrl`, where its latest copy lives: a path or an HTTPS address. The onboarder sets it when the plan is created. A record keeps it so `start` can refresh without being told where to look. |
| Results path | The plan's `resultsPath`, a template such as `records/{account}/record.obk.yaml` that says where each newcomer's record is kept. `{account}` is its only variable. |
| First-steps document | The `first-steps.md` in an invitation, for a newcomer who has nothing installed. It holds the bootstrap sections and a block of standard help: a lost plan or record, a wrong account, where to get the tool, and who to ask. |
| Invitation | The zip file that the `invite` command makes for an account: the plan, a seeded onboarding record, and a first-steps document. `onboard-kit start --invitation` reads it directly. It carries the plan itself, because the plan's location may be behind access the newcomer does not have yet. |
| Hidden answer | A stored answer to a question that the plan no longer declares. A refresh never deletes it. Its owner removes it with a prune. Answers to questions that are only inactive are not hidden. See [0075](decisions/0075-refreshing-a-personal-plan.md). |
| Prune | The action that removes hidden answers. It shows what it would remove first, and removes only after a yes. |
| Reconcile | What a refresh does to a record under 0090: it checks the record against the new plan copy, removes answers the plan no longer accepts, lists hidden answers and new questions, and flags outcomes whose section changed. It never changes the plan. |
| Shared repository notice | The message `start` shows when the record's file sits in a git working tree: anyone who can read the repository can read the answers and notes. A person can silence it per repository, and the header keeps a "shared" marker. See [0085](decisions/0085-walking-sections-and-recording-outcomes.md). |
| Severity | Whether a problem is an `error`, which stops an answer and blocks assembly, or a `warning`, which is advice and never blocks. |
| Section switch | The `enabled` setting on a plan section, true by default. A person turns a section off. It is not a variable or a question. |
| Reserved name | A variable name or template id that lint rejects, because it has a special meaning or YAML reads it as another type. See [section template schema](reference/section-template-schema/reserved-names.md). |
| Suggestion | A value offered for a question: the distinct declared defaults in its group, and answers already given for that name in the plan or any section. |
| Plan editor | Component that returns a copy of a plan with one answer applied. The copy differs only in `answers`. The plan it was given is not altered, and the component holds no state. See [0010](decisions/0010-plan-editor.md). |
| Question finder | Component that finds the questions a plan asks, with their current answers, and which required ones still have no answer. It only reads. |
| Plan validator | Component that checks whether a plan's answers are acceptable: type, options, and the secrets check, and the relations between sections such as dependency cycles and timing conflicts. |
| Walkthrough | The interactive planning run where the tool asks questions and builds a plan. The later onboarding run, for newcomers, is a separate activity. |
| TUI | Terminal user interface. The screen-based walkthrough in a terminal. |
| Driver | Anything that operates the plan components: the TUI, a headless command, or a test. |
| Headless | Running without a person or a screen. A command reads input and writes output. |
| Assemble | The step that turns a plan into a document in an output style. It is deterministic. No model runs. The assembled document is transitional. The intended end state is the plan file and the onboarding helper. |

## The structure

| Term | Meaning |
| -- | -- |
| Component | One part of the tool with one responsibility: it takes one kind of input and produces one kind of output, such as the section selector. It is a part of the tool, not of a plan. Each component lives in its own folder and can be deleted and rebuilt. |
| Section selector | Component that finds the active sections of a plan: each section's `includeIf` and `enabled` setting, ordered by timing and dependencies. |
| Condition evaluator | Component that decides whether a condition holds, given the settled answers of the questions it names. Used for `askIf` and `includeIf`. |
| Settled answer | An answer the person decided for now: any stored value that is not empty and not `deferred`, such as `team_state: established` or a flag set to `false`. A person can still change it. Conditions test settled answers only. |
| Deferred answer | An answer the person chose to give later, stored as the word `deferred`, such as `tracker_grantor: deferred`. It satisfies a required question and becomes a visible gap in the assembled document. A question with nothing stored is unanswered, and that blocks. |
| Plan instantiator | Component that builds the default plan from the templates: every template copied in as a section, with the axis variables. |
| Section validator | Component that checks one section's own structure: variables, defaults, reserved names, the `$` rule, and check shapes. Used by the plan validator and by `onboard-kit lint`. |
| Template linter | Component that checks template text against a rules file: client names, URLs, email addresses, and broken links. Voice rules are checked by Vale, not here. Run by `onboard-kit lint`, and in CI on every change to templates. |
| Document assembler | Component that turns a plan into a document in an output style. It reads only the plan. |
| Section resolver | Component that fills the variables in a section from the plan's answers. |
| Section hasher | Component that computes the content hash of a section. |
| Plan migrator | Component that brings a plan's bytes from one schema version to the next, one step per version. |
| Plan decoder | Component that turns plan file bytes into a Plan, applying the migrator steps, keeping unknown fields, and reporting format guard violations. |
| Plan encoder | Component that turns a Plan into canonical bytes, so the same plan always gives the same bytes. |
| Domain model | The definitions of the data that several components pass to each other, such as Section and Plan. They live in one folder, `model`, and change slowly. See [the domain model](reference/domain-model.md). |
| Boundary | The edge of a component: what it accepts, what it returns, and what it hides. A component's contract describes its boundary. |
| Functional core | Every component except the drivers and the entry point. Its functions are pure and never alter an argument. See [0015](decisions/0015-functional-core-and-imperative-shell.md). |
| Imperative shell | The drivers and the entry point. They do input and output, hold the current plan, and call the core. |
| Pure function | A function whose result depends only on its arguments, and that changes nothing outside itself. |
| Immutable value | A value that is never altered after it is made. A change produces a new value. |
| Side effect | Something a component does outside the plan, such as reading or writing a file. Only drivers and the entry point have them. |
| Plan access | What a component does with a plan it was given: it reads fields, returns a changed copy that differs only in declared fields, or creates a plan from other inputs. No component alters the plan it was given. |
| Dependency | One component using another. The document assembler using the domain model is a dependency. |
| Dependency rules | The rules that limit which dependencies are allowed, so boundaries hold. Components have a place in a dependency order and may depend only on components earlier in it. The plan instantiator may depend on the section selector, and the section selector may never depend on the plan instantiator. |
| Dependency check | An automatic check that fails the build when a component has a dependency the rules do not allow. It works by reading the imports in the code. |
| Rate of change | How fast something changes, and what a change takes. Also known as pace layers, a term from Stewart Brand. The domain model changes slowly and needs an ADR. A component's implementation changes freely. A dependency points from the faster thing to the slower thing. |

## Keeping rebuilds safe

| Term | Meaning |
| -- | -- |
| Rebuild (regenerate) | Delete a component's folder and have it written again from its contract. |
| Contract | The written description of what one component must do, together with its fixtures and contract tests. It outlives any implementation. |
| Fixture | A saved example: a specific input and the output the component must produce for it. |
| Golden | A fixture whose expected output is a whole saved document, compared byte for byte. |
| Contract test | An automatic check that one component, given a fixture's input, produces its expected output. |
| Acceptance test | An automatic check that the finished tool, run as a user would run it, gives the right result. |
| Unit test | A test a component's own author writes for its internals. It is deleted with the component. |
| Protected files | Files a rebuild may not change without a person's approval: contracts, acceptance tests, and the domain model. |
| Walking skeleton | The thinnest end-to-end version of the whole tool, built first, with each component barely working. |

## Records

| Term | Meaning |
| -- | -- |
| ADR | Architecture decision record. A short document that states a decision and the reasons, for decisions the contracts and tests depend on. |
| Schema version | A whole number on a plan or template file. It rises when a change would break older files, and a migration reads the old version. |
| Source hash | A content hash recorded when a template is copied into a plan. Comparing it later shows whether the section was edited by hand, or the template changed upstream. |
| Template set version | The build's commit sha or tag, recorded in provenance. It says which build the templates came from. |
| Stack note | A row in [the stack folder](stack/README.md) recording the current language or tool. Cheap to change. |
| Eval | A check where a model judges the result. Not used for assembly. |
