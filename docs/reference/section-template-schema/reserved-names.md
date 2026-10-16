# Reserved names

Part of the [section template schema](README.md).

Variable names start with a lowercase letter and continue with lowercase letters, digits, or underscores. Lint rejects the names below. The section validator rejects them in a plan too.

| Reserved | Why |
| -- | -- |
| `enabled` | Reserved for the [section toggle](optional-sections.md). It cannot be used as `$enabled` in text |
| `yes`, `no`, `y`, `n`, `true`, `false`, `on`, `off`, `null` | YAML can read these as booleans or null. As keys in `values` they would load as the wrong data, and [0030](../../decisions/0030-self-contained-plan-file.md) rejects them |

Template ids are reserved too. A template id is a lowercase slug, and lint rejects `plan` and the YAML words above as ids. `plan` is the scope of answers to plan variables, so a section named `plan` would make ids like `plan.team_state` ambiguous.

Names are reserved only where a real clash exists. Words such as `default` and `type` are not reserved, because a variable name appears as a value of `name:`, never as a field key. Adding a reserved name later can break existing templates, so each new one needs a schema version bump and a lint message that says what to rename.
