package chat

import (
	"SamaraAI/common/aihelper"
	"SamaraAI/common/code"
	"SamaraAI/common/sse"
	"SamaraAI/dao/message"
	sessiondao "SamaraAI/dao/session"
	"SamaraAI/internal/config"
	"SamaraAI/model"
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func aiHelperConfig(userName string) map[string]interface{} {
	cfg := config.Get()
	return map[string]interface{}{
		"apiKey":   cfg.OpenAI.ApiKey,
		"username": userName,
	}
}

func getUserSessionsByUserName(userName string) ([]model.SessionInfo, error) {
	sessions, err := sessiondao.GetSessionsByUserName(userName)
	if err != nil {
		return nil, err
	}
	out := make([]model.SessionInfo, 0, len(sessions))
	for _, s := range sessions {
		title := s.Title
		if title == "" {
			title = s.ID
		}
		out = append(out, model.SessionInfo{
			SessionID: s.ID,
			Title:     title,
		})
	}
	return out, nil
}

func CreateStreamSessionOnly(userName, userQuestion string) (string, code.Code) {
	newSession := &model.Session{
		ID:       uuid.New().String(),
		UserName: userName,
		Title:    userQuestion,
	}
	createdSession, err := sessiondao.CreateSession(newSession)
	if err != nil {
		log.Println("CreateStreamSessionOnly CreateSession error:", err)
		return "", code.CodeServerBusy
	}
	return createdSession.ID, code.CodeSuccess
}

func StreamMessageToExistingSession(ctx context.Context, userName, sessionID, userQuestion, modelType string, writer http.ResponseWriter) code.Code {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		log.Println("StreamMessageToExistingSession: streaming unsupported")
		return code.CodeServerBusy
	}

	manager := aihelper.GetGlobalManager()
	helper, err := manager.GetOrCreateAIHelper(userName, sessionID, modelType, aiHelperConfig(userName))
	if err != nil {
		log.Println("StreamMessageToExistingSession GetOrCreateAIHelper error:", err)
		return code.AIModelFail
	}

	sse.WriteComment(writer, flusher, "connected")

	cb := func(msg string) {
		if ctx.Err() != nil {
			return
		}
		sse.WriteData(writer, flusher, msg)
	}

	if _, err := helper.StreamResponse(userName, ctx, cb, userQuestion); err != nil {
		if ctx.Err() != nil {
			return code.CodeSuccess
		}
		log.Println("StreamMessageToExistingSession StreamResponse error:", err)
		return code.AIModelFail
	}

	sse.WriteDone(writer, flusher)
	return code.CodeSuccess
}

func getChatHistory(userName, sessionID string) ([]model.History, code.Code) {
	manager := aihelper.GetGlobalManager()
	if helper, exists := manager.GetAIHelper(userName, sessionID); exists {
		return messagesToHistory(helper.GetMessages()), code.CodeSuccess
	}

	msgs, err := message.GetMessagesBySessionID(sessionID)
	if err != nil {
		log.Println("getChatHistory DB error:", err)
		return nil, code.CodeServerBusy
	}
	if len(msgs) > 0 && msgs[0].UserName != userName {
		return nil, code.CodeInvalidToken
	}
	history := make([]model.History, 0, len(msgs))
	for i := range msgs {
		history = append(history, model.History{
			IsUser:  msgs[i].IsUser,
			Content: msgs[i].Content,
		})
	}
	return history, code.CodeSuccess
}

func messagesToHistory(msgs []*model.Message) []model.History {
	history := make([]model.History, 0, len(msgs))
	for _, msg := range msgs {
		history = append(history, model.History{
			IsUser:  msg.IsUser,
			Content: msg.Content,
		})
	}
	return history
}

func ChatStreamSend(ctx context.Context, userName, sessionID, userQuestion, modelType string, writer http.ResponseWriter) code.Code {
	return StreamMessageToExistingSession(ctx, userName, sessionID, userQuestion, modelType, writer)
}

func deleteSession(userName, sessionID string) code.Code {
	if sessionID == "" {
		return code.CodeInvalidParams
	}
	if _, err := sessiondao.GetSessionByIDAndUser(sessionID, userName); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.CodeRecordNotFound
		}
		log.Println("deleteSession GetSession error:", err)
		return code.CodeServerBusy
	}
	if err := message.DeleteMessagesBySessionID(sessionID); err != nil {
		log.Println("deleteSession DeleteMessages error:", err)
		return code.CodeServerBusy
	}
	if err := sessiondao.DeleteSession(sessionID, userName); err != nil {
		log.Println("deleteSession DeleteSession error:", err)
		return code.CodeServerBusy
	}
	aihelper.GetGlobalManager().RemoveAIHelper(userName, sessionID)
	return code.CodeSuccess
}
