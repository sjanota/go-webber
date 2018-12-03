package server

import (
	"github.com/sjanota/webber/view"
	"net/http"
)

type Controller struct {
	Path string
	Get  func(r Request)
	Post func(r Request)
}

func IndexController(renderer view.Renderer) *Controller {
	return &Controller{
		Path: "/",
		Get: func(r Request) {
			r.Response(func(w http.ResponseWriter) error {
				return renderer.RenderIndex(w)
			})
		},
	}
}
