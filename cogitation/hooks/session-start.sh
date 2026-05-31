#!/usr/bin/env bash
# SessionStart hook for the cogitation plugin.
# Injects the full using-cogitation dispatcher so the workflow engages from the
# first message — without it, nothing tells Claude to route through the skills.
# Fires on startup|clear|compact (not --resume; that already has the context).
#
# JSON is built by python3 (robust escaping). If python3 is unavailable we emit
# nothing and exit 0 — degrade silently rather than print malformed JSON.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "${SCRIPT_DIR}/.." && pwd)}"
SKILL="${PLUGIN_ROOT}/skills/using-cogitation/SKILL.md"

command -v python3 >/dev/null 2>&1 || exit 0

python3 - "$SKILL" <<'PY'
import json, sys
try:
    body = open(sys.argv[1], encoding="utf-8").read()
except Exception:
    body = "Error reading using-cogitation skill"
intro = (
    "<EXTREMELY_IMPORTANT>\n"
    "This project uses the cogitation workflow. Below is the full content of your "
    "`using-cogitation` skill — your dispatcher. Follow it before responding to "
    "build/feature/bug/review requests. For every other skill, use the Skill tool.\n"
    "</EXTREMELY_IMPORTANT>\n\n"
)
out = {
    "hookSpecificOutput": {
        "hookEventName": "SessionStart",
        "additionalContext": intro + body,
    }
}
print(json.dumps(out))
PY
