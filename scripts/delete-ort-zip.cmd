@echo off
setlocal
set "ZIP=D:\Gocode\SamaraAI-v2\onnxruntime-win-x64-1.22.0.zip"
if not exist "%ZIP%" (
  echo File not found, already deleted.
  exit /b 0
)
echo Closing lock: please close this zip in Cursor/IDE first, then press any key...
pause >nul
:retry
del /f /q "%ZIP%" 2>nul
if exist "%ZIP%" (
  echo Still in use. Close the file tab and press any key to retry...
  pause >nul
  goto retry
)
echo Deleted: %ZIP%
exit /b 0
