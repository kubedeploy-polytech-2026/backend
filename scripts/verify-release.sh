#!/usr/bin/env bash
set -euo pipefail
: "${RELEASE_TAG:?Set RELEASE_TAG}"
[[ "$RELEASE_TAG" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || {
  echo 'Release tag must match vMAJOR.MINOR.PATCH without leading zeroes' >&2; exit 1;
}
git fetch --quiet --no-tags origin main:refs/remotes/origin/main
revision="$(git rev-parse HEAD)"
[[ "$revision" == "$(git rev-parse origin/main)" ]] || {
  echo 'Release tag must point to the current origin/main commit' >&2; exit 1;
}
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "revision=$revision" >> "$GITHUB_OUTPUT"
fi
