package api

import (
	"net/http"

	"github.com/meladark/special-train/internal/service"
)

func NewRouter(svc *service.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/authorize", svc.AuthorizeHandler)
	mux.HandleFunc("/api/bucket/reset", svc.ResetBucketHandler)
	mux.HandleFunc("/api/bucket/reset/ip", svc.ResetBucketIPHandler)
	mux.HandleFunc("/api/bucket/reset/login", svc.ResetBucketLoginHandler)
	mux.HandleFunc("/api/whitelist/add", svc.WhitelistHandler)
	mux.HandleFunc("/api/blacklist/add", svc.BlacklistHandler)
	mux.HandleFunc("/api/whitelist/del", svc.RemoveFromWhitelistHandler)
	mux.HandleFunc("/api/blacklist/del", svc.RemoveFromBlacklistHandler)
	mux.HandleFunc("/api/view/lists", svc.ViewListsHandler)
	return loggingMiddleware(mux)
}
