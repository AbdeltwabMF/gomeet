<h1 align="center">gomeet</h1>

<p align="center">A meeting reminder and auto-joiner, eliminating the need for manual work to store, retrieve, and open links</p>

## Build

Ensure you have the Go programming language installed, preferably version `go1.22.1` or later.

On windows:

```shell
go build -o bin/gomeet.exe -ldflags "-H windowsgui -s -w" ./cmd
```

On Unix-based Systems:

```shell
go build -o bin/gomeet -ldflags "-s -w" ./cmd
```

## Usage

Before using `gomeet`, ensure that the `config.json` file is populated with your own data in a valid JSON format (see `config.json`).

> [!IMPORTANT]
> The `config.json` file is placed in the configuration directory of your system:
>
> - Windows: it's `%AppData%` (i.e. `C:\Users\<user-name>\AppData\Roaming\gomeet\`).
> - Darwin: it's `$HOME/Library/Application Support/gomeet`.
> - Gnu/Linux: it's `$XDG_CONFIG_HOME`, if non-empty, else `$HOME/.config/gomeet`.

`gomeet` logs important events such as errors or warnings to a log file. You can refer to this log file in case of any issues.

- Windows: `%LocalAppData%\gomeet\logs\` or `C:\Users\<user-name>\AppData\Local\gomeet\logs`.
- Darwin: `~/Library/Logs/gomeet/`.
- Gnu/Linux: `/var/log/gomeet/`


### 1. Per-Usage Basis

Simply execute the built binary to start the daemon in the background.

```shell
./bin/gomeet.exe &
```

### 2. Startup as Service

On windows, to have `gomeet` start automatically on startup, follow these steps:

1. Open the registry editor by typing `regedit` in the Windows search bar and pressing Enter.
2. Navigate to `HKEY_LOCAL_MACHINE\Software\Microsoft\Windows\CurrentVersion\Run`.
3. Right-click on the right pane, select `New` -> `String Value`.
4. Name the new value as `GoMeet`.
5. Double-click the new value and set its data to the full path of the `gomeet` executable. For example: `C:\bin\gomeet.exe`.

On Unix-based systems, you'll need to perform further research on how to set up `gomeet` to start automatically on system startup.
This typically involves creating a `systemd` or `runit` service or adding an entry to the appropriate startup configuration file for your distribution.

## License

Licensed under the GPL-v3 [License](LICENSE).
