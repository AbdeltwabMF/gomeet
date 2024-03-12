# gomeet

Streamlining your meeting scheduling with ease.

## Build Executable

```shell
go build -o bin/gomeet.exe main.go
```

## Usage

You have two options for running it: either on a per-usage basis or by adding the service to startup through the registry editor.

### Per-Usage Basis

Simply execute the built binary to launch the application.

```shell
./bin/gomeet.exe
```

### Startup Service (Windows)

To have `gomeet` launch automatically on startup, follow these steps:

1. Open the registry editor by typing `regedit` in the Windows search bar and pressing Enter.
2. Navigate to `HKEY_LOCAL_MACHINE\Software\Microsoft\Windows\CurrentVersion\Run`.
3. Right-click on the right pane, select `New` -> `String Value`.
4. Name the new value as `GoMeet`.
5. Double-click the new value and set its data to the full path of the `gomeet` executable. For example: `C:\bin\gomeet.exe`.
