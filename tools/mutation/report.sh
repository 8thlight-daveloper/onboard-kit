#!/usr/bin/env bash
set -eu

results="$(realpath "$1")"
cd "$(dirname "$0")"

threshold="${THRESHOLD_MINUTES:-120}"
floors="${FLOOR_DIR:-$PWD/floor}"

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
echo "| Component | Mutation time | Score | Floor | Against floor |"
echo "| -- | -- | -- | -- | -- |"
echo "$rows" | jq -r '.[] | [.component, .seconds, (.score // "none")] | @tsv' | while IFS=$'\t' read -r component seconds score; do
    minutes="$(awk -v s="$seconds" 'BEGIN { printf "%.1f", s / 60 }')"
    floor="none"
    verdict="no floor"
    if [ -f "$floors/$component.txt" ]; then
        floor="$(tr -d '[:space:]' < "$floors/$component.txt")"
        if [ "$score" = "none" ]; then
            verdict="NO SCORE"
        else
            verdict="$(awk -v s="$score" -v f="$floor" 'BEGIN { d = s - f; printf "%+.4f%s", d, (d < 0 ? "  BELOW" : "") }')"
        fi
    fi
    echo "| $component | \`$(./bar.sh "$minutes" "$threshold")\` | $score | $floor | $verdict |"
done
echo
echo '```mermaid'
echo "$rows" | jq -r --argjson threshold "$threshold" '
    (map((.seconds / 60 * 10 | round) / 10)) as $minutes
    | (([$threshold] + $minutes | max) * 1.1 | ceil) as $ceiling
    | ([300, length * 28 + 140] | max) as $height
    | "---",
      "config:",
      "  xyChart:",
      "    width: 900",
      "    height: \($height)",
      "---",
      "xychart-beta horizontal",
      "    title \"Mutation minutes, \(length) components\"",
      "    x-axis [\(map("\"" + .component + "\"") | join(", "))]",
      "    y-axis \"minutes\" 0 --> \($ceiling)",
      "    bar [\($minutes | join(", "))]",
      "    line [\(map($threshold) | join(", "))]"'
echo '```'
