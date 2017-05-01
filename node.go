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
	"net/http"
	"sort"
	"strings"
)

type routeNode struct {
	text    string
	names   map[string]int
	handler Handler

	parent   *routeNode
	wildcard *routeNode
	colon    *routeNode

	nodes   []*routeNode
	start   byte
	max     byte
	indices []uint8
}

func (node *routeNode) nextRoute(path string) (*routeNode, int8, int) {

	if path == "*" {
		if node.wildcard == nil {
			node.wildcard = &routeNode{text: "*"}
		}
		return node.wildcard, 0, 0
	}

	if path == ":" {
		if node.colon == nil {
			node.colon = &routeNode{text: ":"}
		}
		return node.colon, 0, 0
	}

	for i := 0; i < len(node.nodes); i++ {
		cNode := node.nodes[i]
		if cNode.text[0] == path[0] {

			var max = len(cNode.text)
			var lpath = len(path)
			var pathIsBigger int8

			if lpath > max {
				pathIsBigger = 1
			} else if lpath < max {
				max = lpath
				pathIsBigger = -1
			}

			for j := 0; j < max; j++ {
				if path[j] != cNode.text[j] {
					ccNode := &routeNode{text: path[0:j], nodes: []*routeNode{cNode, &routeNode{text: path[j:]}}}
					cNode.text = cNode.text[j:]
					node.nodes[i] = ccNode
					return ccNode.nodes[1], 0, i
				}
			}

			return cNode, pathIsBigger, i
		}
	}

	return nil, 0, 0
}

func (node *routeNode) addRoute(parts []string, names map[string]int, handler Handler) {

	var (
		ccNode *routeNode
		cNode  *routeNode
	)

	cNode, result, idx := node.nextRoute(parts[0])

RESTART:
	if cNode == nil {
		cNode = &routeNode{text: parts[0]}
		node.nodes = append(node.nodes, cNode)
	} else if result == 1 {
		//
		parts[0] = parts[0][len(cNode.text):]
		ccNode, result, idx = cNode.nextRoute(parts[0])
		if cNode != nil {
			node = cNode
			cNode = ccNode
			goto RESTART
		}
		ccNode := &routeNode{text: parts[0]}
		cNode.nodes = append(node.nodes, ccNode)
		cNode = ccNode
	} else if result == -1 {
		ccNode := &routeNode{text: parts[0]}
		cNode.text = cNode.text[len(ccNode.text):]
		ccNode.nodes = []*routeNode{cNode}
		node.nodes[idx] = ccNode
		cNode = ccNode
	}

	if len(parts) == 1 {
		cNode.handler = handler
		cNode.names = names
		return
	}

	cNode.addRoute(parts[1:], names, handler)
}

var redirectNode = &routeNode{
	handler: func(w http.ResponseWriter, r *http.Request, p Parameter) {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
	},
}

func (node *routeNode) findRoute(urlPath string) (*routeNode, int) {

	urlByte := urlPath[0]
	pathLen := len(urlPath)

	if urlByte >= node.start && urlByte <= node.max {
		if i := node.indices[urlByte-node.start]; i != 0 {
			cNode := node.nodes[i-1]
			nodeLen := len(cNode.text)
			if nodeLen < pathLen {
				if cNode.text == urlPath[0:nodeLen] {
					if cNode, wildcard := cNode.findRoute(urlPath[nodeLen:]); cNode != nil {
						return cNode, wildcard
					}
				}
			} else if cNode.text == urlPath {
				if cNode.handler == nil && cNode.wildcard != nil {
					return cNode.wildcard, 0
				}
				return cNode, 0
			} else if nodeLen == pathLen+1 && cNode.text[nodeLen-1] == '/' {
				return redirectNode, 0
			}
		}
	}

	if node.colon != nil && pathLen != 0 {
		ix := strings.IndexByte(urlPath, '/')
		if ix > 0 {
			if cNode, wildcard := node.colon.findRoute(urlPath[ix:]); cNode != nil {
				return cNode, wildcard
			}
		} else if node.colon.handler != nil {
			return node.colon, 0
		}
	}

	if node.wildcard != nil {
		return node.wildcard, pathLen
	}

	return nil, 0
}

func (node *routeNode) optimizeRoutes() {

	if len(node.nodes) > 0 {

		sort.Slice(node.nodes, func(i, j int) bool {
			return node.nodes[i].text[0] < node.nodes[j].text[0]
		})

		for i := 0; i < len(node.indices); i++ {
			node.indices[i] = 0
		}

		node.start = node.nodes[0].text[0]
		node.max = node.nodes[len(node.nodes)-1].text[0]

		for i := 0; i < len(node.nodes); i++ {
			cNode := node.nodes[i]
			cNode.parent = node

			cByte := int(cNode.text[0] - node.start)
			if cByte >= len(node.indices) {
				node.indices = append(node.indices, make([]uint8, cByte+1-len(node.indices))...)
			}
			node.indices[cByte] = uint8(i + 1)
			cNode.optimizeRoutes()
		}
	}

	if node.colon != nil {
		node.colon.parent = node
		node.colon.optimizeRoutes()
	}

	if node.wildcard != nil {
		node.wildcard.parent = node
		node.wildcard.optimizeRoutes()
	}
}

func (node *routeNode) finalize() {
	if len(node.nodes) > 0 {
		for i := 0; i < len(node.nodes); i++ {
			node.nodes[i].finalize()
		}
	}
	if node.colon != nil {
		node.colon.finalize()
	}
	if node.wildcard != nil {
		node.wildcard.finalize()
	}
	*node = routeNode{}
}

func (node *routeNode) string(col int) string {
	var str = "\n" + strings.Repeat(" ", col) + node.text + " -> "
	col += len(node.text) + 4
	for i := 0; i < len(node.indices); i++ {
		if j := node.indices[i]; j != 0 {
			str += node.nodes[j-1].string(col)
		}
	}
	if node.colon != nil {
		str += node.colon.string(col)
	}
	if node.wildcard != nil {
		str += node.wildcard.string(col)
	}
	return str
}

func (node *routeNode) String() string {
	if node.text == "" {
		return node.string(0)
	}
	col := len(node.text) + 4
	return node.text + " -> " + node.string(col)
}
