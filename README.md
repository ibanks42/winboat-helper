# WinBoat Helper

WinBoat Helper is a small Linux tray app for controlling the WinBoat container and launching FreeRDP.

## Install

Install the latest public release for your user account:

```bash
curl -fsSL https://raw.githubusercontent.com/ibanks42/winboat-helper/main/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/ibanks42/winboat-helper/main/install.sh | bash -s -- v0.1.0
```

## Uninstall

```bash
~/.local/lib/winboat-helper/uninstall-user.sh
```

## Build A Release Bundle

```bash
./build-release.sh v0.1.0
```

This creates a tarball in `dist/` that contains the binary plus the user-local installer.
