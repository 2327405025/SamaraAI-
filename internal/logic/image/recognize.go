package image

import (
	imgrec "SamaraAI/common/image"
	"io"
	"log"
	"mime/multipart"
)

func RecognizeImage(file *multipart.FileHeader) (string, error) {
	recognizer, err := imgrec.GetSharedRecognizer()
	if err != nil {
		log.Println("GetSharedRecognizer fail:", err)
		return "", err
	}

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
