<h1 align="center">gomeet</h1>

<h4 align="center">A meeting reminder and auto-joiner that simplifies meeting management by automatically storing, retrieving, and opening meeting links</h4>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/AbdeltwabMF/gomeet"><img src="https://goreportcard.com/badge/github.com/AbdeltwabMF/gomeet" alt="Go Report Card"></a>
  <img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/AbdeltwabMF/gomeet/release.yaml">
</p>

## Configuration

Before using `gomeet`, configure it with your details in a valid JSON format. You'll find an example configuration file [config.json](configs/config.json) to get you started.

> [!IMPORTANT]
> The [config.json](configs/config.json) file is placed in the configuration directory of your system:
>
> - Windows: it's `%AppData%\gomeet\` (i.e. `C:\Users\<user-name>\AppData\Roaming\gomeet\`).
> - Darwin: it's `$HOME/Library/Application Support/gomeet/`.
> - Gnu/Linux: it's `$XDG_CONFIG_HOME/gomeet/`, if non-empty, else `$HOME/.config/gomeet/`.

### Adding calendars

**Local Calendar**: Include your local calendar details in the `"events"` array within [config.json](configs/config.json).

**Google Calendar**:

- [Create a Google Cloud project](https://developers.google.com/workspace/guides/create-project).
- Generate a [credentials.json](configs/credentials.json) file and place it alongside [config.json](configs/config.json) in the gomeet config directory.

## Build

**Requirements**: [Go programming language](https://go.dev/) (version go1.22.1 or later)

```shell
make build
```

## Usage

### Manual launch

Simply run the built binary to start the program as a background daemon.

```shell
./bin/gomeet.exe &
```

### Startup as Service (Windows only)

1. Open the registry editor (regedit).
2. Navigate to `HKEY_LOCAL_MACHINE\Software\Microsoft\Windows\CurrentVersion\Run`.
3. Create a new String Value named `GoMeet`.
4. Set the `GoMeet` value to the full path of the gomeet executable (e.g., `C:\bin\gomeet.exe`).

For automatic startup on Unix-based systems, refer to your specific distribution's documentation on creating systemd or runit services.

> [!NOTE] > `gomeet` logs important events (errors, warnings) to a log file for troubleshooting.
>
> - Windows: `%LocalAppData%\gomeet\logs\` or `C:\Users\<user-name>\AppData\Local\gomeet\logs\`.
> - Darwin: `~/Library/Logs/gomeet/`.
> - Gnu/Linux: `/var/log/gomeet/`

## License

Licensed under the GPL-v3 [License](LICENSE).
