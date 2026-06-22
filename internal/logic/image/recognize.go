package image

import (
	"SamaraAI/common/code"
	imgrec "SamaraAI/common/image"
	"io"
	"log"
	"mime/multipart"
	"strings"
)

func MapRecognizeError(err error) code.Code {
	if err == nil {
		return code.CodeSuccess
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "onnxruntime"), strings.Contains(msg, "model file not found"), strings.Contains(msg, "label file not found"):
		return code.AIModelCannotOpen
	case strings.Contains(msg, "decode"), strings.Contains(msg, "invalid"):
		return code.CodeInvalidParams
	default:
		return code.AIModelFail
	}
}

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
