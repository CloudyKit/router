package router

import (
	"net/http"
)

type Handler func(http.ResponseWriter, *http.Request, Values)

type Router struct {
	PanicHandler        func(http.ResponseWriter, *http.Request, Values, interface{})
	NotFoundHandler     Handler
	RequestStartHandler Handler
	RequestEndHandler   Handler

	trees               map[string]*node
}

func New() *Router {
	return &Router{trees: make(map[string]*node)}
}

func explode(path string) (parts []string, names []string) {

	var (
		nameidx int = -1
		partidx int = 0
	)

	for i := 0; i < len(path); i++ {

		// recording name
		if nameidx != -1 {
			//found /
			if path[i] == '/' {
				names = append(names, path[nameidx:i])
				nameidx = -1 // switch to normal recording
				partidx = i
			}
		} else {

			if path[i] == ':' || path[i] == '*' {

				nameidx = i + 1
				if partidx != i {
					parts = append(parts, path[partidx:i])
				}
				parts = append(parts, path[i:nameidx])

			}

		}

	}

	if nameidx != -1 {
		names = append(names, path[nameidx:])
	} else if partidx < len(path) {
		parts = append(parts, path[partidx:])
	}

	return
}

func (router *Router) FindRoute(method string, path string) (Handler, Values) {
	_node := router.trees[method]
	if _node == nil {
		return nil, Values{}
	}
	fn, names, values := _node.findRoute(path, 0)
	return fn, Values{Keys: names, Values: values}
}

func (router *Router) AddRoute(method string, path string, fn Handler) {
	parts, names := explode(path)
	_node := router.trees[method]
	if _node == nil {
		_node = &node{}
		router.trees[method] = _node
	}
	_node.addRoute(parts, names, fn)
	_node.optimizeRoutes()
}

func (router *Router) String() string {
	var lines string
	for method, _node := range router.trees {
		lines += method + " " + _node.String()
	}
	return lines
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, variables := router.FindRoute(r.Method, r.URL.Path)

	defer func() {
		if router.PanicHandler != nil {
			if err := recover(); err != nil {
				router.PanicHandler(w, r, variables, err)
			}
		}
		if router.RequestEndHandler != nil {
			router.RequestEndHandler(w, r, variables)
		}
	}()

	if router.RequestStartHandler != nil {
		router.RequestStartHandler(w, r, variables)
	}

	if handler != nil {
		handler(w, r, variables)
	} else if router.NotFoundHandler != nil {
		router.NotFoundHandler(w, r, variables)
	} else {
		http.NotFound(w, r)
	}

}
