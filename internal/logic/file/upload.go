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
	"sort"
	"strings"
)

func UploadRagFile(username string, file *multipart.FileHeader) (storedName string, displayName string, err error) {
	if err := utils.ValidateFile(file); err != nil {
		log.Printf("File validation failed: %v", err)
		return "", "", err
	}

	userDir := rag.UserUploadDir(username)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		log.Printf("Failed to create user directory %s: %v", userDir, err)
		return "", "", err
	}

	if err := trimUserFiles(username, userDir); err != nil {
		return "", "", err
	}

	displayName = filepath.Base(file.Filename)
	ext := filepath.Ext(displayName)
	filename := utils.GenerateUUID() + ext
	filePath := filepath.Join(userDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", "", err
	}

	if err := rag.SaveFileDisplayName(username, filename, displayName); err != nil {
		os.Remove(filePath)
		return "", "", err
	}

	indexer, err := rag.NewRAGIndexer(filename, config.Get().Rag.EmbeddingModel)
	if err != nil {
		os.Remove(filePath)
		rag.RemoveFileDisplayName(username, filename)
		return "", "", err
	}

	if err := indexer.IndexFile(context.Background(), filePath); err != nil {
		os.Remove(filePath)
		rag.RemoveFileDisplayName(username, filename)
		rag.DeleteIndex(context.Background(), filename)
		return "", "", err
	}

	return filename, displayName, nil
}

func trimUserFiles(username string, userDir string) error {
	maxFiles := config.Get().Rag.MaxFilesPerUser
	if maxFiles <= 0 {
		maxFiles = 5
	}

	entries, err := os.ReadDir(userDir)
	if err != nil {
		return nil
	}

	type fileEntry struct {
		name    string
		modTime int64
	}
	var files []fileEntry
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".display") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileEntry{name: e.Name(), modTime: info.ModTime().Unix()})
	}

	if len(files) < maxFiles {
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime < files[j].modTime
	})

	toRemove := len(files) - maxFiles + 1
	for i := 0; i < toRemove; i++ {
		name := files[i].name
		if err := rag.DeleteIndex(context.Background(), name); err != nil {
			log.Printf("Failed to delete index for %s: %v", name, err)
		}
		if err := os.Remove(filepath.Join(userDir, name)); err != nil {
			log.Printf("Failed to remove file %s for user %s: %v", name, username, err)
		}
		rag.RemoveFileDisplayName(username, name)
	}
	return nil
}
