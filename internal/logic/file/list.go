package file

import (
	"SamaraAI/common/rag"
)

func ListRagFiles(username string) ([]rag.UserFileInfo, error) {
	return rag.ListUserFiles(username)
}
