package file

import (
	"SamaraAI/common/code"
	filelogic "SamaraAI/internal/logic/file"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/middleware"
	"SamaraAI/internal/types"
	"log"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteRagFileHandler(w http.ResponseWriter, r *http.Request) {
	var req types.DeleteRagFileReq
	if err := httpx.Parse(r, &req); err != nil {
		httpx.OkJson(w, resp.Base(code.CodeInvalidParams))
		return
	}

	username := middleware.UserNameFromContext(r.Context())
	if username == "" {
		httpx.OkJson(w, resp.Base(code.CodeInvalidToken))
		return
	}

	if err := filelogic.DeleteRagFile(username, req.FileId); err != nil {
		log.Println("DeleteRagFile fail", err)
		if strings.Contains(err.Error(), "not found") {
			httpx.OkJson(w, resp.Base(code.CodeRecordNotFound))
			return
		}
		httpx.OkJson(w, resp.Base(code.CodeServerBusy))
		return
	}

	httpx.OkJson(w, types.DeleteRagFileResp{BaseResp: resp.Success()})
}
