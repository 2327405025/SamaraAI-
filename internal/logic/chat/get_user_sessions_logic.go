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
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSessionsLogic {
	return &GetUserSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSessionsLogic) GetUserSessions() (*types.GetUserSessionsResp, error) {
	userName := middleware.UserNameFromContext(l.ctx)
	userSessions, err := getUserSessionsByUserName(userName)
	if err != nil {
		return &types.GetUserSessionsResp{BaseResp: resp.Base(code.CodeServerBusy)}, nil
	}

	sessions := make([]types.SessionInfo, 0, len(userSessions))
	for _, s := range userSessions {
		sessions = append(sessions, types.SessionInfo{
			SessionId: s.SessionID,
			Name:      s.Title,
		})
	}
	return &types.GetUserSessionsResp{BaseResp: resp.Success(), Sessions: sessions}, nil
}
