# The first-steps document

The `first-steps.md` file in an invitation. See [0090](../decisions/0090-plan-and-onboarding-record.md). It is for a newcomer who has nothing installed and may have no access yet. The assembler writes it for the onboardee, deterministically, with no model. A blank in an onboardee answer shows as `<name>`, as in any assembled document.

Each part has a source. **Tool** means fixed wording that the tool supplies. **Plan** means content that comes from the plan or its directory.

| # | Part | What it says | Source |
| -- | -- | -- | -- |
| 1 | Welcome | Who sent this, what onboarding the organization expects, and what the newcomer will end up with. Names the onboarding owner | Plan |
| 2 | What is in the invitation | The plan, the newcomer's record, and this document, and what each is for. It says that the first question is whether the account is theirs, and what to do if it is not | Tool |
| 3 | Get the tool | Where to download `onboard-kit` and how to check that it runs. The address is part of the tool | Tool |
| 4 | Start | The command that opens the invitation. It asks first whether the account is theirs, and writes nothing until they say yes. Then where it saves their files | Tool |
| 5 | Do these first | The bootstrap sections of the plan, as the assembler writes them, in order | Plan |
| 6 | Connect to the shared repository | After access works: where to clone, that the record goes at the plan's results path, and that running `start` inside the checkout finds it. The git commands the tool prints for adding, committing, and pushing the record | Tool and plan |
| 7 | Get the latest plan | How to ask the tool to check for the latest plan once the repository can be reached | Tool |
| 8 | Day to day | How to come back, how to mark a section verified, failed, or deferred, and what each means | Tool |
| 9 | Who can see your record | In a shared repository, anyone who can read it can read the answers and notes | Tool |
| 10 | If something goes wrong | A lost plan, a lost record, a wrong account, a tool that will not run, a plan the tool says is newer than it understands, and access that does not work. Each says what to do and who to ask | Tool and plan |
| 11 | Who to ask | The onboarding owner and the general contacts, resolved from the directory, with vacant roles shown as vacant | Plan |
| 12 | About this plan | The plan's id and date, so a newcomer and an onboarder can tell which plan a question is about | Plan |

## Rules

- The standard wording is embedded in the tool and filled in from the plan: the owner, the contacts, the plan location, and the results path.
- A part whose source data is missing says so plainly. It never shows an empty name, as the directory rules already require.
- The document holds no secrets and no personal data beyond what the plan already holds.
- Parts 1 and 5 to 7 change with the plan. The rest are the same for every plan.
