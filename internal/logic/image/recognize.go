package image

import (
	imgrec "SamaraAI/common/image"
	"SamaraAI/internal/config"
	"io"
	"log"
	"mime/multipart"
)

func RecognizeImage(file *multipart.FileHeader) (string, error) {
	cfg := config.Get()
	recognizer, err := imgrec.NewImageRecognizer(
		cfg.Image.ModelPath,
		cfg.Image.LabelPath,
		cfg.Image.InputH,
		cfg.Image.InputW,
	)
	if err != nil {
		log.Println("NewImageRecognizer fail:", err)
		return "", err
	}
	defer recognizer.Close()

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}
	return recognizer.PredictFromBuffer(buf)
}
