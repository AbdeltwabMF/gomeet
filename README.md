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

On its first run, `gomeet` creates a valid, empty `config.json` and prints its
location. Add your meetings to its `events` array and run the command again.
The generated file starts with:

```json
{
  "auto_start": false,
  "events": []
}
```

A configuration with sample meetings is available at
[config.example.json](examples/config.example.json).

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

The local and Google calendar sources are independent; you can enable either
one or both. The `auto_start` setting applies to meetings from both sources.

**Local Calendar**: Include your local calendar details in the `"events"` array within `config.json`.

**Google Calendar**:

- [Create or select a Google Cloud project](https://console.cloud.google.com/projectcreate).
- [Enable the Google Calendar API](https://console.cloud.google.com/apis/library/calendar-json.googleapis.com).
- Configure the OAuth consent screen. If the app is in testing mode, add your Google account as a test user.
- Open [Google Cloud credentials](https://console.cloud.google.com/apis/credentials), create an OAuth client ID with application type **Desktop app**, and download its JSON file. See Google's [Desktop app credential guide](https://developers.google.com/workspace/guides/create-credentials#desktop-app) for more details.
- Rename the downloaded file to `credentials.json` and place it alongside `config.json` in the gomeet config directory. See [credentials.example.json](examples/credentials.example.json) for the expected format.
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

**Requirements**: [Go programming language](https://go.dev/) (version go1.26.0 or later)

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

### Foreground (default)

Run the binary directly. It stays attached to the terminal and writes warnings
and errors both to the terminal and to the persistent log file. Press `Ctrl+C`
to stop it.

```shell
./bin/gomeet
```

### Background

Use `--background` to start a detached process. The command prints its process
ID and log-file location before returning.

```shell
./bin/gomeet --background
```

### Diagnostics

`--doctor` is a self-contained system check. It creates the empty config if it
is missing, validates the config, verifies log access and required platform
commands, prints the relevant paths, and gives installation or configuration
instructions for anything missing.

```shell
./bin/gomeet --doctor
```

### Verbose logging

Warnings and errors are logged by default. Add `--verbose` (or `-v`) to include
informational and debug messages with source locations. It can be combined
with foreground or background mode.

```shell
./bin/gomeet --verbose
./bin/gomeet --background --verbose
```

Run `./bin/gomeet --help` for the complete option summary.

### Startup as Service (Windows only)

1. Open the registry editor (regedit).
2. Navigate to `HKEY_LOCAL_MACHINE\Software\Microsoft\Windows\CurrentVersion\Run`.
3. Create a new String Value named `GoMeet`.
4. Set the `GoMeet` value to the full path of the gomeet executable (e.g., `C:\bin\gomeet.exe`).

For automatic startup on Unix-based systems, refer to your specific distribution's documentation on creating systemd or runit services.

> [!NOTE]
> `gomeet` logs warnings and errors to a log file for troubleshooting. With
> `--verbose`, informational and debug messages are included too. Logs are
> appended, so previous runs are preserved.
>
> - Windows: `%LocalAppData%\gomeet\logs\` or `C:\Users\<user-name>\AppData\Local\gomeet\logs\`.
> - Darwin: `~/Library/Logs/gomeet/`.
> - Gnu/Linux: `$XDG_CACHE_HOME/gomeet/logs/`, if set, otherwise `$HOME/.cache/gomeet/logs/`.

## License

Licensed under the GPL-v3 [License](LICENSE).
