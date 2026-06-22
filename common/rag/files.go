package rag

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type UserFileInfo struct {
	FileID   string `json:"fileId"`
	FileName string `json:"fileName"`
}

func displayNamePath(userDir, storedName string) string {
	return filepath.Join(userDir, storedName+".display")
}

func SaveFileDisplayName(username, storedName, displayName string) error {
	userDir := UserUploadDir(username)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(displayNamePath(userDir, storedName), []byte(displayName), 0644)
}

func ReadFileDisplayName(username, storedName string) string {
	b, err := os.ReadFile(displayNamePath(UserUploadDir(username), storedName))
	if err != nil || len(b) == 0 {
		return storedName
	}
	return string(b)
}

func RemoveFileDisplayName(username, storedName string) {
	_ = os.Remove(displayNamePath(UserUploadDir(username), storedName))
}

// ListUserFiles 列出用户已上传的知识库文件（含展示名）
func ListUserFiles(username string) ([]UserFileInfo, error) {
	userDir := UserUploadDir(username)
	entries, err := os.ReadDir(userDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []UserFileInfo
	for _, f := range entries {
		if f.IsDir() || strings.HasSuffix(f.Name(), ".display") {
			continue
		}
		name := f.Name()
		files = append(files, UserFileInfo{
			FileID:   name,
			FileName: ReadFileDisplayName(username, name),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].FileName < files[j].FileName
	})
	return files, nil
}

// DeleteUserFile 删除用户已上传的知识库文件及 Redis 索引
func DeleteUserFile(ctx context.Context, username, fileID string) error {
	if fileID == "" || strings.Contains(fileID, "..") || strings.ContainsAny(fileID, `/\`) {
		return fmt.Errorf("invalid file id")
	}

	userDir := UserUploadDir(username)
	filePath := filepath.Join(userDir, fileID)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found")
		}
		return err
	}

	if err := DeleteIndex(ctx, fileID); err != nil {
		return err
	}
	if err := os.Remove(filePath); err != nil {
		return err
	}
	RemoveFileDisplayName(username, fileID)
	return nil
}
