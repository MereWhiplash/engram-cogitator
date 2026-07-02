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

# Workflow overlay: deltas-only manifest resolved into injected posture.
# EC_COG_CONFIG overrides the config path for side-effect-free tests.
COG_CONFIG="${EC_COG_CONFIG:-${CLAUDE_PROJECT_DIR:-$PWD}/.cogitation/config.json}"

python3 - "$SKILL" "$COG_CONFIG" <<'PY'
import json, sys

def read(path):
    try:
        return open(path, encoding="utf-8").read()
    except Exception:
        return None

skill_body = read(sys.argv[1]) or "Error reading using-cogitation skill"
cfg = {}
raw = read(sys.argv[2]) if len(sys.argv) > 2 else None
if raw:
    try:
        cfg = json.loads(raw)
    except Exception:
        cfg = {}
if not isinstance(cfg, dict):
    cfg = {}
workflow = cfg.get("workflow") or {}
skills = workflow.get("skills") or {}
customized = bool(workflow.get("customized"))

posture = []
for name, delta in sorted(skills.items()):
    if not isinstance(delta, dict):
        continue
    if delta.get("enabled") is False:
        posture.append(f"- `{name}` is DISABLED in this project — do not route to it.")
        continue
    rig = delta.get("rigidity")
    if rig and rig != "strict":
        posture.append(
            f"- `{name}` rigidity is **{rig}** in this project — "
            f"treat its discipline as {rig}, not mandatory."
        )

sections = []
if posture:
    sections.append(
        "## Active workflow posture (resolved from .cogitation/config.json)\n"
        "Apply these per-skill overrides when routing:\n" + "\n".join(posture)
    )
if not customized:
    sections.append(
        "## You are on the default cogitation profile\n"
        "This project has not been customized — the workflow is the opinionated "
        "default. To tailor which skills run and how strict they are, say "
        "**customise**; cogitation will walk you through it and not nag again."
    )

intro = (
    "<EXTREMELY_IMPORTANT>\n"
    "This project uses the cogitation workflow. Below is the full content of your "
    "`using-cogitation` skill — your dispatcher. Follow it before responding to "
    "build/feature/bug/review requests. For every other skill, use the Skill tool.\n"
    "</EXTREMELY_IMPORTANT>\n\n"
)
extra = ("\n\n" + "\n\n".join(sections)) if sections else ""
out = {"hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": intro + skill_body + extra,
}}
print(json.dumps(out))
PY
