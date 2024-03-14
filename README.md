# gomeet

GoMeet is designed to elevate productivity by seamlessly connecting you to meetings while eliminating the need to manually retrieve and open meeting links.

## Build Binary

On windows:

```shell
go build -o bin/gomeet.exe -ldflags "-H windowsgui" main.go
```

On Unix-based Systems:

```shell
go build -o bin/gomeet main.go
```

## Usage

You have two options for running it: either on a per-usage basis or by adding the service to startup through the registry editor.

### Per-Usage Basis

Simply execute the built binary to start the daemon in the background.

```shell
./bin/gomeet.exe &
```

### Startup as Background Service

On windows, to have `gomeet` start automatically on startup, follow these steps:

1. Open the registry editor by typing `regedit` in the Windows search bar and pressing Enter.
2. Navigate to `HKEY_LOCAL_MACHINE\Software\Microsoft\Windows\CurrentVersion\Run`.
3. Right-click on the right pane, select `New` -> `String Value`.
4. Name the new value as `GoMeet`.
5. Double-click the new value and set its data to the full path of the `gomeet` executable. For example: `C:\bin\gomeet.exe`.

On Unix-based systems, you'll need to perform further research on how to set up gomeet to start automatically on system startup.
This typically involves creating a `systemd` or `runit` service or adding an entry to the appropriate startup configuration file for your distribution.

> [!IMPORTANT]
> Ensure that the `meetings.json` file is placed in the config directory of your system:
> On Windows, it's `%AppData%` (i.e. `C:\Users\User\AppData\Roaming\`).
> On Darwin, it's `$HOME/Library/Application Support`.
> On Unix systems, it's `$XDG_CONFIG_HOME`, if non-empty, else `$HOME/.config`.

## License

Licensed under the GPL-v3 [License](LICENSE).
