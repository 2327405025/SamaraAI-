# 下载图像识别所需模型与标签（Windows PowerShell）
# 在项目根目录执行: .\scripts\download-models.ps1

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

$ModelDir = Join-Path $Root "models\mobilenetv2"
$ModelFile = Join-Path $ModelDir "mobilenetv2-7.onnx"
$LabelFile = Join-Path $Root "imagenet_classes.txt"

New-Item -ItemType Directory -Force -Path $ModelDir | Out-Null

Write-Host ">>> 下载 MobileNetV2 ONNX 模型 (~14MB) ..."
Write-Host "    来源: Hugging Face ONNX Model Zoo 镜像"
$ModelUrl = "https://huggingface.co/onnxmodelzoo/mobilenetv2-7/resolve/main/mobilenetv2-7.onnx"
Invoke-WebRequest -Uri $ModelUrl -OutFile $ModelFile -UseBasicParsing
Write-Host "    已保存: $ModelFile"

Write-Host ""
Write-Host ">>> 下载 ImageNet 标签文件 ..."
Write-Host "    来源: PyTorch 官方仓库"
$LabelUrl = "https://raw.githubusercontent.com/pytorch/hub/master/imagenet_classes.txt"
Invoke-WebRequest -Uri $LabelUrl -OutFile $LabelFile -UseBasicParsing
Write-Host "    已保存: $LabelFile"

Write-Host ""
Write-Host "完成。请确认 etc/samara.yaml 中 Image 配置为:"
Write-Host "  ModelPath: ./models/mobilenetv2/mobilenetv2-7.onnx"
Write-Host "  LabelPath: ./imagenet_classes.txt"
