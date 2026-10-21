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
    if [ -f go.mod ]; then test -z "$(gofmt -l .)" && go vet $(just _go-packages); else echo "no Go module yet"; fi

# Go tests. A no-op until the Go module exists.
test:
    if [ -f go.mod ]; then go test -race $(just _go-packages); else echo "no Go module yet"; fi

# Build the binary. A no-op until the Go module exists.
build:
    if [ -f go.mod ]; then go build -o output/onboard-kit ./binary; else echo "no Go module yet"; fi

# Run the acceptance scenarios against the built binary. Skips until ./binary has Go files.
acceptance:
    if [ -f go.mod ]; then go test -race -count=1 -v ./acceptance/; else echo "no Go module yet"; fi

# Run the binary. A no-op until the Go module exists.
run *args:
    if [ -f go.mod ]; then go run ./binary {{ args }}; else echo "no Go module yet"; fi

# Link the dev shell's Go at .direnv/go, so an IDE can use a path that does not change. Point GoLand's GOROOT at .direnv/go/share/go.
ide:
    nix build .#go --out-link .direnv/go
    @echo "GOROOT for the IDE: $(pwd)/.direnv/go/share/go"

# Mutation-test one component against its contract, for example `just mutate plan/encoder`. Survivors are gaps in the contract, except those a person accepted in tools/mutation/accepted/.
mutate component:
    mkdir -p .direnv/bin
    GOBIN="$(pwd)/.direnv/bin" go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@v0.0.0-20261002081621-9ce5f731e321
    accepted="$(mktemp)"; if [ -f tools/mutation/accepted/{{ component }}.txt ]; then awk '!/^#/ && NF { print $1 }' tools/mutation/accepted/{{ component }}.txt > "$accepted"; fi; COMPONENT={{ component }} .direnv/bin/go-mutesting --blacklist="$accepted" --exec="$(pwd)/tools/mutation/exec.sh" --exec-timeout=120 ./internal/{{ component }}/; rm -f "$accepted" report.json

# Try the onboarder walk on a new ACME plan with an empty directory, in output/try. Reuses the plan if it exists, so your answers stay. `just try-reset` starts over.
try-onboarder:
    mkdir -p output/try
    test -f output/try/new.obk.yaml || go run ./binary new --templates docs/reference/examples/acme/templates --template-set example --out output/try/new.obk.yaml
    go run ./binary walk --plan output/try/new.obk.yaml

# Try the onboarder walk on a copy of the finished ACME example, with its directory, in output/try. The example in docs is never changed.
try-acme:
    mkdir -p output/try
    test -f output/try/acme.obk.yaml || cp docs/reference/examples/acme/acme.obk.yaml output/try/acme.obk.yaml
    go run ./binary walk --plan output/try/acme.obk.yaml

# Try the newcomer's side: start from the ACME copy in output/try, with the newcomer's files in output/try/home.
try-onboardee:
    mkdir -p output/try
    test -f output/try/acme.obk.yaml || cp docs/reference/examples/acme/acme.obk.yaml output/try/acme.obk.yaml
    go run ./binary start --org-plan output/try/acme.obk.yaml --home output/try/home

# Delete everything `try-*` made.
try-reset:
    rm -rf output/try

# Install the git hook that runs check before every commit.
hook:
    git config core.hooksPath .githooks
    @echo "git hook installed: .githooks/pre-commit"

# Go packages to check. A component's contract runs once the component's folder exists.
_go-packages:
    @go list ./... | while read -r package; do case "$package" in onboard-kit/contracts/harness/*) echo "$package";; onboard-kit/contracts/*) rel="${package#onboard-kit/contracts/}"; if [ -d "internal/$rel" ]; then echo "$package"; fi;; *) echo "$package";; esac; done
