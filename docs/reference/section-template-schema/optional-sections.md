# Optional sections

Part of the [section template schema](README.md). The decision is [0050](../../decisions/0050-section-switch.md).

Any section in a plan can be switched off. The switch is a setting on the plan section, not a variable and not a question.

```yaml
sections:
  - id: review-design-system
    enabled: false
```

- **On by default.** `enabled` is optional and true by default. A template never declares it, and a section added to a plan is on.
- **Only a person turns a section off.** The editor offers the switch as an action on a section. It is never asked, never blocks, and cannot be deferred.
- **A template cannot declare a variable named `enabled`.** The name is reserved, see [reserved names](reserved-names.md).
- **Sub-sections follow.** Turning a section off turns off its sub-sections too, see [ordering](ordering.md#sub-sections). A sub-section can be turned off alone.
- **When the section is off**, it is not active, so it is left out of the document and its questions are not asked. Its answers stay stored, so turning it on again loses nothing.
- **The content hash leaves it out.** The switch is the plan's own setting. Flipping it does not change what a template says.
- **Conditions can test it.** A section clause tests another section's switch, see [conditions](conditions.md#section-clause).
