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

package Router

import (
	"sort"
	"strings"
)

type node struct {
	text     string
	names    map[string]int
	handler  Handler

	parent   *node
	wildcard *node
	colon    *node

	nodes    nodes
	start    byte
	max      byte
	indices  []uint8
}

type nodes []*node

func (s nodes) Len() int {
	return len(s)
}

func (s nodes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s nodes) Less(i, j int) bool {
	return s[i].text[0] < s[j].text[0]
}

func (_node *node) nextRoute(path string) (*node, int8, int) {

	if path == "*" {
		if _node.wildcard == nil {
			_node.wildcard = &node{text: "*"}
		}
		return _node.wildcard, 0, 0
	}

	if path == ":" {
		if _node.colon == nil {
			_node.colon = &node{text: ":"}
		}
		return _node.colon, 0, 0
	}

	for i := 0; i < len(_node.nodes); i++ {
		cNode := _node.nodes[i]
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
					ccNode := &node{text: path[0:j], nodes: nodes{cNode, &node{text: path[j:]}}}
					cNode.text = cNode.text[j:]
					_node.nodes[i] = ccNode
					return ccNode.nodes[1], 0, i
				}
			}

			return cNode, pathIsBigger, i
		}
	}

	return nil, 0, 0
}

func (_node *node) addRoute(parts []string, names map[string]int, handler Handler) {

	var (
		ccNode *node
		cNode  *node
	)

	cNode, result, idx := _node.nextRoute(parts[0])

	RESTART:
	if cNode == nil {
		cNode = &node{text: parts[0]}
		_node.nodes = append(_node.nodes, cNode)
	} else if result == 1 {
		//
		parts[0] = parts[0][len(cNode.text):]
		ccNode, result, idx = cNode.nextRoute(parts[0])
		if cNode != nil {
			_node = cNode
			cNode = ccNode
			goto RESTART
		}
		ccNode := &node{text: parts[0]}
		cNode.nodes = append(_node.nodes, ccNode)
		cNode = ccNode
	} else if result == -1 {
		ccNode := &node{text: parts[0]}
		cNode.text = cNode.text[len(ccNode.text):]
		ccNode.nodes = nodes{cNode}
		_node.nodes[idx] = ccNode
		cNode = ccNode
	}

	if len(parts) == 1 {
		cNode.handler = handler
		cNode.names = names
		return
	}

	cNode.addRoute(parts[1:], names, handler)
}

func (_node *node) findRoute(urlPath string) (*node, int) {

	urlByte := urlPath[0]
	pathLen := len(urlPath)

	if urlByte >= _node.start && urlByte <= _node.max {
		if i := _node.indices[urlByte - _node.start]; i != 0 {
			cNode := _node.nodes[i - 1]
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
			}
		}
	}

	if _node.colon != nil && pathLen != 0 {
		ix := strings.IndexByte(urlPath, '/')
		if ix > 0 {
			if cNode, wildcard := _node.colon.findRoute(urlPath[ix:]); cNode != nil {
				return cNode, wildcard
			}
		} else if _node.colon.handler != nil {
			return _node.colon, 0
		}
	}

	if _node.wildcard != nil {
		return _node.wildcard, pathLen
	}

	return nil, 0
}

func (_node *node) optimizeRoutes() {

	if len(_node.nodes) > 0 {
		sort.Sort(_node.nodes)
		for i := 0; i < len(_node.indices); i++ {
			_node.indices[i] = 0
		}

		_node.start = _node.nodes[0].text[0]
		_node.max = _node.nodes[len(_node.nodes) - 1].text[0]

		for i := 0; i < len(_node.nodes); i++ {
			cNode := _node.nodes[i]
			cNode.parent = _node

			cByte := int(cNode.text[0] - _node.start)
			if cByte >= len(_node.indices) {
				_node.indices = append(_node.indices, make([]uint8, cByte + 1 - len(_node.indices))...)
			}
			_node.indices[cByte] = uint8(i + 1)
			cNode.optimizeRoutes()
		}
	}

	if _node.colon != nil {
		_node.colon.parent = _node
		_node.colon.optimizeRoutes()
	}

	if _node.wildcard != nil {
		_node.wildcard.parent = _node
		_node.wildcard.optimizeRoutes()
	}
}

func (_node *node) finalize() {
	if len(_node.nodes) > 0 {
		for i := 0; i < len(_node.nodes); i++ {
			_node.nodes[i].finalize()
		}
	}
	if _node.colon != nil {
		_node.colon.finalize()
	}
	if _node.wildcard != nil {
		_node.wildcard.finalize()
	}
	*_node = node{}
}

func (_node *node) string(col int) string {
	var str = "\n" + strings.Repeat(" ", col) + _node.text + " -> "
	col += len(_node.text) + 4
	for i := 0; i < len(_node.indices); i++ {
		if j := _node.indices[i]; j != 0 {
			str += _node.nodes[j - 1].string(col)
		}
	}
	if _node.colon != nil {
		str += _node.colon.string(col)
	}
	if _node.wildcard != nil {
		str += _node.wildcard.string(col)
	}
	return str
}

func (_node *node) String() string {
	if _node.text == "" {
		return _node.string(0)
	}
	col := len(_node.text) + 4
	return _node.text + " -> " + _node.string(col)
}
