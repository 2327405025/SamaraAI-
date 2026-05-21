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

type ChatSendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatSendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatSendLogic {
	return &ChatSendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatSendLogic) ChatSend(req *types.ChatSendReq) (*types.ChatSendResp, error) {
	userName := middleware.UserNameFromContext(l.ctx)
	aiInformation, code_ := sessionsvc.ChatSend(userName, req.SessionId, req.Question, req.ModelType)
	if code_ != code.CodeSuccess {
		return &types.ChatSendResp{BaseResp: resp.Base(code_)}, nil
	}
	return &types.ChatSendResp{BaseResp: resp.Success(), Information: aiInformation}, nil
}
