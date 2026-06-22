@echo off
setlocal EnableExtensions
cd /d "%~dp0.."

set "ORT_ZIP=onnxruntime-win-x64-1.22.0.zip"
set "ORT_DIR=onnxruntime-win-x64-1.22.0"
set "ORT_URL=https://github.com/microsoft/onnxruntime/releases/download/v1.22.0/%ORT_ZIP%"
set "ORT_MIRROR=https://ghfast.top/https://github.com/microsoft/onnxruntime/releases/download/v1.22.0/%ORT_ZIP%"
set "MIN_ZIP_SIZE=65000000"
set "DL_ZIP=%TEMP%\samara-%ORT_ZIP%"

echo.
echo === ONNX Runtime 1.22.0 (Windows x64) ===
echo.

if exist "onnxruntime.dll" (
  echo Removing old onnxruntime.dll ...
  del /f /q "onnxruntime.dll"
)
if exist "%ORT_ZIP%" (
  echo Removing old zip ...
  del /f /q "%ORT_ZIP%" 2>nul
)
if exist "%DL_ZIP%" del /f /q "%DL_ZIP%" 2>nul
if exist "onnxruntime_tmp" (
  rmdir /s /q "onnxruntime_tmp"
)

echo Downloading %ORT_ZIP% (~69MB) ...
curl.exe -L --retry 8 --retry-delay 5 -o "%DL_ZIP%" "%ORT_MIRROR%"
if errorlevel 1 (
  echo Mirror failed, trying GitHub ...
  curl.exe -L --retry 8 --retry-delay 5 -o "%DL_ZIP%" "%ORT_URL%"
)
if errorlevel 1 goto download_fail

for %%A in ("%DL_ZIP%") do set ZIP_SIZE=%%~zA
if %ZIP_SIZE% LSS %MIN_ZIP_SIZE% (
  echo ERROR: zip too small (%ZIP_SIZE% bytes^), download incomplete.
  del /f /q "%DL_ZIP%" 2>nul
  goto download_fail
)

echo Extracting onnxruntime.dll only ...
mkdir onnxruntime_tmp 2>nul
tar -xf "%DL_ZIP%" -C onnxruntime_tmp "%ORT_DIR%/lib/onnxruntime.dll"
if errorlevel 1 goto extract_fail

copy /y "onnxruntime_tmp\%ORT_DIR%\lib\onnxruntime.dll" "onnxruntime.dll" >nul
if errorlevel 1 goto extract_fail
if not exist "onnxruntime.dll" goto extract_fail

for %%A in ("onnxruntime.dll") do set DLL_SIZE=%%~zA
if %DLL_SIZE% LSS 12000000 (
  echo ERROR: onnxruntime.dll too small (%DLL_SIZE% bytes^).
  del /f /q "onnxruntime.dll"
  goto extract_fail
)

del /f /q "%DL_ZIP%" 2>nul
rmdir /s /q "onnxruntime_tmp"

echo.
echo Done.
echo   %CD%\onnxruntime.dll  (%DLL_SIZE% bytes)
echo.
echo Restart backend and check log: image recognizer init success
exit /b 0

:download_fail
echo.
echo ERROR: download failed.
echo Open in browser and save manually:
echo   %ORT_URL%
echo Then extract lib\onnxruntime.dll to:
echo   %CD%\onnxruntime.dll
exit /b 1

:extract_fail
echo.
echo ERROR: extract failed. Delete zip and retry, or extract manually from:
echo   %ORT_ZIP%
echo Copy lib\onnxruntime.dll to %CD%\onnxruntime.dll
exit /b 1
