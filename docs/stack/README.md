# Stack

The current language and tooling. These are implementation choices, cheap to change, so they live here and not in an ADR. Anything the contracts or acceptance tests depend on belongs in an [ADR](../decisions/) instead.

Every choice must satisfy [0005](../decisions/0005-self-contained-install.md): one self-contained file, no secondary installs.

## Topics

| File | Covers |
| -- | -- |
| [language.md](language.md) | Language and terminal toolkit |
| [testing.md](testing.md) | Test runner and its settings, the acceptance runner, the mutation tool |
| [yaml.md](yaml.md) | The YAML library the plan decoder reads files with |
| [dependency-check.md](dependency-check.md) | The extractor that checks dependency rules, and the packages core components may not use |
| [code-layout.md](code-layout.md) | The folders and file names that map the structure in [0020](../decisions/0020-binary-structure-and-regenerative-components.md) to the current language |
| [dev-tooling.md](dev-tooling.md) | The task runner, dev shell, Markdown format and lint, link check, voice rules, and CI |
| [release.md](release.md) | Release build, and the Homebrew and Nix channel files |

Each topic file has the same three sections: Current, Retired, and Considered and not chosen.

## Changing a row

0. Record the options you weighed in Considered and not chosen, with the date and the evidence.
1. Move the old row from Current to Retired. Fill in Until, Replaced by, and Why it changed.
2. Add or edit the Current row with a new Since date.
3. Run the contract and acceptance suites. A stack change that needs a test change is probably an ADR, not a stack note.

Git holds the exact diffs. The Retired table holds the reasons, so nobody has to dig through history to learn why something was replaced.
