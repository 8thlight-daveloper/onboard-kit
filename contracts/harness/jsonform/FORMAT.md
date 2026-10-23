# The JSON form of a plan

Contract tests keep their fixtures as plain data. This page says how a Plan is written down as JSON. The names are the domain model's, in camelCase. Only the code that reads and writes this form is specific to a language.

- An object has the fields of the type in the [domain model](../../../docs/reference/domain-model.md). An absent key means the zero value. An unknown key is an error, so a typo in a fixture fails loudly.
- A Value is a JSON string for Text, a boolean for Flag, an array of strings for List, and `{"deferred": true}` for Deferred. The Text `"deferred"` is a plain string and differs from Deferred.
- `unknownFields` on a plan or a section is an object of any JSON data. Integers stay integers and decimals stay decimals.
- A plan's `directory` is an object with `contacts`, `roles`, and `teams`, each a list of objects with the fields in the domain model. A team's `holders` is an object of role id to a list of contact ids. A plan has an optional `id` and an `origin` with `planId`, `source`, `contentHash`, and `refreshedOn`.
- An onboarding record is an object with `account`, `startDate`, `firstOpenedAt`, `plan` (`id`, `hash`, `url`), `answers`, `outcomes` (each with `section`, `outcome`, `on`, `sectionHash`, `note`, `differently`), and `unknownFields`. The optional `firstOpenedAt` is exact text, omitted when empty; see [0230](../../../docs/decisions/0230-first-use-record-compatibility.md). A plan also has optional `planUrl` and `resultsPath`. `Record` reads one and `RecordJSON` writes one.
- A section's `source` is an object with `template` and `hash`.
- A section's `requires` is an array of exact strings, independent of `after`.
  Reading and writing preserve its order and repeats; neither sorts it as a
  hash set. An absent or empty array reads as the zero list, and the writer
  omits nil and empty lists. Other types, including null and nonstring items,
  are fixture errors. See [0260](../../../docs/decisions/0260-prerequisites-and-action-order.md).
- A section's `bootstrap` is a boolean, false when absent. The harness writer
  omits false and writes true; a nonboolean fixture value is an error. This
  fixture representation is separate from YAML's empty/null optional-field
  rules. See [0250](../../../docs/decisions/0250-bootstrap-content-and-projection.md).
- A condition is `{"all": [{"question": "...", "answer": {"is": ...}}]}` or, for a section clause, `{"all": [{"section": "access", "enabled": true}]}`. A section has an optional boolean `enabled`.

The harness reads a whole plan with `Plan`, and one part of a plan with `Section`, `Condition`, `Value`, and `Answers`. The forms of the parts are the ones above. `JSON` and `SectionJSON` write a plan and a section back, so a failing test can show the values side by side.
