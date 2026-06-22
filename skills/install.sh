#!/usr/bin/env bash
# Symlink the asana-cli skill into the local agent skill directories so Claude
# Code (.claude) and other agents (.agents) can discover it.
#
# Usage:
#   ./skills/install.sh            # install into ~/.claude and ~/.agents
#   SKILL_SCOPE=project ./skills/install.sh   # install into ./.claude and ./.agents
#   ./skills/install.sh --uninstall
set -euo pipefail

SKILL_NAME="asana-cli"
# Resolve the directory this script lives in (the repo's skills/ dir).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$SCRIPT_DIR/$SKILL_NAME"

if [[ ! -d "$SRC" ]]; then
  echo "error: skill source not found at $SRC" >&2
  exit 1
fi

# Scope: 'home' (default) links into $HOME; 'project' links into the repo root.
SCOPE="${SKILL_SCOPE:-home}"
if [[ "$SCOPE" == "project" ]]; then
  BASE="$(cd "$SCRIPT_DIR/.." && pwd)"
else
  BASE="$HOME"
fi

TARGET_DIRS=(".claude/skills" ".agents/skills")

UNINSTALL=0
[[ "${1:-}" == "--uninstall" ]] && UNINSTALL=1

for rel in "${TARGET_DIRS[@]}"; do
  dest_dir="$BASE/$rel"
  dest="$dest_dir/$SKILL_NAME"

  if [[ "$UNINSTALL" == "1" ]]; then
    if [[ -L "$dest" ]]; then
      rm "$dest"
      echo "removed $dest"
    fi
    continue
  fi

  mkdir -p "$dest_dir"
  # Replace any existing symlink; refuse to clobber a real directory.
  if [[ -L "$dest" ]]; then
    rm "$dest"
  elif [[ -e "$dest" ]]; then
    echo "error: $dest exists and is not a symlink; leaving it alone" >&2
    continue
  fi
  ln -s "$SRC" "$dest"
  echo "linked $dest -> $SRC"
done
