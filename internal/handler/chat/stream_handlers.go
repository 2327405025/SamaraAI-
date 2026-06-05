package chat

import (
	"SamaraAI/common/code"
	"SamaraAI/common/sse"
	chatlogic "SamaraAI/internal/logic/chat"
	"SamaraAI/internal/middleware"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
}

func CreateStreamSessionAndSendMessageHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatQuestionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, fmt.Errorf("invalid parameters"))
			return
		}
		userName := middleware.UserNameFromContext(r.Context())
		setSSEHeaders(w)

		flusher, ok := w.(http.Flusher)
		if !ok {
			httpx.Error(w, fmt.Errorf("streaming unsupported"))
			return
		}
		sse.WriteComment(w, flusher, "connected")

		sessionID, code_ := chatlogic.CreateStreamSessionOnly(userName, req.Question)
		if code_ != code.CodeSuccess {
			sse.WriteData(w, flusher, `{"message":"Failed to create session"}`)
			return
		}

		_ = sse.WriteJSON(w, flusher, map[string]string{"sessionId": sessionID})

		code_ = chatlogic.StreamMessageToExistingSession(r.Context(), userName, sessionID, req.Question, req.ModelType, w)
		if code_ != code.CodeSuccess && r.Context().Err() == nil {
			sse.WriteData(w, flusher, `{"message":"Failed to send message"}`)
		}
	}
}

func ChatStreamSendHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatSendReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, fmt.Errorf("invalid parameters"))
			return
		}
		userName := middleware.UserNameFromContext(r.Context())
		setSSEHeaders(w)

		flusher, ok := w.(http.Flusher)
		if !ok {
			httpx.Error(w, fmt.Errorf("streaming unsupported"))
			return
		}
		sse.WriteComment(w, flusher, "connected")

		code_ := chatlogic.ChatStreamSend(r.Context(), userName, req.SessionId, req.Question, req.ModelType, w)
		if code_ != code.CodeSuccess && r.Context().Err() == nil {
			sse.WriteData(w, flusher, `{"message":"Failed to send message"}`)
		}
	}
}
