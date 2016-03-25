package Router

import "strings"

// Parameter holds the parameters matched in the route
type Parameter struct {
	*node           // matched node
	path     string // url path given
	wildcard int    // size of the wildcard match in the end of the string
}

// Index returns the index of the argument by name
func (vv *Parameter) Index(name string) int {
	if i, has := vv.names[name]; has {
		return i
	}
	return -1
}

//func (vv *Parameter) Has(name string) (has bool) {
//	_, has = vv.names[name]
//	return
//}

// Get returns the url parameter by name
func (vv *Parameter) Get(name string) string {
	if i, has := vv.names[name]; has {
		return vv.findParam(i)
	}
	return ""
}

// findParam walks up the matched node looking for parameters returns the last parameter
func (vv *Parameter) findParam(idx int) (param string) {

	curIndex := len(vv.names) - 1
	urlPath := vv.path
	pathLen := len(vv.path)
	_node := vv.node

	if _node.text == "*" {
		pathLen -= vv.wildcard
		if curIndex == idx {
			param = urlPath[pathLen:]
			return
		}
		curIndex--
		_node = _node.parent
	}

	for ; _node != nil; _node = _node.parent {
		if _node.text == ":" {
			ctn := strings.LastIndexByte(urlPath, '/')
			if ctn == -1 {
				break
			}
			pathLen = ctn + 1
			if curIndex == idx {
				param = urlPath[pathLen:]
				break
			}
			curIndex--
		} else {
			pathLen -= len(_node.text)
		}
		urlPath = urlPath[0:pathLen]
	}
	return
}
