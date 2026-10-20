<h1 align="center">gomeet</h1>

<h4 align="center">A meeting reminder and auto-joiner that simplifies meeting management by automatically storing, retrieving, and opening meeting links</h4>

<p align="center">
  <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/AbdeltwabMF/gomeet">
  <a href="https://goreportcard.com/report/github.com/AbdeltwabMF/gomeet"><img src="https://goreportcard.com/badge/github.com/AbdeltwabMF/gomeet" alt="Go Report Card"></a>
  <img alt="GitHub License" src="https://img.shields.io/github/license/AbdeltwabMF/gomeet">
  <img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/AbdeltwabMF/gomeet/release.yaml">
  <img alt="Static Badge" src="https://img.shields.io/badge/viva-palestina-149954">
</p>

## Configuration

Before using `gomeet`, configure it with your details in a valid JSON format. You'll find an example configuration file [config.example.json](examples/config.example.json) to get you started.

> [!IMPORTANT]
> The `config.json` file is placed in the configuration directory of your system:
>
> - Windows: it's `%AppData%\gomeet\` (i.e. `C:\Users\<user-name>\AppData\Roaming\gomeet\`).
> - Darwin: it's `$HOME/Library/Application Support/gomeet/`.
> - Gnu/Linux: it's `$XDG_CONFIG_HOME/gomeet/`, if non-empty, else `$HOME/.config/gomeet/`.

The configuration is validated when `gomeet` starts. Unknown fields, invalid
times, unknown weekdays and unsafe URLs are rejected with a message that
identifies the offending event, so a typo is reported instead of silently
disabling a meeting.

| Field | Description |
| --- | --- |
| `auto_start` | Open the meeting link automatically when a meeting starts. |
| `events[].summary` | Name shown in the notification. |
| `events[].url` | Meeting link. Only `http` and `https` URLs with a host are accepted. |
| `events[].start.time` | Start time in 24-hour `HH:MM` format, in the local timezone. |
| `events[].start.days` | Weekdays the meeting repeats on, for example `"Monday"`. |

### Adding calendars

**Local Calendar**: Include your local calendar details in the `"events"` array within `config.json`.

**Google Calendar**:

- [Create a Google Cloud project](https://developers.google.com/workspace/guides/create-project).
- Generate a `credentials.json` file and place it alongside `config.json` in the gomeet config directory. See [credentials.example.json](examples/credentials.example.json) for the expected format.
- On the first start, a browser window opens so that you can authorize access. The token is stored in the configuration directory and refreshed automatically.

> [!IMPORTANT]
> The URL of the event should be placed in **Location** (الموقع الجغرافي) field.

> [!NOTE]
> Google Calendar is optional. When no `credentials.json` is present, only the
> local calendar is monitored.

### Notifications

Notifications are shown with the notification service of your system:

- **Windows**: a toast notification.
- **Darwin**: `osascript`, which is part of macOS.
- **Gnu/Linux**: `notify-send`, provided by `libnotify`. Install it with `sudo apt install libnotify-bin` if it is missing.

## Build

**Requirements**: [Go programming language](https://go.dev/) (version go1.26.1 or later)

```shell
make build
```

## Development

```shell
make test     # run the test suite
make vet      # run go vet
make check    # formatting, vet and tests
```

## Usage

### Manual launch

Simply run the built binary to start the program as a background daemon.

```shell
./bin/gomeet &
```

### Startup as Service (Windows only)

1. Open the registry editor (regedit).
2. Navigate to `HKEY_LOCAL_MACHINE\Software\Microsoft\Windows\CurrentVersion\Run`.
3. Create a new String Value named `GoMeet`.
4. Set the `GoMeet` value to the full path of the gomeet executable (e.g., `C:\bin\gomeet.exe`).

For automatic startup on Unix-based systems, refer to your specific distribution's documentation on creating systemd or runit services.

> [!NOTE]
> `gomeet` logs important events (errors, warnings) to a log file for troubleshooting. Logs are appended, so previous runs are preserved.
>
> - Windows: `%LocalAppData%\gomeet\logs\` or `C:\Users\<user-name>\AppData\Local\gomeet\logs\`.
> - Darwin: `~/Library/Logs/gomeet/`.
> - Gnu/Linux: `$XDG_CACHE_HOME/gomeet/logs/`, if set, otherwise `$HOME/.cache/gomeet/logs/`.

## License

Licensed under the GPL-v3 [License](LICENSE).
