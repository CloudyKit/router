// Copyright 2016 José Santos <henrique_1609@me.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"fmt"
	"net/http"
)

type Handler func(http.ResponseWriter, *http.Request, Parameter)

type Router struct {
	trees map[string]*routeNode
}

func New() *Router {
	return &Router{trees: make(map[string]*routeNode)}
}

func splitURLpath(path string) (parts []string, names map[string]int) {

	var (
		nameidx      int = -1
		partidx      int
		paramCounter int
	)

	for i := 0; i < len(path); i++ {
		// recording name
		if nameidx != -1 {
			//found /
			if path[i] == '/' {

				if names == nil {
					names = make(map[string]int)
				}

				names[path[nameidx:i]] = paramCounter
				paramCounter++

				nameidx = -1 // switch to normal recording
				partidx = i
			}
		} else {
			if path[i] == ':' || path[i] == '*' {
				if path[i-1] != '/' {
					panic(fmt.Errorf("Inválid parameter : or * comes anwais after / - %q", path))
				}
				nameidx = i + 1
				if partidx != i {
					parts = append(parts, path[partidx:i])
				}
				parts = append(parts, path[i:nameidx])
			}
		}
	}

	if nameidx != -1 {
		if names == nil {
			names = make(map[string]int)
		}
		names[path[nameidx:]] = paramCounter
		paramCounter++
	} else if partidx < len(path) {
		parts = append(parts, path[partidx:])
	}
	return
}

func (router *Router) Finalize() {
	for _, _node := range router.trees {
		_node.finalize()
	}
}

func (router *Router) FindRoute(method string, path string) (Handler, Parameter) {
	_node := router.trees[method]
	if _node == nil {
		return nil, Parameter{}
	}
	fn, wildcard := _node.findRoute(path)

	if fn != nil {
		return fn.handler, Parameter{routeNode: fn, path: path, wildcard: wildcard}
	}
	return nil, Parameter{}
}

func (router *Router) AddRoute(method string, path string, fn Handler) {
	parts, names := splitURLpath(path)
	_node := router.trees[method]
	if _node == nil {
		_node = &routeNode{}
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
	if handler != nil {
		handler(w, r, variables)
	} else {
		http.NotFound(w, r)
	}
}
