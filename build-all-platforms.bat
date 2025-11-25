@echo off
REM Build AdGuard Home for all major platforms

echo Building AdGuard Home for multiple platforms...
echo.

REM Create dist directory
if not exist "dist" mkdir dist

REM Get version
for /f "delims=" %%i in ('git describe --tags --abbrev^=4 HEAD') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=v0.0.0-dev

echo Version: %VERSION%
echo.

REM Windows AMD64
echo Building Windows AMD64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist\AdGuardHome_windows_amd64.exe
if errorlevel 1 (
    echo Failed to build Windows AMD64
    exit /b 1
)

REM Windows ARM64
echo Building Windows ARM64...
set GOOS=windows
set GOARCH=arm64
go build -ldflags="-s -w" -o dist\AdGuardHome_windows_arm64.exe
if errorlevel 1 (
    echo Failed to build Windows ARM64
    exit /b 1
)

REM Linux AMD64
echo Building Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o dist\AdGuardHome_linux_amd64
if errorlevel 1 (
    echo Failed to build Linux AMD64
    exit /b 1
)

REM Linux ARM64
echo Building Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o dist\AdGuardHome_linux_arm64
if errorlevel 1 (
    echo Failed to build Linux ARM64
    exit /b 1
)

REM Linux ARMv7
echo Building Linux ARMv7...
set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -ldflags="-s -w" -o dist\AdGuardHome_linux_armv7
if errorlevel 1 (
    echo Failed to build Linux ARMv7
    exit /b 1
)

REM macOS AMD64
echo Building macOS AMD64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o dist\AdGuardHome_darwin_amd64
if errorlevel 1 (
    echo Failed to build macOS AMD64
    exit /b 1
)

REM macOS ARM64 (Apple Silicon)
echo Building macOS ARM64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w" -o dist\AdGuardHome_darwin_arm64
if errorlevel 1 (
    echo Failed to build macOS ARM64
    exit /b 1
)

echo.
echo Build completed successfully!
echo.
echo Built files:
dir /b dist\AdGuardHome_*
echo.
echo All binaries are in the 'dist' directory.
