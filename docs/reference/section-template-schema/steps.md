# Steps

Part of the [section template schema](README.md).

Steps are plain YAML with a vocabulary borrowed from CI files. They are optional. The MVP renders them as a checklist. A later walkthrough runtime acts on them, under the trust rules in [#29](https://github.com/8thlight-daveloper/onboard-kit/issues/29).

| Field | Meaning |
| -- | -- |
| `id`, `name` | Stable identifier and display name |
| `notes` | Optional Markdown prose |
| `owner`, `timing` | Who acts and when. A step inherits `timing` from its section, see [ordering](ordering.md) |
| `run` | Command as an argument list, no shell. It is always a list. A variable of type `command` is written as the list item `$name`, as in `run: [$test_command]`, and its list of arguments replaces that item, and a bare `run: $name` or `run: make test` is an error |
| `open` | A URL or file to open |
| `directory` | Directory to run `run` in |
| `env` | Environment variables the command runs with. See below |
| `escalation` | Who to ask when the step is blocked. See [escalation](escalation.md). Inherits the section's when absent |
| `checks` | A list of checks. See below |

A step with `run` and no `checks` passes when the command exits with code 0.

## Checks

A check says how to tell that a step worked. `checks` is a list. The step passes when every check passes. Each check has exactly one of these keys:

| Key | Value | Passes when |
| -- | -- | -- |
| `run` | A command as an argument list, no shell | The command exits with code 0 |
| `ask` | A yes or no question | The person answers yes |
| `env` | The name of an environment variable | The variable is set and not empty. The value is never read into the plan or shown |

```yaml
checks:
  - run: [git, ls-remote, $repo_url]
  - ask: Can you open the CI dashboard?
  - env: STAGING_API_TOKEN
```

Lint rejects a check with none of the keys or with more than one.

## Environment variables

Commands are argument lists and no shell parses them, so a variable such as `$HOME` inside `run` is never expanded. Use these instead.

- **A step needs a variable the person sets themselves**, such as an API token. Add an `env` check with its name, and say in `notes` where to get the value. The template and the plan hold only the name. The value is a secret and stays out of both.
- **A command needs a variable with a known value.** Give the step an `env` map. Values may use `$name` variables. The secrets check runs on them like any other value.

```yaml
steps:
  - id: run-tests
    name: Run the tests
    run: [$test_command]
    env:
      APP_ENV: test
      DATABASE_URL: $test_database_url
    checks:
      - env: STAGING_API_TOKEN
```

- Variable names are uppercase letters, digits, and underscores, and start with a letter.
- In prose, such as `notes` or `body`, a variable is written with the escape described under [variables](variables.md#dollar-signs), for example `$$STAGING_API_TOKEN`.
