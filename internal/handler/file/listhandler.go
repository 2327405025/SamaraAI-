package file

import (
	"SamaraAI/common/code"
	filelogic "SamaraAI/internal/logic/file"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/middleware"
	"SamaraAI/internal/types"
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListRagFilesHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.UserNameFromContext(r.Context())
	if username == "" {
		httpx.OkJson(w, resp.Base(code.CodeInvalidToken))
		return
	}

	files, err := filelogic.ListRagFiles(username)
	if err != nil {
		log.Println("ListRagFiles fail", err)
		httpx.OkJson(w, resp.Base(code.CodeServerBusy))
		return
	}

	items := make([]types.RagFileInfo, 0, len(files))
	for _, f := range files {
		items = append(items, types.RagFileInfo{
			FileId:   f.FileID,
			FileName: f.FileName,
		})
	}
	httpx.OkJson(w, types.ListRagFilesResp{BaseResp: resp.Success(), Files: items})
}
