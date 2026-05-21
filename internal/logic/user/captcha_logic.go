// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"SamaraAI/common/code"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"
	usersvc "SamaraAI/service/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaLogic {
	return &CaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaptchaLogic) Captcha(req *types.CaptchaReq) (*types.CaptchaResp, error) {
	code_ := usersvc.SendCaptcha(req.Email)
	if code_ != code.CodeSuccess {
		return &types.CaptchaResp{BaseResp: resp.Base(code_)}, nil
	}
	return &types.CaptchaResp{BaseResp: resp.Success()}, nil
}
