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
	mux.HandleFunc("/api/whitelist", svc.WhitelistHandler)
	mux.HandleFunc("/api/blacklist", svc.BlacklistHandler)
	return loggingMiddleware(mux)
}
