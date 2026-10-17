#!/usr/bin/env bash
set -u
cd "$(git rev-parse --show-toplevel)"
mv "$MUTATE_ORIGINAL" "$MUTATE_ORIGINAL.tmp"
cp "$MUTATE_CHANGED" "$MUTATE_ORIGINAL"
restore() { mv "$MUTATE_ORIGINAL.tmp" "$MUTATE_ORIGINAL"; }
trap restore EXIT
if ! go build "./internal/$COMPONENT/" >/dev/null 2>&1; then
  exit 2
fi
if go test -count=1 -timeout 60s "./contracts/$COMPONENT/" >/dev/null 2>&1; then
  diff -u "$MUTATE_ORIGINAL.tmp" "$MUTATE_CHANGED"
  exit 1
fi
exit 0
