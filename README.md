# win-sound-dev-go-bridge

Go + cgo bridge to `SoundAgentApiDll.dll` for monitoring/querying Windows default audio devices.

## Build (powershell)
Prereqs: CGO enabled and a GCC-style toolchain (MinGW-w64 gcc or LLVM-mingw clang).
E.g. if you download llvm-mingw-20251118-msvcrt-x86_64.zip, copy its bin, include,
lib and x86_64-w64-mingw32 folders to C:\tools\llvm-mingw 

```powershell
$Env:CGO_ENABLED = "1"
$Env:CC = "C:\tools\llvm-mingw\bin\x86_64-w64-mingw32-clang.exe"
$Env:CXX = "C:\tools\llvm-mingw\bin\x86_64-w64-mingw32-clang++.exe"

go build -o (Join-Path $PWD.Path 'bin/')

.\scripts\fetch-native.ps1

```


## Run
```powershell
.\bin\win-sound-dev-go-bridge.exe
```

## External module
- github.com/eduarddanziger/sound-win-scanner/v4 (pkg/soundlibwrap): cgo wrapper around SoundAgentApi, see [soundlibwrap documentation](https://pkg.go.dev/github.com/eduarddanziger/sound-win-scanner/v4/pkg/soundlibwrap)

## Advanced
Use clang instead of gcc:
```bat
set CC=x86_64-w64-mingw32-clang
set CXX=x86_64-w64-mingw32-clang++
```
