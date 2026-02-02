[![.github/workflows/test-workflow.yml](https://github.com/OpenFactorioServerManager/factorio-server-manager/workflows/.github/workflows/test-workflow.yml/badge.svg)](https://github.com/OpenFactorioServerManager/factorio-server-manager/actions)
[![Discord](https://img.shields.io/discord/779512040934342687?label=Discord)](https://discord.gg/SB647WmSbU)

# Factorio Server Manager

### A tool for managing Factorio servers.
This tool runs on a Factorio server and allows management of the Factorio server, saves, mods and many other features.

## Features
* Allows control of the Factorio Server, starting and stopping the Factorio binary.
* Allows the management of save files, upload, download and delete saves.
* Manage installed mods, upload new ones and more
* **Load mods from Factorio 2.0 save files** - automatically install all mods used in a save
* **Load mods from mod-list.json** - upload your local mod list to sync mods to the server
* Manage modpacks, so it is easier to play with different configurations
* Allow viewing of the server logs and current configuration.
* Authentication for protecting against unauthorized users
* Available as a Docker container

## Compatibility
* **Factorio 1.x and 2.0+** - Full support for both Factorio 1.x and 2.0 save file formats
* The save file parser has been updated to handle the new Factorio 2.0 save header format
* **Broad CPU support** - Binaries are built with GOAMD64=v1 for compatibility with older x86-64 processors (no AVX-512 required)

## Quick Start (Docker)

```bash
docker pull petejm/ofsm:dev && docker run -d \
  --name factorio-server-manager \
  -p 9090:80 \
  -v /path/to/factorio:/opt/factorio \
  -v /path/to/fsm-data:/opt/fsm/data \
  petejm/ofsm:dev
```

Access the web UI at `http://your-server:9090`. Default credentials are displayed in the container logs on first run.

### Environment Variables
* `FACTORIO_VERSION` - Factorio version to download (default: `latest`)
* `RCON_PASS` - RCON password for server communication

#### Manage Factorio Server
![Factorio Server Manager Screenshot](screenshots/Screenshot_Controls.png)

#### Manage save files
![Factorio Server Manager Screenshot](screenshots/Screenshot_Saves.png)

#### Manage mods
![Factorio Server Manager Screenshot](screenshots/Screenshot_Mods.png)

## [Installation and Usage](https://github.com/OpenFactorioServerManager/factorio-server-manager/wiki/Installation-and-Usage)

## [Development](https://github.com/OpenFactorioServerManager/factorio-server-manager/wiki/Development)

## Contributing
1. Fork it!
2. Checkout the develop branch, only use that as a base: `git checkout develop`
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Add your changes a in human readable way into CHANGELOG.md
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request, with `develop` as base :D

## Authors

* **Mitch Roote** - [roote.ca](https://roote.ca)
* **[knoxfighter](https://github.com/knoxfighter)**
* **[Jannaahs](https://github.com/jannaahs)**
* **[Pete McDade](https://github.com/petejm)**

## Special Thanks
- **[All Contributions](https://github.com/OpenFactorioServerManager/factorio-server-manager/graphs/contributors)**
- **mickael9** for reverseengineering the factorio-save-file: https://forums.factorio.com/viewtopic.php?f=5&t=8568#
- **[Anthropic](https://anthropic.com)** for Claude AI assistance in modernizing and improving this codebase

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
