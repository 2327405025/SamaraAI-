@echo off
setlocal EnableExtensions
cd /d "%~dp0.."

if not exist "models\mobilenetv2" mkdir "models\mobilenetv2"

echo.
if exist "models\mobilenetv2\mobilenetv2-7.onnx" (
  echo [1/3] mobilenetv2-7.onnx already exists, skip.
) else (
  echo [1/3] Download mobilenetv2-7.onnx ...
  curl.exe -L --retry 3 -o "models\mobilenetv2\mobilenetv2-7.onnx" "https://hf-mirror.com/onnxmodelzoo/mobilenetv2-7/resolve/main/mobilenetv2-7.onnx"
  if errorlevel 1 (
    echo HuggingFace mirror failed, trying huggingface.co ...
    curl.exe -L --retry 3 -o "models\mobilenetv2\mobilenetv2-7.onnx" "https://huggingface.co/onnxmodelzoo/mobilenetv2-7/resolve/main/mobilenetv2-7.onnx"
  )
  if errorlevel 1 goto download_fail
  if not exist "models\mobilenetv2\mobilenetv2-7.onnx" goto download_fail
)

echo.
if exist "imagenet_classes.txt" (
  echo [2/3] imagenet_classes.txt already exists, skip.
) else (
  echo [2/3] Download imagenet_classes.txt ...
  curl.exe -L --retry 3 -o "imagenet_classes.txt" "https://raw.githubusercontent.com/pytorch/hub/master/imagenet_classes.txt"
  if errorlevel 1 goto label_fail
  if not exist "imagenet_classes.txt" goto label_fail
)

echo.
if exist "onnxruntime.dll" (
  echo [3/3] onnxruntime.dll already exists, skip.
  echo       To reinstall: scripts\download-onnxruntime.cmd
  goto done
)
echo [3/3] Download onnxruntime.dll (v1.22.0) ...
call "%~dp0download-onnxruntime.cmd"
if errorlevel 1 exit /b 1

:done
echo.
echo Done.
echo Model:  %CD%\models\mobilenetv2\mobilenetv2-7.onnx
echo Labels: %CD%\imagenet_classes.txt
echo ORT:    %CD%\onnxruntime.dll
echo.
echo Check etc\samara.yaml Image section uses relative paths above.
exit /b 0

:download_fail
echo.
echo ERROR: model download failed.
echo Open in browser and save manually:
echo https://hf-mirror.com/onnxmodelzoo/mobilenetv2-7/resolve/main/mobilenetv2-7.onnx
echo Save to: %CD%\models\mobilenetv2\mobilenetv2-7.onnx
exit /b 1

:label_fail
echo.
echo ERROR: label download failed.
echo Open in browser and save manually:
echo https://raw.githubusercontent.com/pytorch/hub/master/imagenet_classes.txt
echo Save to: %CD%\imagenet_classes.txt
exit /b 1
