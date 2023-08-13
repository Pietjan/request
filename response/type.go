package response

import "net/http"

type Handler interface {
	Handle(res *http.Response) error
	When(func(res *http.Response) bool) Handler
}

type HandlerFn = func(res *http.Response) error
type Condition = func(res *http.Response) bool
