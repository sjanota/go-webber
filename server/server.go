package server

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"net/http"
)

type Server struct {
	router *mux.Router
}

const (
	sessionCookieName = "cookhub-session"
)

var store = sessions.NewCookieStore([]byte("key"))

func New() *Server {
	s := &Server{router: mux.NewRouter()}
	return s
}

func (s Server) Start(handlers ...*Controller) error {
	for _, h := range handlers {
		s.addHandler(h)
	}
	return http.ListenAndServe("localhost:8080", s.router)
}

func (s Server) addHandler(h *Controller) {
	if h.Get != nil {
		s.router.
			Path(h.Path).
			Methods(http.MethodGet).
			HandlerFunc(s.wrapHandlerFunc(h.Get))
	}

	if h.Post != nil {
		s.router.
			Path(h.Path).
			Methods(http.MethodPost).
			HandlerFunc(s.wrapHandlerFunc(h.Post))
	}
}

func (s Server) wrapHandlerFunc(hf func(r Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			switch err := recover(); err.(type) {
			case requestHandled:
				return
			default:
				if err != nil {
					panic(err)
				}
			}
		}()
		req := &request{writer: w, request: r}
		hf(req)
	}
}

type Controller struct {
	Path string
	Get  func(r Request)
	Post func(r Request)
}
