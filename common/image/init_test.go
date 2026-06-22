package image

import (
	"os"
	"testing"

	"SamaraAI/internal/config"
)

func TestSharedRecognizerInit(t *testing.T) {
	if _, err := os.Stat("../../onnxruntime.dll"); err != nil {
		t.Skip("onnxruntime.dll not in project root")
	}
	if _, err := os.Stat("../../models/mobilenetv2/mobilenetv2-7.onnx"); err != nil {
		t.Skip("model not found")
	}
	if _, err := os.Stat("../../imagenet_classes.txt"); err != nil {
		t.Skip("labels not found")
	}

	c := config.Config{}
	c.Image.ModelPath = "../../models/mobilenetv2/mobilenetv2-7.onnx"
	c.Image.LabelPath = "../../imagenet_classes.txt"
	c.Image.InputH = 224
	c.Image.InputW = 224
	config.InitGlobal(&c)

	r, err := GetSharedRecognizer()
	if err != nil {
		t.Fatalf("GetSharedRecognizer: %v", err)
	}
	if r == nil {
		t.Fatal("recognizer is nil")
	}
}
