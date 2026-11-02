#!/usr/bin/env bash
set -eu

component="$1"
log="$(realpath "$2")"
seconds="$3"
start="$PWD"
cd "$(dirname "$0")"
threshold="${THRESHOLD_MINUTES:-120}"

minutes="$(awk -v s="$seconds" 'BEGIN { printf "%.1f", s / 60 }')"
score="$(grep -oE '^The mutation score is [0-9.]+' "$log" | grep -oE '[0-9.]+$' || true)"

jq -n --arg component "$component" --argjson seconds "$seconds" --arg score "$score" \
    '{component: $component, seconds: $seconds, score: (if $score == "" then null else ($score | tonumber) end)}' > "$start/result.json"

echo "## $component"
echo
if [ -n "$score" ]; then
    grep -E '^The mutation score' "$log"
else
    echo "No score. The run did not finish."
fi
echo
echo '```'
echo "mutation time  $(./bar.sh "$minutes" "$threshold")"
echo '```'
echo
echo '```'
sed -n '/^Processing /,$p' "$log" | grep -vE '^(The mutation score|Report written)' | head -c 900000
echo '```'
