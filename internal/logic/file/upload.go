package file

import (
	"SamaraAI/common/rag"
	"SamaraAI/internal/config"
	"SamaraAI/utils"
	"context"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

func UploadRagFile(username string, file *multipart.FileHeader) (string, error) {
	if err := utils.ValidateFile(file); err != nil {
		log.Printf("File validation failed: %v", err)
		return "", err
	}

	userDir := filepath.Join("uploads", username)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		log.Printf("Failed to create user directory %s: %v", userDir, err)
		return "", err
	}

	files, err := os.ReadDir(userDir)
	if err == nil {
		for _, f := range files {
			if !f.IsDir() {
				if err := rag.DeleteIndex(context.Background(), f.Name()); err != nil {
					log.Printf("Failed to delete index for %s: %v", f.Name(), err)
				}
			}
		}
	}
	if err := utils.RemoveAllFilesInDir(userDir); err != nil {
		log.Printf("Failed to clean user directory %s: %v", userDir, err)
		return "", err
	}

	ext := filepath.Ext(file.Filename)
	filename := utils.GenerateUUID() + ext
	filePath := filepath.Join(userDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	indexer, err := rag.NewRAGIndexer(filename, config.Get().Rag.EmbeddingModel)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	if err := indexer.IndexFile(context.Background(), filePath); err != nil {
		os.Remove(filePath)
		rag.DeleteIndex(context.Background(), filename)
		return "", err
	}

	return filePath, nil
}
