@echo off
setlocal EnableExtensions EnableDelayedExpansion

cd /d "%~dp0.."

rem Prefer the standard 64-bit Go installation when an older 32-bit Go is also on PATH.
if exist "C:\Program Files\Go\bin\go.exe" set "PATH=C:\Program Files\Go\bin;%PATH%"

where go >nul 2>nul || (
  echo Go is required. Install 64-bit Go, then run this script again.
  exit /b 1
)

where gcc >nul 2>nul
if errorlevel 1 if exist "C:\msys64\ucrt64\bin\gcc.exe" set "PATH=C:\msys64\ucrt64\bin;%PATH%"
where gcc >nul 2>nul
if errorlevel 1 (
  for /f "usebackq delims=" %%G in (`"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -File scripts\ensure-build-tools.ps1`) do set "GCC_BIN=%%G"
  if not defined GCC_BIN exit /b 1
  set "PATH=!GCC_BIN!;!PATH!"
)
where gcc >nul 2>nul || (
  echo A 64-bit GCC compiler could not be prepared.
  exit /b 1
)

for /f "tokens=*" %%A in ('go env GOARCH') do set "BUILD_ARCH=%%A"
if /I not "%BUILD_ARCH%"=="amd64" (
  echo This release build requires 64-bit Go. Current architecture: %BUILD_ARCH%
  exit /b 1
)

set "CGO_ENABLED=1"
if not exist "dist" mkdir "dist"
if not exist "build" mkdir "build"

"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -File scripts\fetch-assets.ps1
if errorlevel 1 exit /b 1

"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -File scripts\prepare-bundle.ps1
if errorlevel 1 exit /b 1

go test -tags bundled ./cmd/sentiment-server ./internal/...
if errorlevel 1 exit /b 1

go build -tags bundled -ldflags "-s -w -extldflags=-static" -o .\build\cea-sentiment-local-api.exe .\cmd\sentiment-server
if errorlevel 1 exit /b 1

copy /Y build\cea-sentiment-local-api.exe dist\cea-sentiment-local-api.exe >nul
if errorlevel 1 (
  echo Could not replace dist\cea-sentiment-local-api.exe because it is running.
  echo Stop the local service, then run this script again.
  exit /b 1
)

copy /Y cea-sentiment-local-api.json dist\cea-sentiment-local-api.json >nul
if errorlevel 1 exit /b 1

echo.
echo Release files created in dist\
echo   dist\cea-sentiment-local-api.exe
echo   dist\cea-sentiment-local-api.json
