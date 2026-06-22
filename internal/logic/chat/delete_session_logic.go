package chat

import (
	"context"

	"SamaraAI/common/code"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/middleware"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSessionLogic {
	return &DeleteSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSessionLogic) DeleteSession(req *types.DeleteSessionReq) (*types.DeleteSessionResp, error) {
	userName := middleware.UserNameFromContext(l.ctx)
	code_ := deleteSession(userName, req.SessionId)
	if code_ != code.CodeSuccess {
		return &types.DeleteSessionResp{BaseResp: resp.Base(code_)}, nil
	}
	return &types.DeleteSessionResp{BaseResp: resp.Success()}, nil
}
