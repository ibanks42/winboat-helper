# WinBoat Helper

WinBoat Helper is a small Linux tray app for controlling the WinBoat container and launching FreeRDP.

It supports FreeRDP installed as:

- native `xfreerdp`
- native `xfreerdp3`
- Flatpak `com.freerdp.FreeRDP`

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

## Releases

Tagged releases like `v0.1.0` are built and packaged by GitHub Actions for Linux `amd64`.
