# 0610. Start new plans from the embedded catalog only

Status: proposed.

Builds on [0320](0320-embedded-catalog-availability.md),
[0430](0430-reviewed-catalog-creation.md), and
[0530](0530-new-plan-from-core-set.md). Replaces the line in 0530 that says `new`
with `--templates` is unchanged.

## Context

`new`, `walk`, and `stdio` each take `--templates DIR`, a folder of section
files that the plan instantiator turns into a plan. The option comes from before
the catalog existed. [0005](0005-self-contained-install.md) says templates are
embedded in the binary, and the embedded set is now the catalog.

The folder path has three problems:

- `new` fails with `--templates is required` when given no options, though 0530
  says it should open a reviewed draft with a core set staged.
- A folder builds a plan with no authoring check. The five example templates
  under `docs/reference/examples/acme/templates` give a plan that `new` and the
  walk accept and `start` refuses, with two authoring errors:
  `authoring.default` on `access.onboarding_owner_role` and
  `authoring.unused-onboardee` on `plan.team`.
- The README quick start uses the example folder, while its first lines say a
  plan starts from the built-in section templates.

Nobody has asked for custom template folders, and no release has shipped.

## Decision

New plans start from the embedded catalog and from nothing else.

- `new` with no options opens the reviewed creation draft of 0430 with the core
  set staged, as 0530 describes.
- `new`, `walk`, and `stdio` lose `--templates` and `--template-set`. The
  template set recorded in a plan is the build's version, as
  [0030](0030-self-contained-plan-file.md) already says.
- Custom template folders are not supported. A later decision may add them. It
  must run the authoring validator before it writes a plan.
- The example templates stay as documentation of the format and as fixtures. They
  stop being a way to start a plan. Their two authoring errors get fixed, so the
  example is a valid set.

## Consequences

- Ten acceptance scenarios that pass `--templates` change to start from an
  existing plan file, such as `acme.obk.yaml`, or from a catalog-created plan.
  Scenarios live in `acceptance/`, so a person authorizes each change.
- The README quick start and command table change to match.
- `new` needs a build before the decision is whole: the core set draft of 0530 is
  not built.
- A plan made from a hand-written folder no longer comes out of the tool. An
  author who wants that edits a plan file or adds sections through the catalog.

## Not decided

1. **A missing plan file.** What `walk --plan FILE` and `stdio --plan FILE` do
   when `FILE` does not exist. They can refuse and point to `new`, or they can
   start from the core set.
2. **Where the format examples live.** Whether the example templates move from
   `docs/reference/examples` into `testdata` once they are fixtures only.
