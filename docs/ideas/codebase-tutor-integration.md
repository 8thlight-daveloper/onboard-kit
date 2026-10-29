# Codebase tutor integration

Status: idea, 2026-10-29.

## Idea

`repo-tutor` is a separate program that reads a repository and teaches it to a
new developer, with lessons, quizzes and exams that cite real files and lines. It
fits the weakest section in the catalog, `codebase-orientation`, which can only
ask the newcomer to work it out alone. [0550](../decisions/0550-developer-onboarding-catalog-additions.md)
adds a `codebase-tutor` entry. This note holds the larger integration ideas that
the ADR leaves out.

## Three levels

1. **A section that points to it.** Steps to install and run the tutor against
   the team's repository. The onboarder supplies the repository, the starting
   goal, and the agent tool. This needs only content. It is the level 0550 adopts.
2. **The kit runs it.** The step launches the tutor for the newcomer. This needs
   the later package that runs commands, placed in
   [0020](../decisions/0020-binary-structure-and-regenerative-components.md). It
   has not been checked whether that package exists yet.
3. **A progress bridge.** The newcomer chooses to export tutor progress, such as
   topics read and an exam score, and the kit records it as an outcome under
   [0290](../decisions/0290-explicit-outcome-recording.md). It must be the
   newcomer's explicit act and never an automatic report to the team, because
   quiz scores are sensitive.

## Constraints

- **License boundary.** The tutor is GPL version 3 or later and the kit is MIT.
  Running it as a separate program touches neither license. Embedding or
  linking it would put the kit binary under the GPL and break
  [0005](../decisions/0005-self-contained-install.md).
- **Model access.** The tutor sends excerpts of the repository to a model through
  the person's installed agent tool. An organization may not allow that, so the
  entry requires `ai-tool-policy`.
- **Delivery.** The tutor is nearly ready for a first release, as a download,
  through Nix, and through Homebrew. The install step depends on that release.
- **Generic fallback.** Keep the tool-neutral `codebase-orientation` for
  organizations that use something else.

## What a scan of this repository showed

The tutor's offline scan found 112 packages, 200 documents, and four entry
points. It orders topics from the dependency graph, so contract packages come
before the code they describe. The most complex files were in the invitation
starter and reviewer, the session editor, and the JSON form test code. Only the
scan was run. No lesson or quiz was run, so the quality of the teaching is
unchecked.

## Open questions

- What a progress export looks like, and who sees it.
- Whether the kit should show the tutor's next recommended step as the step's
  own status.
