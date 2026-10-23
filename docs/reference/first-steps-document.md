# The first-steps document

The `first-steps.md` file in an invitation. See [0090](../decisions/0090-plan-and-onboarding-record.md) and [0280](../decisions/0280-first-steps-help-and-sources.md). It is for a newcomer who has nothing installed and may have no access yet. The first steps assembler supplies standard help around the real document assembler's scheduled bootstrap fragment, deterministically, with no model. A blank in an onboardee answer shows as `<name>`, as in any assembled document. The contract is a preview boundary; successful rendering does not establish invitation readiness or an implemented import/start flow.

Each part has a source. **Tool** means fixed wording that the tool supplies. **Plan** means content that comes from the plan or its directory.

[0140](../decisions/0140-person-specific-preparation.md) adds a separate
onboarder-owned preparation file to the invitation. Explain its purpose and
ownership alongside the plan and the newcomer's record. The
[guided preparation flow](../decisions/0150-onboarding-library-and-preparation-flow.md#guide-invitation-preparation)
uses the same preparation data as later updates. Preparation editing and
[explicit export](invitation-export.md) are implemented. A
[recipient guide](preparation-command.md#review-before-saving) stages details
before saving; guided arrangements remain planned. The pure
[preparation reviewer](../decisions/0330-preparation-review.md)
combines validation, blockers and preview; it is not a new command, readiness gate
or save operation.

| # | Part | What it says | Source |
| -- | -- | -- | -- |
| 1 | Welcome | The organization, the purpose of onboarding, and the onboarding owner | Plan and preparation |
| 2 | What is in the invitation | The plan, the onboarder's preparation, the newcomer's record, and this document, and what each is for. It says that the first question is whether the account is theirs, and what to do if it is not | Tool |
| 3 | Get the tool | How to obtain the supported standalone tool and setup help. No invented distribution address | Tool |
| 4 | Start | Supported opening instructions or a clear help route while startup is pending. Account and proposed destinations must be confirmed before import writes | Tool, explicit account and locator paths |
| 5 | Do these first | The bootstrap sections of the plan, as the assembler writes them, in order | Plan |
| 6 | Connect to the shared repository | After access works: the plan location, record destination and supported sharing help. The guide runs no Git operation | Tool and plan |
| 7 | Get the latest plan | Supported update help after access works, preserving saved work | Tool |
| 8 | Day to day | Supported return/progress help and the meaning of verified, failed and deferred | Tool |
| 9 | Who can see your record | In a shared repository, anyone who can read it can read the answers and notes | Tool |
| 10 | If something goes wrong | A lost plan, a lost record, a wrong account, a tool that will not run, a plan the tool says is newer than it understands, and access that does not work. Each says what to do and who to ask | Tool and plan |
| 11 | Who to ask | The preparation-selected owner and explicit onboarder-selected directory help references, with vacant roles shown as vacant | Preparation, explicit help selection and plan directory |
| 12 | About this plan | The plan's id and date, so a newcomer and an onboarder can tell which plan a question is about | Plan |

## Rules

- The standard wording is embedded in the tool. Until supported runtime commands and distribution are established, it uses truthful owner-guided help, not invented executable instructions.
- Render bootstrap activities from the complete original plan through the real AssembleBootstrap operation. Preserve its exact fragment, action order, context and warnings; refuse a structural error without a partial guide.
- Preparation supplies the owner. General help references require explicit onboarder selection; the renderer does not guess a support-template binding or pick the first directory contact.
- Archive names follow 0240. Proposed destinations come from the locator and remain distinct from archive entry names. Displaying them authorizes no extraction or write.
- A part whose source data is missing says so plainly. It never shows an empty name, as the directory rules already require.
- Standard help includes only the needed invitation account, paths, owner and selected business help routes. It does not copy recipient name/email, planned dates, arrangements, notes or unknown preparation data. Business routes remain visible PII, not private by file ownership.
- Wrapper values are literal display text, not Markdown structure or executable instructions. Bootstrap content keeps its existing rendering rules.
- Reading or rendering this document records no first use. First use is not an employment start date, an accepted default or completion of any activity.
