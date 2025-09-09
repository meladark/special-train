package service

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/meladark/special-train/internal/storage"
)

type Service struct {
	store storage.Storage
}

func New(store storage.Storage) *Service {
	return &Service{store: store}
}

type AuthorizeRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

type AuthorizeResponse struct {
	Ok     bool   `json:"ok"`
	Reason string `json:"reason,omitempty"`
}

func (s *Service) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	ip := net.ParseIP(req.IP)
	if ip == nil {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	}
	if s.store.InWhitelist(ip) {
		json.NewEncoder(w).Encode(AuthorizeResponse{Ok: true})
		return
	}
	if s.store.InBlacklist(ip) {
		json.NewEncoder(w).Encode(AuthorizeResponse{Ok: false, Reason: "ip in blacklist"})
		return
	}
	// TODO: redis buckets
	json.NewEncoder(w).Encode(AuthorizeResponse{Ok: true})
}

func (s *Service) ResetBucketHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"result":"reset stub"}`))
}

func (s *Service) WhitelistHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"result":"whitelist stub"}`))
}

func (s *Service) BlacklistHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"result":"blacklist stub"}`))
}
