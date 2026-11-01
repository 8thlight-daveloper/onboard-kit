#!/usr/bin/env bash
set -eu
cd "$(git rev-parse --show-toplevel)"

base="${1:-}"

components() {
    for folder in internal/*/ internal/*/*/; do
        name="${folder#internal/}"
        name="${name%/}"
        [ -d "contracts/$name" ] || continue
        ls "$folder"*.go > /dev/null 2>&1 || continue
        echo "$name"
    done
}

everything() {
    components | jq -R . | jq -sc .
}

if [ -z "$base" ] || ! git cat-file -e "$base^{commit}" 2> /dev/null; then
    everything
    exit 0
fi

shared="$(git diff --name-only "$base" HEAD -- internal/model go.mod go.sum tools/mutation flake.nix flake.lock justfile .github/workflows/mutate.yml)"
if [ -n "$shared" ]; then
    everything
    exit 0
fi

components | while read -r name; do
    if [ -n "$(git diff --name-only "$base" HEAD -- "internal/$name" "contracts/$name")" ]; then
        echo "$name"
    fi
done | jq -R . | jq -sc .
