# One command surface for the repo. Run `just` to list recipes.
set shell := ["bash", "-euo", "pipefail", "-c"]

md_paths := "docs AGENTS.md README.md CONTRIBUTING.md"
link_paths := "'docs/**/*.md' AGENTS.md README.md CONTRIBUTING.md"

default:
    @just --list

# Everything CI runs: format check, every linter, and the tests.
check: fmt-check lint-md lint-links lint-prose test

# Format Markdown in place.
fmt:
    mdformat {{ md_paths }}

# Fail if any Markdown needs formatting.
fmt-check:
    mdformat --check {{ md_paths }}

# Lint every kind of docs problem we check mechanically.
lint: lint-md lint-links lint-prose

# Markdown structure.
lint-md:
    markdownlint-cli2

# Relative links and heading anchors, offline.
lint-links:
    lychee --offline --include-fragments --no-progress {{ link_paths }}

# Voice rules: no em dashes, banned filler words, sentence case headings.
lint-prose:
    vale --minAlertLevel=warning {{ md_paths }}

# Go tests. A no-op until the Go module exists.
test:
    if [ -f go.mod ]; then go test -race ./...; else echo "no Go module yet"; fi

# Build the binary. A no-op until the Go module exists.
build:
    if [ -f go.mod ]; then go build -o output/onboard-kit ./binary; else echo "no Go module yet"; fi

# Run the binary. A no-op until the Go module exists.
run *args:
    if [ -f go.mod ]; then go run ./binary {{ args }}; else echo "no Go module yet"; fi

# Install the git hook that runs check before every commit.
hook:
    git config core.hooksPath .githooks
    @echo "git hook installed: .githooks/pre-commit"
