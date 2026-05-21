package chat

import (
	"SamaraAI/common/code"
	"SamaraAI/internal/middleware"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"
	sessionsvc "SamaraAI/service/session"
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

		sessionID, code_ := sessionsvc.CreateStreamSessionOnly(userName, req.Question)
		if code_ != code.CodeSuccess {
			fmt.Fprintf(w, "event: error\ndata: {\"message\":\"Failed to create session\"}\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return
		}

		fmt.Fprintf(w, "data: {\"sessionId\": \"%s\"}\n\n", sessionID)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		code_ = sessionsvc.StreamMessageToExistingSession(userName, sessionID, req.Question, req.ModelType, w)
		if code_ != code.CodeSuccess {
			fmt.Fprintf(w, "event: error\ndata: {\"message\":\"Failed to send message\"}\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
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

		code_ := sessionsvc.ChatStreamSend(userName, req.SessionId, req.Question, req.ModelType, w)
		if code_ != code.CodeSuccess {
			fmt.Fprintf(w, "event: error\ndata: {\"message\":\"Failed to send message\"}\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}
