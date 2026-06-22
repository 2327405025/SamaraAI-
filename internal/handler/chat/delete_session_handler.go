package chat

import (
	"net/http"

	"SamaraAI/internal/logic/chat"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteSessionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteSessionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chat.NewDeleteSessionLogic(r.Context(), svcCtx)
		resp, err := l.DeleteSession(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
