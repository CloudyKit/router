package router

import (
	"net/http"
)

type Router struct {
	trees map[string]*node
}

func New() *Router {
	return &Router{trees: make(map[string]*node)}
}

func split(path string) (parts []string, names []string) {

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

func (tree *Router) Lookup(method string, path string) (func(http.ResponseWriter, *http.Request, Variables), Variables) {
	_node := tree.trees[method]

	if _node == nil {
		return nil, Variables{}
	}

	fn, names, values := _node.getHandler(path, 0)
	return fn, Variables{Keys: names, Values: values}
}

func (tree *Router) Handle(method string, path string, fn func(http.ResponseWriter, *http.Request, Variables)) {
	parts, names := split(path)
	_node := tree.trees[method]
	if _node == nil {
		_node = &node{}
		tree.trees[method] = _node
	}
	_node.addNode(parts, names, fn)
	_node.sort()
}

func (tree *Router) String() string {
	var lines string
	for method, _node := range tree.trees {
		lines += method + " " + _node.String()
	}
	return lines
}
