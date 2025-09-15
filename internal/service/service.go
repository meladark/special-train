package service

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/meladark/special-train/internal/bucket"
	"github.com/meladark/special-train/internal/storage"
)

type Service struct {
	store storage.Storage
	rl    *bucket.RateLimiter
}

func New(store storage.Storage, bucket *bucket.RateLimiter) *Service {
	return &Service{store: store, rl: bucket}
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

type ListRequest struct {
	IP    string `json:"ip"`
	Force bool   `json:"force"`
}

type ListReponse struct {
	Ok     bool   `json:"ok"`
	Reason string `json:"reason,omitempty"`
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
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
	log.Print("\tLogin: ", req.Login, "\n\t\t\tPassword: ", req.Password, "\n\t\t\tIP: ", req.IP)
	ip := net.ParseIP(req.IP)
	if ip == nil {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	}
	if s.store.InWhitelist(ip) {
		writeJSON(w, AuthorizeResponse{Ok: true})
		return
	}
	if s.store.InBlacklist(ip) {
		writeJSON(w, AuthorizeResponse{Ok: false, Reason: "ip in blacklist"})
		return
	}
	ctx := context.Background()
	allow, stat, err := s.rl.CheckAll(ctx, req.Login, req.Password, req.IP)
	log.Print("\tLogin: ", stat["login"], "\n\t\t\tPassword: ", stat["pass"], "\n\t\t\tIP: ", stat["ip"])
	if err != nil {
		http.Error(w, "service error: "+err.Error(), http.StatusMethodNotAllowed)
		return
	}
	if !allow {
		writeJSON(w, AuthorizeResponse{Ok: false, Reason: "rate limit exceeded"})
		return
	}
	writeJSON(w, AuthorizeResponse{Ok: true})
}

func (s *Service) ResetBucketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := context.Background()
	err := s.rl.ResetAll(ctx, "*")
	if err != nil {
		http.Error(w, "service error: "+err.Error(), http.StatusMethodNotAllowed)
	}
	writeJSON(w, AuthorizeResponse{Ok: true})
}

func (s *Service) ResetBucketIPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err == io.EOF {
			http.Error(w, "Empty field", http.StatusBadRequest)
			return
		}
	}
	log.Print(req.IP)
	if ip := net.ParseIP(req.IP); ip != nil {
		if err := s.rl.ResetIP(context.Background(), req.IP); err != nil {
			http.Error(w, "service error: "+err.Error(), http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, AuthorizeResponse{Ok: true})
		return
	}
	http.Error(w, "invalid ip", http.StatusBadRequest)
}

func (s *Service) ResetBucketLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err == io.EOF {
			http.Error(w, "Empty field", http.StatusBadRequest)
			return
		}
	}
	log.Print(req.Login)
	if err := s.rl.ResetLogin(context.Background(), req.Login); err != nil {
		http.Error(w, "service error: "+err.Error(), http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, AuthorizeResponse{Ok: true})
}

func (s *Service) WhitelistHandler(w http.ResponseWriter, r *http.Request) {
	s.handleListOperation(w, r, s.store.AddToWhitelist)
}

func (s *Service) BlacklistHandler(w http.ResponseWriter, r *http.Request) {
	s.handleListOperation(w, r, s.store.AddToBlacklist)
}

func (s *Service) handleListOperation(
	w http.ResponseWriter,
	r *http.Request,
	addFunc func(net.IPNet, bool) (bool, error),
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err == io.EOF {
			http.Error(w, "Empty field", http.StatusBadRequest)
			return
		}
	}
	log.Print(req.IP, req.Force)
	_, ipnet, err := net.ParseCIDR(req.IP)
	if err != nil {
		if ip := net.ParseIP(req.IP); ip != nil {
			ipnet = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	res, err := addFunc(*ipnet, req.Force)
	if err != nil {
		writeJSON(w, ListReponse{Ok: res, Reason: err.Error()})
		return
	}
	writeJSON(w, ListReponse{Ok: res})
}
