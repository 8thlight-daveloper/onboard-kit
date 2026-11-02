# One command surface for the repo. Run `just` to list recipes.
set shell := ["bash", "-euo", "pipefail", "-c"]

md_paths := "docs AGENTS.md README.md CONTRIBUTING.md"
link_paths := "'docs/**/*.md' AGENTS.md README.md CONTRIBUTING.md"

default:
    @just --list

# Everything CI runs: format check, every linter, and the tests.
check: fmt-check lint-md lint-links lint-prose lint-go test

# Format Markdown in place.
fmt:
    mdformat {{ md_paths }}

# Fail if any Markdown needs formatting.
fmt-check:
    mdformat --check {{ md_paths }}

# Lint every kind of docs problem we check mechanically.
lint: lint-md lint-links lint-prose lint-go

# Markdown structure.
lint-md:
    markdownlint-cli2

# Relative links and heading anchors, offline.
lint-links:
    lychee --offline --include-fragments --no-progress {{ link_paths }}

# Voice rules: no em dashes, banned filler words, sentence case headings. Also reads prose inside YAML.
lint-prose:
    vale --minAlertLevel=warning {{ md_paths }} $(test -d templates && echo templates)

# Go formatting and vet. A no-op until the Go module exists.
lint-go:
    if [ -f go.mod ]; then \
      tracked_and_new_go_files=$(git ls-files --cached --others --exclude-standard '*.go'); \
      unformatted=$(echo "$tracked_and_new_go_files" | xargs gofmt -l); \
      test -z "$unformatted" && go vet $(just _go-packages); \
    else echo "no Go module yet"; fi

# Go tests. A no-op until the Go module exists.
test:
    if [ -f go.mod ]; then mkdir -p .scratch/tmp && TMPDIR="$PWD/.scratch/tmp" go test -race $(just _go-packages); else echo "no Go module yet"; fi

# Empty .scratch/tmp, which holds only disposable files.
clean-scratch:
    rm -rf .scratch/tmp && mkdir -p .scratch/tmp

# Remove worktrees under .scratch/worktrees that are clean and already merged into main.
clean-worktrees:
    for path in .scratch/worktrees/*/; do \
      [ -d "$path" ] || continue; \
      head=$(git -C "$path" rev-parse HEAD); \
      if git merge-base --is-ancestor "$head" main && [ -z "$(git -C "$path" status --porcelain)" ]; then \
        git worktree remove "$path"; \
      else echo "kept $path"; fi; \
    done; git worktree prune

# Build the binary. A no-op until the Go module exists.
build:
    if [ -f go.mod ]; then go build -o output/onboard-kit ./binary; else echo "no Go module yet"; fi

# Run the acceptance scenarios against the built binary. Skips until ./binary has Go files.
acceptance:
    if [ -f go.mod ]; then go test -race -count=1 -v ./acceptance/; else echo "no Go module yet"; fi

release-check *flags:
    go run ./tools/release/verify {{ flags }}

# Run the binary. A no-op until the Go module exists.
run *args:
    if [ -f go.mod ]; then go run ./binary {{ args }}; else echo "no Go module yet"; fi

# Build a separate manual-test binary, leaving the verified release alone.
demo-build:
    test -d output && test ! -L output
    if [ ! -e output/demo-bin ]; then mkdir -m 700 output/demo-bin; fi
    test -d output/demo-bin && test ! -L output/demo-bin && test ! -L output/demo-bin/onboard-kit
    GOMAXPROCS=2 GOMEMLIMIT=128MiB go build -p=1 -o output/demo-bin/onboard-kit ./binary

# Create an unanswered Acme-template plan, or retain the existing manual plan.
demo-init: demo-build
    go run ./tools/demo/run init

# Edit the shared organization plan, saving only the manual copy.
demo-plan: demo-build
    go run ./tools/demo/run plan

# Review and save one fictional recipient's preparation, named by ACCOUNT.
[positional-arguments]
demo-prepare account="": demo-build
    go run ./tools/demo/run prepare "$1"

# Export an invitation using an explicitly selected pre-access contact ID.
[positional-arguments]
demo-invite contact: demo-build
    go run ./tools/demo/run invite "$1"

# Open the invitation in the separate onboardee workspace.
demo-onboardee: demo-build
    go run ./tools/demo/run onboardee

# Edit answers or outcomes in the one registered demo session.
demo-progress: demo-build
    go run ./tools/demo/run progress

# View current preparation and returned outcomes through read-only JSON.
demo-results: demo-build
    go run ./tools/demo/run results

# Print every manual-test input/output path without changing data.
demo-status:
    go run ./tools/demo/run status

# Archive one stage, or all five, without deleting recoverable test work.
[positional-arguments]
demo-reset stage:
    go run ./tools/demo/run reset "$1"

# Link the dev shell's Go at .direnv/go, so an IDE can use a path that does not change. Point GoLand's GOROOT at .direnv/go/share/go.
ide:
    nix build .#go --out-link .direnv/go
    @echo "GOROOT for the IDE: $(pwd)/.direnv/go/share/go"

# Mutation-test one component against its contract, for example `just mutate plan/encoder`. It runs gomu over an overlay, so it never changes a source file and can run beside `just check`. One mutant at a time by default, with Go limited to two parallel jobs, so it is gentle on the machine. Survivors are gaps in the contract, except those a person accepted in tools/mutation/accepted/. Pass `--workers 2` to go faster.
mutate component *flags:
    mkdir -p .direnv/bin
    test -x .direnv/bin/gomu || GOBIN="$(pwd)/.direnv/bin" go install github.com/sivchari/gomu/cmd/gomu@v0.3.0
    go run ./tools/mutation/gomurun {{ flags }} {{ component }}

# The earlier mutation recipe, on go-mutesting. It rewrites source files in place and skips mutants that assign to a struct, slice, or map field, so it understates survivors. Kept for a while as a fallback. Do not run it beside `just check`.
mutate-old component:
    mkdir -p .direnv/bin
    GOBIN="$(pwd)/.direnv/bin" go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@v0.0.0-20261002081621-9ce5f731e321
    accepted="$(mktemp)"; if [ -f tools/mutation/accepted/{{ component }}.txt ]; then awk '!/^#/ && NF { print $1 }' tools/mutation/accepted/{{ component }}.txt > "$accepted"; fi; COMPONENT={{ component }} .direnv/bin/go-mutesting --blacklist="$accepted" --exec="$(pwd)/tools/mutation/exec.sh" --exec-timeout=120 ./internal/{{ component }}/; rm -f "$accepted" report.json

# Install the git hook that runs check before every commit.
hook:
    git config core.hooksPath .githooks
    @echo "git hook installed: .githooks/pre-commit"

# Go packages to check. A component's contract runs once the component's folder exists.
_go-packages:
    @go list ./... | while read -r package; do case "$package" in onboard-kit/contracts/harness/*) echo "$package";; onboard-kit/contracts/*) rel="${package#onboard-kit/contracts/}"; if [ -d "internal/$rel" ]; then echo "$package"; fi;; *) echo "$package";; esac; done
