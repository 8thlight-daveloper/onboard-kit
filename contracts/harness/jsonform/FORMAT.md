# The JSON form of a plan

Contract tests keep their fixtures as plain data. This page says how a Plan is written down as JSON. The names are the domain model's, in camelCase. Only the code that reads and writes this form is specific to a language.

- An object has the fields of the type in the [domain model](../../../docs/reference/domain-model.md). An absent key means the zero value. An unknown key is an error, so a typo in a fixture fails loudly.
- A Value is a JSON string for Text, a boolean for Flag, an array of strings for List, and `{"unknown": true}` for Unknown. The Text `"unknown-yet"` is a plain string and differs from Unknown.
- `unknownFields` on a plan or a section is an object of any JSON data. Integers stay integers and decimals stay decimals.
- A section's `source` is an object with `template` and `hash`.
- A condition is `{"all": [{"question": "...", "answer": {"is": ...}}]}`.
