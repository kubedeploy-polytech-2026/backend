#!/usr/bin/env bash
set -euo pipefail
version=0.75.0
case "$(uname -m)" in
  x86_64) architecture=64bit ;;
  aarch64|arm64) architecture=ARM64 ;;
  *) echo 'Unsupported Trivy architecture' >&2; exit 1 ;;
esac
work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT
archive="trivy_${version}_Linux-${architecture}.tar.gz"
gh release download "v$version" --repo aquasecurity/trivy \
  --pattern "$archive" --pattern "trivy_${version}_checksums.txt" --dir "$work_dir"
(cd "$work_dir" && sha256sum --check --ignore-missing "trivy_${version}_checksums.txt")
mkdir -p .tools/bin
tar -xzf "$work_dir/$archive" -C .tools/bin trivy
