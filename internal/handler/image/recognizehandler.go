package image

import (
	"SamaraAI/common/code"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/types"
	imagesvc "SamaraAI/service/image"
	"log"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RecognizeImageHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpx.OkJson(w, resp.Base(code.CodeInvalidParams))
		return
	}

	_, header, err := r.FormFile("image")
	if err != nil {
		log.Println("FormFile fail", err)
		httpx.OkJson(w, resp.Base(code.CodeInvalidParams))
		return
	}

	className, err := imagesvc.RecognizeImage(header)
	if err != nil {
		log.Println("RecognizeImage fail", err)
		httpx.OkJson(w, resp.Base(code.CodeServerBusy))
		return
	}

	httpx.OkJson(w, types.RecognizeImageResp{BaseResp: resp.Success(), ClassName: className})
}
