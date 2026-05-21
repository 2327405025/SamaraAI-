package file

import (
	"SamaraAI/common/code"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/middleware"
	"SamaraAI/internal/types"
	filesvc "SamaraAI/service/file"
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadRagFileHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpx.OkJson(w, resp.Base(code.CodeInvalidParams))
		return
	}

	_, header, err := r.FormFile("file")
	if err != nil {
		log.Println("FormFile fail", err)
		httpx.OkJson(w, resp.Base(code.CodeInvalidParams))
		return
	}

	username := middleware.UserNameFromContext(r.Context())
	if username == "" {
		httpx.OkJson(w, resp.Base(code.CodeInvalidToken))
		return
	}

	filePath, err := filesvc.UploadRagFile(username, header)
	if err != nil {
		log.Println("UploadFile fail", err)
		httpx.OkJson(w, resp.Base(code.CodeServerBusy))
		return
	}

	httpx.OkJson(w, types.UploadFileResp{BaseResp: resp.Success(), FilePath: filePath})
}
