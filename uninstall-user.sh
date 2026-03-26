#!/usr/bin/env bash
set -euo pipefail

bin_name="winboat-helper"
app_id="dev.ibanks.winboat-helper"

target_bin="${HOME}/.local/bin/$bin_name"
target_icon="${HOME}/.local/share/icons/hicolor/256x256/apps/$bin_name.png"
desktop_file="${HOME}/.local/share/applications/$bin_name.desktop"
support_dir="${HOME}/.local/lib/$bin_name"
autostart_file="${XDG_CONFIG_HOME:-$HOME/.config}/autostart/${app_id}.desktop"

rm -f "$target_bin" "$target_icon" "$desktop_file" "$autostart_file"
rm -rf "$support_dir"

if command -v update-desktop-database >/dev/null 2>&1; then
	update-desktop-database "${HOME}/.local/share/applications" >/dev/null 2>&1 || true
fi

printf 'Removed WinBoat Helper from your user profile.\n'
printf 'Your saved settings in %s were left intact.\n' "${XDG_CONFIG_HOME:-$HOME/.config}/winboat-rdp"
