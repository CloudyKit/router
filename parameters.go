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

import "strings"

//Parameter holds the parameters matched in the route
type Parameter struct {
	*node           // matched node
	path     string // url path given
	wildcard int    // size of the wildcard match in the end of the string
}

//IndexOf returns the index of the argument by name
func (vv *Parameter) IndexOf(name string) int {
	if i, has := vv.names[name]; has {
		return i
	}
	return -1
}

//Len returns number arguments matched in the provided URL
func (vv *Parameter) Len() int {
	return len(vv.names)
}

//ByName returns the url parameter by name
func (vv *Parameter) ByName(name string) string {
	if i, has := vv.names[name]; has {
		return vv.findParam(i)
	}
	return ""
}

//findParam walks up the matched node looking for parameters returns the last parameter
func (vv *Parameter) findParam(idx int) (param string) {

	curIndex := len(vv.names) - 1
	urlPath := vv.path
	pathLen := len(vv.path)
	_node := vv.node

	if _node.text[0] == '*' {
		pathLen -= vv.wildcard
		if curIndex == idx {
			param = urlPath[pathLen:]
			return
		}
		curIndex--
		_node = _node.parent
	}

	for _node != nil {
		if _node.text[0] == ':' {
			ctn := strings.LastIndexByte(urlPath, '/')
			if ctn == -1 {
				return
			}
			pathLen = ctn + 1
			if curIndex == idx {
				param = urlPath[pathLen:]
				return
			}
			curIndex--
		} else {
			pathLen -= len(_node.text)
		}
		urlPath = urlPath[0:pathLen]
		_node = _node.parent

	}
	return
}
