#shellcheck shell="bash"
HOOK_DIR=${HOOK_DIR:-"/etc/odin/argocd/"}

run_hook_dir() {
  hook_name="$1"
  hook_path="${HOOK_DIR}/${hook_name}.d"

  # No dir => nothing to do
  [ -d "$hook_path" ] || return 0

  for f in "$hook_path"/*; do
    # Empty dir => glob doesn't expand; bail
    [ -e "$f" ] || break

    # Only run regular executable files
    [ -f "$f" ] || continue
    [ -x "$f" ] || continue

    echo "Running hook: $f" >&2
    "$f"
  done
}

function set_home() {
  if [ -z "${HOME:-}" ]; then
    export HOME=/home/argocd
  fi
}