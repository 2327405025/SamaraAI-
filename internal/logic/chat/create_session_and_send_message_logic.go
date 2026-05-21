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

type CreateSessionAndSendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSessionAndSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSessionAndSendMessageLogic {
	return &CreateSessionAndSendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSessionAndSendMessageLogic) CreateSessionAndSendMessage(req *types.ChatQuestionReq) (*types.CreateSessionAndSendMessageResp, error) {
	userName := middleware.UserNameFromContext(l.ctx)
	sessionID, aiInformation, code_ := sessionsvc.CreateSessionAndSendMessage(userName, req.Question, req.ModelType)
	if code_ != code.CodeSuccess {
		return &types.CreateSessionAndSendMessageResp{BaseResp: resp.Base(code_)}, nil
	}
	return &types.CreateSessionAndSendMessageResp{
		BaseResp:    resp.Success(),
		Information: aiInformation,
		SessionId:   sessionID,
	}, nil
}
