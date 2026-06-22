#!/usr/bin/env bash
# 下载图像识别所需模型与标签（Linux / macOS / Git Bash）
# 在项目根目录执行: bash scripts/download-models.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MODEL_DIR="$ROOT/models/mobilenetv2"
MODEL_FILE="$MODEL_DIR/mobilenetv2-7.onnx"
LABEL_FILE="$ROOT/imagenet_classes.txt"

mkdir -p "$MODEL_DIR"

echo ">>> 下载 MobileNetV2 ONNX 模型 (~14MB) ..."
echo "    来源: Hugging Face ONNX Model Zoo 镜像"
MODEL_URL="https://huggingface.co/onnxmodelzoo/mobilenetv2-7/resolve/main/mobilenetv2-7.onnx"
curl -L "$MODEL_URL" -o "$MODEL_FILE"
echo "    已保存: $MODEL_FILE"

echo ""
echo ">>> 下载 ImageNet 标签文件 ..."
echo "    来源: PyTorch 官方仓库"
LABEL_URL="https://raw.githubusercontent.com/pytorch/hub/master/imagenet_classes.txt"
curl -L "$LABEL_URL" -o "$LABEL_FILE"
echo "    已保存: $LABEL_FILE"

echo ""
echo ">>> 下载 ONNX Runtime 动态库 (v1.22.0) ..."
mkdir -p "$ROOT/lib"
ORT_TMP="$ROOT/onnxruntime_tmp"
ORT_TGZ="$ROOT/onnxruntime-linux-x64-1.22.0.tgz"
ORT_URL="https://github.com/microsoft/onnxruntime/releases/download/v1.22.0/onnxruntime-linux-x64-1.22.0.tgz"
curl -L "$ORT_URL" -o "$ORT_TGZ"
tar -xzf "$ORT_TGZ" -C "$ROOT"
cp "$ROOT/onnxruntime-linux-x64-1.22.0/lib/libonnxruntime.so.1.22.0" "$ROOT/lib/"
rm -rf "$ORT_TGZ" "$ROOT/onnxruntime-linux-x64-1.22.0"
echo "    已保存: $ROOT/lib/libonnxruntime.so.1.22.0"

echo ""
echo "完成。请确认 etc/samara.yaml 中 Image 配置为:"
echo "  ModelPath: ./models/mobilenetv2/mobilenetv2-7.onnx"
echo "  LabelPath: ./imagenet_classes.txt"
echo "  RuntimeLibPath: ./lib/libonnxruntime.so.1.22.0  # Linux 可选，留空则自动查找"
