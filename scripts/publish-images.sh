#!/usr/bin/env bash
set -euo pipefail
: "${RELEASE_TAG:?Set RELEASE_TAG}"
: "${REVISION:?Set REVISION}"
: "${GITHUB_REPOSITORY:?Set GITHUB_REPOSITORY}"
: "${GITHUB_OUTPUT:?Set GITHUB_OUTPUT}"
image="ghcr.io/${GITHUB_REPOSITORY,,}"
mkdir -p dist
for arch in amd64 arm64; do
  docker load --input "candidates/$arch/image.tar"
  docker tag kubedeploy-candidate:build "$image:$RELEASE_TAG-$arch"
  docker push "$image:$RELEASE_TAG-$arch"
  digest="$(docker buildx imagetools inspect "$image:$RELEASE_TAG-$arch" --format '{{json .}}' | jq -er '.manifest.digest')"
  echo "$arch=$digest" >> "$GITHUB_OUTPUT"
  cp "candidates/$arch/sbom.cdx.json" "dist/sbom-$arch.cdx.json"
  cp "candidates/$arch/trivy.json" "dist/trivy-$arch.json"
done
docker buildx imagetools create --tag "$image:$RELEASE_TAG" --tag "$image:sha-$REVISION" \
  "$image:$RELEASE_TAG-amd64" "$image:$RELEASE_TAG-arm64"
digest="$(docker buildx imagetools inspect "$image:$RELEASE_TAG" --format '{{json .}}' | jq -er '.manifest.digest')"
sha_digest="$(docker buildx imagetools inspect "$image:sha-$REVISION" --format '{{json .}}' | jq -er '.manifest.digest')"
[[ "$digest" == "$sha_digest" ]]
echo "image=$image" >> "$GITHUB_OUTPUT"
echo "digest=$digest" >> "$GITHUB_OUTPUT"
jq -n --arg repository "$GITHUB_REPOSITORY" --arg tag "$RELEASE_TAG" \
  --arg revision "$REVISION" --arg image "$image" --arg digest "$digest" \
  '{schema_version: "1.0.0", repository: $repository, tag: $tag, revision: $revision,
    image: $image, digest: $digest, platforms: ["linux/amd64", "linux/arm64"]}' > dist/release.json
