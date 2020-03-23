package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/evgeniuz/shortener/shortener/store"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type Shortener struct {
	store store.Store
}

func NewShortener(s store.Store) (*Shortener, error) {
	return &Shortener{s}, nil
}

func (s *Shortener) router() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/{hash}/stats", s.statsHandler)
	r.Get("/{hash}", s.redirectHandler)
	r.Post("/", s.shortenHandler)

	return r
}

type statsResponse struct {
	Day   uint64 `json:"day"`
	Week  uint64 `json:"week"`
	Total uint64 `json:"total"`
}

func (s *Shortener) statsHandler(res http.ResponseWriter, req *http.Request) {
	hash := chi.URLParam(req, "hash")

	stats, err := s.store.Stats(hash)
	if err != nil {
		renderError(res, err)
		return
	}

	renderSuccess(res, statsResponse{stats.Day, stats.Week, stats.Total})
}

func (s *Shortener) redirectHandler(res http.ResponseWriter, req *http.Request) {
	hash := chi.URLParam(req, "hash")

	url, err := s.store.Get(hash)
	if err != nil {
		renderError(res, err)
	}

	if url == "" {
		http.NotFound(res, req)
		return
	}

	http.Redirect(res, req, url, http.StatusMovedPermanently)

	go func() {
		err := s.store.Visit(hash)
		if err != nil {
			log.Printf("cannot register visit: %v", err)
		}
	}()
}

type shortenRequest struct {
	Url string `json:"url"`
}

type shortenResponse struct {
	Hash string `json:"hash"`
}

func (s *Shortener) shortenHandler(res http.ResponseWriter, req *http.Request) {
	var sr shortenRequest

	d := json.NewDecoder(req.Body)
	err := d.Decode(&sr)
	if err != nil {
		renderError(res, err)
		return
	}

	hash, err := s.store.Set(sr.Url)
	if err != nil {
		renderError(res, err)
		return
	}

	renderSuccess(res, shortenResponse{hash})
}

func (s *Shortener) Listen(port int) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", port), s.router())
}

type errorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type successResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func renderError(res http.ResponseWriter, e error) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusInternalServerError)

	log.Printf("server error: %v", e)

	err := json.NewEncoder(res).Encode(errorResponse{false, e.Error()})
	if err != nil {
		log.Printf("cannot render error: %v", err)
	}
}

func renderSuccess(res http.ResponseWriter, data interface{}) {
	res.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(res).Encode(successResponse{true, data})
	if err != nil {
		log.Printf("cannot render success: %v", err)
	}
}
