package image

import (
	"SamaraAI/internal/config"
	"fmt"
	"sync"
)

var (
	poolOnce sync.Once
	poolInst *ImageRecognizer
	poolErr  error
)

// GetSharedRecognizer 全局复用 ONNX Session，避免每次请求重新加载模型
func GetSharedRecognizer() (*ImageRecognizer, error) {
	poolOnce.Do(func() {
		cfg := config.Get()
		poolInst, poolErr = NewImageRecognizer(
			cfg.Image.ModelPath,
			cfg.Image.LabelPath,
			cfg.Image.InputH,
			cfg.Image.InputW,
		)
	})
	if poolErr != nil {
		return nil, poolErr
	}
	if poolInst == nil {
		return nil, fmt.Errorf("image recognizer not initialized")
	}
	return poolInst, nil
}
