#!/usr/bin/env bash
set -eu

results="$(realpath "$1")"
cd "$(dirname "$0")"

threshold="${THRESHOLD_MINUTES:-120}"
shown=15

rows="$(find "$results" -name result.json -exec cat {} + | jq -s 'sort_by(-.seconds)')"
count="$(echo "$rows" | jq 'length')"
if [ "$count" = "0" ]; then
    echo "No result.json under $results" >&2
    exit 1
fi

echo "## Mutation run"
echo
echo "$count components. Time threshold: $threshold min."
echo
echo "| Component | Mutation time | Score |"
echo "| -- | -- | -- |"
echo "$rows" | jq -r '.[] | [.component, .seconds, (.score // "none")] | @tsv' | while IFS=$'\t' read -r component seconds score; do
    minutes="$(awk -v s="$seconds" 'BEGIN { printf "%.1f", s / 60 }')"
    echo "| $component | \`$(./bar.sh "$minutes" "$threshold")\` | $score |"
done
echo
echo '```mermaid'
echo "$rows" | jq -r --argjson threshold "$threshold" --argjson shown "$shown" '
    .[:$shown] as $top
    | ($top | map((.seconds / 60 * 10 | round) / 10)) as $minutes
    | (([$threshold] + $minutes | max) * 1.1 | ceil) as $ceiling
    | "xychart-beta",
      "    title \"Mutation minutes, slowest \($top | length)\"",
      "    x-axis [\($top | map("\"" + .component + "\"") | join(", "))]",
      "    y-axis \"minutes\" 0 --> \($ceiling)",
      "    line [\($minutes | join(", "))]",
      "    line [\($top | map($threshold) | join(", "))]"'
echo '```'
