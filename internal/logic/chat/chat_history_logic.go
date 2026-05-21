// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package chat

import (
	"context"

	"SamaraAI/common/code"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/middleware"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"
	sessionsvc "SamaraAI/service/session"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatHistoryLogic {
	return &ChatHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatHistoryLogic) ChatHistory(req *types.ChatHistoryReq) (*types.ChatHistoryResp, error) {
	userName := middleware.UserNameFromContext(l.ctx)
	history, code_ := sessionsvc.GetChatHistory(userName, req.SessionId)
	if code_ != code.CodeSuccess {
		return &types.ChatHistoryResp{BaseResp: resp.Base(code_)}, nil
	}

	items := make([]types.HistoryItem, 0, len(history))
	for _, h := range history {
		items = append(items, types.HistoryItem{
			IsUser:  h.IsUser,
			Content: h.Content,
		})
	}
	return &types.ChatHistoryResp{BaseResp: resp.Success(), History: items}, nil
}
