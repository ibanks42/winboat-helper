#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
dist_dir="$repo_dir/dist"
version="${1:-0.1.0}"
target_os="linux"
target_arch="$(go env GOARCH)"
stage_dir="$dist_dir/winboat-helper-${version}-${target_os}-${target_arch}"
archive_path="$dist_dir/$(basename "$stage_dir").tar.xz"

mkdir -p "$dist_dir"
rm -rf "$stage_dir"
mkdir -p "$stage_dir"

fyne build --release --source-dir "$repo_dir" -o "$stage_dir/winboat-helper"

install -m 0644 "$repo_dir/Icon.png" "$stage_dir/Icon.png"
install -m 0755 "$repo_dir/install-user.sh" "$stage_dir/install-user.sh"
install -m 0755 "$repo_dir/uninstall-user.sh" "$stage_dir/uninstall-user.sh"
install -m 0644 "$repo_dir/packaging/README.txt" "$stage_dir/README.txt"

tar -C "$dist_dir" -cJf "$archive_path" "$(basename "$stage_dir")"

printf 'Built release at %s\n' "$stage_dir"
printf 'Packed archive at %s\n' "$archive_path"
