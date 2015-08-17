package router

import (
	"strings"
)

const MaxUint8 = ^uint8(0)

//TODO: Idea, add more types of node ex: /services?/?index.html will match /services,/services/,/services/index.html

type node struct {
	text     string
	names    []string
	handler  Handler
	wildcard *node
	colon    *node
	nodes    nodes
	indices  [MaxUint8]uint8
}

type nodes []*node

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

func (_node *node) addRoute(parts []string, names []string, handler Handler) {

	var (
		ccNode *node
		cNode  *node
	)

	cNode, result, idx := _node.nextRoute(parts[0])

RESTART:
	if cNode == nil {
		cNode = &node{text: parts[0]}
		_node.nodes = append(_node.nodes, cNode)
	} else if result == 1 { //
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

func (_node *node) findRoute(urlPath string, paramIndex uint8) (Handler, []string, []string) {

	pathLen := len(urlPath)
	if i := _node.indices[urlPath[0]]; i != MaxUint8 {

		cNode := _node.nodes[i]

		nodeLen := len(cNode.text)
		if nodeLen > pathLen {
			goto COLON
		}

		if nodeLen < pathLen {
			if cNode.text == urlPath[0:nodeLen] {
				if handle, names, values := cNode.findRoute(urlPath[nodeLen:], paramIndex); handle != nil {
					return handle, names, values
				}
			}
			goto COLON
		}

		if cNode.text == urlPath {
			if cNode.handler == nil && cNode.wildcard != nil {
				values := make([]string, paramIndex+1, paramIndex+1)
				values[paramIndex] = ""
				return cNode.wildcard.handler, cNode.wildcard.names, values
			}
			return cNode.handler, cNode.names, nil
		}
	}

COLON:
	if _node.colon != nil {
		for ix := 0; ix < pathLen; ix++ {
			if urlPath[ix] == '/' {
				if fn, names, values := _node.colon.findRoute(urlPath[ix:], paramIndex+1); fn != nil {
					if values == nil {
						values = make([]string, paramIndex+1, paramIndex+1)
					}
					values[paramIndex] = urlPath[0:ix]
					return fn, names, values
				}
				goto WILDCARD
			}
		}

		if pathLen > 0 {
			values := make([]string, paramIndex+1, paramIndex+1)
			values[paramIndex] = urlPath
			return _node.colon.handler, _node.colon.names, values
		}
	}

WILDCARD:
	if _node.wildcard != nil {
		values := make([]string, paramIndex+1, paramIndex+1)
		values[paramIndex] = urlPath
		return _node.wildcard.handler, _node.wildcard.names, values
	}

	return nil, nil, nil
}

func (_node *node) optimizeRoutes() {

	for i := uint8(0); i < MaxUint8; i++ {
		_node.indices[i] = MaxUint8
	}

	for i := 0; i < len(_node.nodes); i++ {
		_node.indices[_node.nodes[i].text[0]] = uint8(i)
		_node.nodes[i].optimizeRoutes()
	}
	if _node.colon != nil {
		_node.colon.optimizeRoutes()
	}
	if _node.wildcard != nil {
		_node.wildcard.optimizeRoutes()
	}
}

func (_node *node) string(col int) string {
	var str = "\n" + strings.Repeat(" ", col) + _node.text + " -> "
	col += len(_node.text) + 4
	for i := 0; i < len(_node.nodes); i++ {
		str += _node.nodes[i].string(col)
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
