# win-sound-dev-go-bridge

Go + cgo bridge to `SoundAgentApiDll.dll` for monitoring and querying Windows default audio devices.

## Build (cmd.exe)
Prereq: CGO enabled and a GCC-style toolchain (MinGW-w64 gcc or llvm-mingw clang). MSVC `cl.exe` is not supported by cgo.

```bat
set CGO_ENABLED=1

go build .
```
Place `SoundAgentApiDll.dll` next to the built `.exe` (or on `PATH`).

## Run
```bat
win-sound-dev-go-bridge.exe
```

## Advanced (optional)
- Choose toolchain:
  - `set CC=gcc & set CXX=g++`
  - or `set CC=x86_64-w64-mingw32-clang & set CXX=x86_64-w64-mingw32-clang++`
- Inject version:
  - `go build -ldflags "-X win-sound-dev-go-bridge/pkg/appinfo.Version=1.2.3" .`
