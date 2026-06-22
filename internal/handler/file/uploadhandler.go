package file

import (
	"SamaraAI/common/code"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/middleware"
	"SamaraAI/internal/types"
	filelogic "SamaraAI/internal/logic/file"
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

	storedName, displayName, err := filelogic.UploadRagFile(username, header)
	if err != nil {
		log.Println("UploadFile fail", err)
		httpx.OkJson(w, resp.Base(code.CodeServerBusy))
		return
	}

	httpx.OkJson(w, types.UploadFileResp{
		BaseResp: resp.Success(),
		FileId:   storedName,
		FileName: displayName,
	})
}
