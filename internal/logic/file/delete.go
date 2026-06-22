package file

import (
	"SamaraAI/common/rag"
	"context"
	"fmt"
	"strings"
)

func DeleteRagFile(username, fileID string) error {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return fmt.Errorf("invalid file id")
	}
	return rag.DeleteUserFile(context.Background(), username, fileID)
}
