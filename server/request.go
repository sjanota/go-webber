package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"net/http"
)

type Request interface {
	ParseForm(v interface{})
	HandleError(error)
	StringVar(name string) string
	ParseVar(name string, parse func(string) error)
	ParseQuery(name string, parse func(string) error)
	ParseQueries(name string, parse func(string) error)

	Session() *sessions.Session
	SetUserError(msg string)
	UserError() string

	Redirect(path string)
	Response(func(w http.ResponseWriter) error)
	Error(statusCode int, msg string)
	Errorf(statusCode int, format string, args ...interface{})
}

type request struct {
	session *sessions.Session
	writer  http.ResponseWriter
	request *http.Request
}

type requestHandled struct{}

func (r *request) StringVar(name string) string {
	vars := mux.Vars(r.request)
	return vars[name]
}

func (r *request) ParseVar(name string, parse func(string) error) {
	val := r.StringVar(name)
	err := parse(val)
	if err != nil {
		r.Errorf(http.StatusBadRequest, "Malformed parameter %s", name)
		panic(requestHandled{})
	}
}

func (r *request) ParseQuery(name string, parse func(string) error) {
	val := r.request.URL.Query().Get(name)
	if val == "" {
		return
	}
	err := parse(val)
	if err != nil {
		r.Errorf(http.StatusBadRequest, "Malformed query %s", name)
		panic(requestHandled{})
	}
	return
}

func (r *request) ParseQueries(name string, parse func(string) error) {
	vals := r.request.URL.Query()[name]
	for _, val := range vals {
		err := parse(val)
		if err != nil {
			r.Errorf(http.StatusBadRequest, "Malformed query %s", name)
			panic(requestHandled{})
		}
	}
	return
}

func (r *request) HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func (r *request) Session() *sessions.Session {
	if r.session == nil {
		var err error
		r.session, err = store.Get(r.request, sessionCookieName)
		if r.session == nil {
			panic(err)
		}
	}
	return r.session
}

func (r *request) ParseForm(v interface{}) {
	err := r.request.ParseForm()
	if err != nil {
		r.Error(http.StatusBadRequest, "Malformed form data")
		panic(requestHandled{})
	}

	decoder := schema.NewDecoder()
	err = decoder.Decode(v, r.request.PostForm)
	if err != nil {
		r.Error(http.StatusUnprocessableEntity, "Invalid form data")
		panic(requestHandled{})
	}
}

func (r *request) SetUserError(msg string) {
	r.Session().AddFlash(msg)
}

func (r *request) UserError() string {
	flashes := r.Session().Flashes()
	if len(flashes) == 0 {
		return ""
	}
	return flashes[0].(string)
}

func (r *request) Redirect(path string) {
	r.saveSession()
	http.Redirect(r.writer, r.request, path, http.StatusSeeOther)
}

func (r *request) Response(rsp func(w http.ResponseWriter) error) {
	r.saveSession()
	err := rsp(r.writer)
	r.HandleError(err)
}

func (r *request) Error(statusCode int, msg string) {
	r.saveSession()
	http.Error(r.writer, msg, statusCode)
}

func (r *request) Errorf(statusCode int, format string, args ...interface{}) {
	r.saveSession()
	http.Error(r.writer, fmt.Sprintf(format, args...), statusCode)
}

func (r *request) saveSession() {
	if r.session != nil {
		_ = r.session.Save(r.request, r.writer)
	}
}
