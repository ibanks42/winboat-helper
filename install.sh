#!/usr/bin/env bash
set -euo pipefail

repo="${WINBOAT_HELPER_REPO:-ibanks42/winboat-helper}"
version="${1:-latest}"

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		printf 'Missing required command: %s\n' "$1" >&2
		exit 1
	fi
}

need_cmd curl
need_cmd tar
need_cmd mktemp

arch="$(uname -m)"
case "$arch" in
	x86_64|amd64) arch="amd64" ;;
	*)
		printf 'Unsupported architecture: %s\n' "$arch" >&2
		exit 1
		;;
esac

if [[ "$version" == "latest" ]]; then
	api_url="https://api.github.com/repos/$repo/releases/latest"
	version="$({ curl -fsSL "$api_url" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1; } || true)"
	if [[ -z "$version" ]]; then
		printf 'Could not determine latest release for %s\n' "$repo" >&2
		exit 1
	fi
fi

asset="winboat-helper-${version}-linux-${arch}.tar.xz"
download_url="https://github.com/${repo}/releases/download/${version}/${asset}"

work_dir="$(mktemp -d)"
cleanup() {
	rm -rf "$work_dir"
}
trap cleanup EXIT

archive_path="$work_dir/$asset"
printf 'Downloading %s\n' "$download_url"
curl -fL "$download_url" -o "$archive_path"

tar -C "$work_dir" -xf "$archive_path"

bundle_dir="$work_dir/winboat-helper-${version}-linux-${arch}"
if [[ ! -d "$bundle_dir" ]]; then
	printf 'Unexpected archive layout; missing %s\n' "$bundle_dir" >&2
	exit 1
fi

bash "$bundle_dir/install-user.sh"

printf 'Installed WinBoat Helper from release %s\n' "$version"
printf 'Launch with ~/.local/bin/winboat-helper\n'
printf 'Uninstall with ~/.local/lib/winboat-helper/uninstall-user.sh\n'
