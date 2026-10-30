#!/usr/bin/env bash
set -euo pipefail
: "${IMAGE:=kubedeploy-candidate:build}"
name="kubedeploy-image-check-$$-$RANDOM"
work_dir="$(mktemp -d)"
cleanup() {
  result=$?
  if (( result != 0 )); then docker logs "$name" >&2 2>/dev/null || true; fi
  docker rm -f "$name" >/dev/null 2>&1 || true
  rm -rf "$work_dir"
}
trap cleanup EXIT
docker run --detach --name "$name" --read-only --cap-drop ALL \
  --security-opt no-new-privileges  --publish 127.0.0.1::8080 "$IMAGE" >/dev/null
binding="$(docker port "$name" 8080/tcp)"
url="http://127.0.0.1:${binding##*:}"
ready=false
for ((attempt=0; attempt<30; attempt++)); do
  if curl --fail --silent --max-time 2 "$url/health" >"$work_dir/health"; then
    ready=true
    break
  fi
  sleep 1
done
[[ "$ready" == true ]]
jq -e '.status == "ok"' "$work_dir/health" >/dev/null
curl --fail --silent --show-error --max-time 5 "$url/api/version" | \
  jq -e '.service == "kubedeploy-backend" and .version == "0.2.0"' >/dev/null
for path in /readyz /api/notes; do
  status="$(curl --silent --max-time 5 --output /dev/null --write-out '%{http_code}' "$url$path")"
  [[ "$status" == 503 ]]
done
echo 'Container smoke passed with read-only filesystem, no-new-privileges and dropped capabilities'
