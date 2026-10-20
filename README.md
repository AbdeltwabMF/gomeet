<h1 align="center">gomeet</h1>
<p align="center">Get meeting reminders and automatically open meeting links.</p>

## Get started

Requires [Go](https://go.dev/) 1.26.0 or later.

```sh
make check build
./bin/gomeet
```

On the first run, gomeet creates a `config.json` file and prints its location. Add your meetings, then run gomeet again:

```json
{
  "auto_start": false,
  "events": [
    {
      "summary": "Team meeting",
      "url": "https://meet.google.com/your-meeting",
      "start": {
        "time": "09:00",
        "days": ["Monday", "Wednesday"]
      }
    }
  ]
}
```

Times use your local timezone. Set `auto_start` to `true` to open meeting links automatically when meetings start.

Keep gomeet running to receive reminders. To run it in the background:

```sh
./bin/gomeet --background
```

On Linux, notifications require `notify-send`.

<details>
<summary>Connect Google Calendar (optional)</summary>

1. [Create a Google Cloud project](https://console.cloud.google.com/projectcreate) and [enable the Google Calendar API](https://console.cloud.google.com/apis/library/calendar-json.googleapis.com).
2. Configure the OAuth consent screen. If the app is in testing mode, add your Google account as a test user.
3. In [Credentials](https://console.cloud.google.com/apis/credentials), create an OAuth client ID for a **Desktop app** and download its JSON file.
4. Rename it to `credentials.json` and place it beside `config.json`.
5. Start gomeet and authorize access in your browser.

Put each meeting link in the Google Calendar event’s **Location** field.

You can use Google Calendar alongside meetings in `config.json`.

</details>

<details>
<summary>Start automatically on Windows</summary>

Keep the executable in a permanent location, then run:

```powershell
.\bin\gomeet.exe --install-startup --yes
```

To remove automatic startup:

```powershell
.\bin\gomeet.exe --uninstall-startup
```

If you move the executable, run the install command again.

</details>

On macOS and Linux, use your platform's systemd, launchd, xinitrc, or runit documentation instead.

## Troubleshooting

Check your setup:

```sh
./bin/gomeet --doctor
```

See all options with `./bin/gomeet --help`.

## License

Licensed under the GPL-v3 [License](LICENSE).
