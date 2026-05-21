// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"net/http"

	"SamaraAI/internal/logic/chat"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateSessionAndSendMessageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatQuestionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := chat.NewCreateSessionAndSendMessageLogic(r.Context(), svcCtx)
		resp, err := l.CreateSessionAndSendMessage(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
