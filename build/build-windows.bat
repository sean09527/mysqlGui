@echo off
REM MySQL Manager - Windows Build Script

echo ======================================
echo MySQL Manager - Windows Build
echo ======================================
echo.

REM Check if wails is installed
where wails >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: Wails CLI is not installed
    echo Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    exit /b 1
)

REM Clean previous builds
echo Cleaning previous builds...
if exist build\bin rmdir /s /q build\bin
mkdir build\bin
echo Clean complete
echo.

REM Build for Windows
echo Building for Windows (amd64)...
wails build -clean -platform windows/amd64
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b 1
)
echo Windows build complete
echo.

REM Build with NSIS installer (optional)
echo.
set /p BUILD_INSTALLER="Build NSIS installer? (y/n): "
if /i "%BUILD_INSTALLER%"=="y" (
    echo Building NSIS installer...
    wails build -platform windows/amd64 -nsis
    if %ERRORLEVEL% NEQ 0 (
        echo Installer build failed!
        exit /b 1
    )
    echo Installer build complete
    echo.
)

REM List built files
echo ======================================
echo Build Summary
echo ======================================
echo.
echo Built files in build\bin\:
dir /b build\bin\
echo.

echo ======================================
echo Build Complete!
echo ======================================
echo.
echo Output location: build\bin\
echo.
pause
