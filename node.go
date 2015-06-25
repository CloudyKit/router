package router

import (
	"net/http"
	"strings"
)

const MaxUint8 = ^uint8(0)

type node struct {
	text     string
	names    []string
	fn       func(http.ResponseWriter, *http.Request, Variables)
	wildcard *node
	colon    *node
	nodes    nodes
	indices  [MaxUint8]uint8
}

type nodes []*node

func (a nodes) Len() int           { return len(a) }
func (a nodes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a nodes) Less(i, j int) bool { return a[i].text[0] < a[j].text[0] }

func (_node *node) cmp(path string) (*node, int8, int) {

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

func (_node *node) addNode(parts []string, names []string, fn func(http.ResponseWriter, *http.Request, Variables)) {

	var (
		ccNode *node
		cNode  *node
	)

	cNode, cmpr, idx := _node.cmp(parts[0])

RESTART:
	if cNode == nil {
		cNode = &node{text: parts[0]}
		_node.nodes = append(_node.nodes, cNode)
	} else if cmpr == 1 { //
		parts[0] = parts[0][len(cNode.text):]
		ccNode, cmpr, idx = cNode.cmp(parts[0])
		if cNode != nil {
			_node = cNode
			cNode = ccNode
			goto RESTART
		}
		ccNode := &node{text: parts[0]}
		cNode.nodes = append(_node.nodes, ccNode)
		cNode = ccNode
	} else if cmpr == -1 {
		ccNode := &node{text: parts[0]}
		cNode.text = cNode.text[len(ccNode.text):]
		ccNode.nodes = nodes{cNode}
		_node.nodes[idx] = ccNode
		cNode = ccNode
	}

	if len(parts) == 1 {
		cNode.fn = fn
		cNode.names = names
		return
	}

	cNode.addNode(parts[1:], names, fn)
}

func (_node *node) getHandler(path string, pindex uint8) (func(http.ResponseWriter, *http.Request, Variables), []string, []string) {

	var llPath int

	llPath = len(path)
	if i := _node.indices[path[0]]; i != MaxUint8 {

		cNode := _node.nodes[i]

		llNode := len(cNode.text)
		if llNode > llPath {
			goto COLON
		}

		if llNode < llPath {
			if cNode.text == path[0:llNode] {
				if fn, names, values := cNode.getHandler(path[llNode:], pindex); fn != nil {
					return fn, names, values
				}
			}
			goto COLON
		}
		if cNode.text == path {
			return cNode.fn, cNode.names, nil
		}
	}

COLON:
	if _node.colon != nil {
		for ix := 0; ix < llPath; ix++ {
			if path[ix] == '/' {
				if fn, names, values := _node.colon.getHandler(path[ix:], pindex+1); fn != nil {
					if values == nil {
						values = make([]string, pindex+1, pindex+1)
					}
					values[pindex] = path[0:ix]
					return fn, names, values
				}
				goto WILDCARD
			}
		}

		if llPath > 0 {
			values := make([]string, pindex+1, pindex+1)
			values[pindex] = path
			return _node.colon.fn, _node.colon.names, values
		}
	}

WILDCARD:
	if _node.wildcard != nil {
		values := make([]string, pindex+1, pindex+1)
		values[pindex] = path
		return _node.wildcard.fn, _node.wildcard.names, values
	}

	return nil, nil, nil
}

func (_node *node) sort() {

	for i := uint8(0); i < MaxUint8; i++ {
		_node.indices[i] = MaxUint8
	}

	for i := 0; i < len(_node.nodes); i++ {
		_node.indices[_node.nodes[i].text[0]] = uint8(i)
		_node.nodes[i].sort()
	}
	if _node.colon != nil {
		_node.colon.sort()
	}
	if _node.wildcard != nil {
		_node.wildcard.sort()
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
