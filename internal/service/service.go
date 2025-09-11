package service

import (
	"encoding/json"
	"io"
	"log"
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

type WhitelistRequest struct {
	IP    string `json:"ip"`
	Force bool   `json:"force"`
}

type WhitelistReponse struct {
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
	log.Print("\tLogin: ", req.Login, "\n\t\t\tPassword: ", req.Password, "\n\t\t\tIP: ", req.IP)
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
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req WhitelistRequest
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
	res, err := s.store.AddToWhitelist(*ipnet, req.Force)
	if err != nil {
		json.NewEncoder(w).Encode(WhitelistReponse{Ok: res, Reason: err.Error()})
		return
	}
	json.NewEncoder(w).Encode(WhitelistReponse{Ok: res})
}

func (s *Service) BlacklistHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"result":"blacklist stub"}`))
}
