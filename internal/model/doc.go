// Package model holds the types that components share. It holds types only,
// with no behavior, and imports nothing at all. It changes only by ADR,
// because the stored types are the plan file format in memory. See
// docs/reference/domain-model.md.
//
// Optional fields use the zero value for absent. A zero Value has no kind, so
// it means no value was given. Defaults that the reference names, such as the
// string type or the planning phase, are applied by the components that read
// these types, not stored here.
package model
