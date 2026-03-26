#!/usr/bin/env bash
set -euo pipefail

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
bin_name="winboat-helper"
app_name="WinBoat Helper"
source_bin="$script_dir/$bin_name"
source_icon="$script_dir/Icon.png"

target_bin_dir="${HOME}/.local/bin"
target_bin="$target_bin_dir/$bin_name"
target_icon_dir="${HOME}/.local/share/icons/hicolor/256x256/apps"
target_icon="$target_icon_dir/$bin_name.png"
desktop_dir="${HOME}/.local/share/applications"
desktop_file="$desktop_dir/$bin_name.desktop"
support_dir="${HOME}/.local/lib/$bin_name"
installed_uninstall="$support_dir/uninstall-user.sh"

if [[ ! -f "$source_bin" ]]; then
	printf 'Missing %s next to installer. Run the release build first.\n' "$bin_name" >&2
	exit 1
fi

if [[ ! -f "$source_icon" ]]; then
	printf 'Missing Icon.png next to installer.\n' >&2
	exit 1
fi

mkdir -p "$target_bin_dir" "$target_icon_dir" "$desktop_dir" "$support_dir"

install -m 0755 "$source_bin" "$target_bin"
install -m 0644 "$source_icon" "$target_icon"
install -m 0755 "$script_dir/uninstall-user.sh" "$installed_uninstall"

cat >"$desktop_file" <<EOF
[Desktop Entry]
Type=Application
Version=1.0
Name=$app_name
Comment=Launch the WinBoat Helper tray app
Exec=$target_bin
Icon=$target_icon
Terminal=false
Categories=Utility;Network;
Keywords=winboat;rdp;docker;windows;
StartupNotify=false
EOF

if command -v update-desktop-database >/dev/null 2>&1; then
	update-desktop-database "$desktop_dir" >/dev/null 2>&1 || true
fi

printf 'Installed %s to %s\n' "$app_name" "$target_bin"
printf 'Launcher installed to %s\n' "$desktop_file"
printf 'You may need to log out and back in if your launcher does not refresh immediately.\n'
