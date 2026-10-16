# Optional sections

Part of the [section template schema](README.md).

A section that can be switched off declares a placeholder named `enabled`. It is an ordinary [placeholder](placeholders.md), so it has a `description`, `required`, `default`, and the rest.

```yaml
placeholders:
  - name: enabled
    description: Include the design review?
    type: boolean
    default: true
```

- **A section with no `enabled` placeholder is always on** when its `includeIf` condition holds, or it has none. Nobody is asked.
- **The question** has the id `<section>.enabled`. The finder returns it before that section's other questions, in document order, see [ordering](ordering.md). The answer is stored in `values.sections.<id>.enabled`.
- **When the answer is false**, the section's other questions are inactive. They keep their stored values, are not asked, and do not block.
- **No answer.** The `default` applies. With neither an answer nor a default, the section is off. A `required` toggle with no answer is a blocker.
- **Restrictions.** `enabled` must be `boolean`. `unknown-yet` is not accepted for it, because a section is either in or out. `askIf` is not supported on it yet.
- **Conditions can test it.** A clause can name `<section>.enabled` like any other question. Only an explicit answer counts, not the default, so `is: true` is false until someone answers.
