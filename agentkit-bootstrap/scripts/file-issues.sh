#!/usr/bin/env bash
# Create the milestones and issues described in ISSUES.md.
#
# Requires the gh CLI, authenticated with issue-write access to the repository.
# Safe to re-run: existing milestones are reused, and issues whose exact title
# already exists are skipped rather than duplicated.
#
#   ./scripts/file-issues.sh [owner/repo]

set -euo pipefail

REPO="${1:-agenticgokit/agentkit}"
ISSUES_FILE="$(dirname "$0")/../ISSUES.md"

if ! command -v gh >/dev/null; then
  echo "gh CLI not found: https://cli.github.com" >&2
  exit 1
fi
if [ ! -f "$ISSUES_FILE" ]; then
  echo "cannot find $ISSUES_FILE" >&2
  exit 1
fi

echo "Target repository: $REPO"

# --- milestones -------------------------------------------------------------
declare -a MILESTONES=(
  "v0.1 Core contracts"
  "v0.2 Durable"
  "v0.3 Workflows"
  "v0.4 Multi-agent"
  "v0.5 Dev loop"
)

existing_milestones="$(gh api "repos/$REPO/milestones?state=all&per_page=100" --jq '.[].title')"
for m in "${MILESTONES[@]}"; do
  if grep -Fxq "$m" <<<"$existing_milestones"; then
    echo "milestone exists: $m"
  else
    gh api "repos/$REPO/milestones" -f title="$m" >/dev/null
    echo "milestone created: $m"
  fi
done

# --- issues -----------------------------------------------------------------
existing_titles="$(gh issue list --repo "$REPO" --state all --limit 500 --json title --jq '.[].title')"

python3 - "$ISSUES_FILE" <<'PY' > /tmp/agentkit-issues.tsv
import re, sys, json

raw = open(sys.argv[1]).read()
blocks = raw.split("\n===\n")[1:]  # everything before the first delimiter is prose
for b in blocks:
    m = re.match(r"\s*---\n(.*?)\n---\n(.*)", b, re.S)
    if not m:
        continue
    fm, body = m.group(1), m.group(2).strip()
    meta = {}
    for line in fm.splitlines():
        if ":" not in line:
            continue
        k, v = line.split(":", 1)
        meta[k.strip()] = v.strip().strip('"')
    labels = meta.get("labels", "[]").strip("[]")
    labels = ",".join(x.strip() for x in labels.split(",") if x.strip())
    print("\t".join([
        meta.get("title", ""),
        meta.get("milestone", ""),
        labels,
        json.dumps(body),
    ]))
PY

created=0 skipped=0
while IFS=$'\t' read -r title milestone labels body_json; do
  [ -z "$title" ] && continue
  if grep -Fxq "$title" <<<"$existing_titles"; then
    echo "skip (exists): $title"
    skipped=$((skipped + 1))
    continue
  fi

  body="$(python3 -c 'import json,sys; print(json.loads(sys.argv[1]))' "$body_json")"

  args=(--repo "$REPO" --title "$title" --body "$body")
  [ -n "$milestone" ] && args+=(--milestone "$milestone")
  if [ -n "$labels" ]; then
    IFS=',' read -ra ls <<<"$labels"
    for l in "${ls[@]}"; do
      # Create the label on demand; ignore "already exists".
      gh label create "$l" --repo "$REPO" >/dev/null 2>&1 || true
      args+=(--label "$l")
    done
  fi

  gh issue create "${args[@]}" >/dev/null
  echo "created: $title"
  created=$((created + 1))
done < /tmp/agentkit-issues.tsv

rm -f /tmp/agentkit-issues.tsv
echo
echo "done — $created created, $skipped skipped"
