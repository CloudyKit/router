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
	"net/http"
	"strings"
	"testing"
)

func assert_equals_string(one, two []string) bool {
	if len(one) != len(two) {
		return false
	}
	for i := 0; i < len(one); i++ {
		if one[i] != two[i] {
			return false
		}
	}
	return true
}

func TestSplitURLPath(t *testing.T) {

	var table = map[string][2][]string{
		"/*name":                               {{"/", "*"}, {"name"}},
		"/users/:name":                         {{"/users/", ":"}, {"name"}},
		"/users/:name/put":                     {{"/users/", ":", "/put"}, {"name"}},
		"/users/:name/put/:section":            {{"/users/", ":", "/put/", ":"}, {"name", "section"}},
		"/customers/:name/put/:section":        {{"/customers/", ":", "/put/", ":"}, {"name", "section"}},
		"/customers/groups/:name/put/:section": {{"/customers/groups/", ":", "/put/", ":"}, {"name", "section"}},
	}

	for path, result := range table {
		parts, names := splitURLpath(path)
		if !assert_equals_string(parts, result[0]) {
			t.Errorf("Expected %v %v: %v %v", result[0], result[1], parts, names)
		}
	}
}

var testTable = [][]string{
	{"/public/*fpath", "/public/index.html", "/public/favicon.png", "/public/images/bg.gif"},
	{"/index", "/index"},
	{"/users/:userId", "/users/666f24b7-cf7f-4176-bf07-c6d937e622c9"},
	{"/users/:userId/companies/:companyId", "/users/666f24b7-cf7f-4176-bf07-c6d937e622c9/companies/666f24b7-cf7f-4176-bf07-c6d937e622c9"},
	{"/us", "/us"},
	{"/*fpath", "/", "/site_1/index.html", "/site_1/favicon.png", "/site_1/images/bg.gif"},
}

func TestTreeLookupSimple(t *testing.T) {
	router := New()

	for _, v := range testTable {
		v := v
		router.AddRoute("GET", v[0], func(w http.ResponseWriter, r *http.Request, vp Parameter) {
			for i := 1; i < len(v); i++ {
				if r.URL.Path == v[i] {
					return
				}
			}
			t.Errorf("GOT %s EXPECTED %s\n", r.URL.Path, v)
		})
	}

	t.Log(router.String())

	for _, v := range testTable {
		for i := 1; i < len(v); i++ {
			t.Log("GET " + v[i])
			fn, variables := router.FindRoute("GET", v[i])
			if fn == nil {
				t.Error("Not Found", v[i], variables)
				continue
			}
			req, _ := http.NewRequest("GET", v[i], nil)
			fn(nil, req, variables)
		}
	}

}

func TestTreeIndicesBug(t *testing.T) {
	router := New()
	testTable := [][]string{
		{"/", "/"},
		{"/books", "/books"},
		{"/source", "/source"},
	}

	for _, v := range testTable {
		v := v
		router.AddRoute("GET", v[0], func(w http.ResponseWriter, r *http.Request, vp Parameter) {
			for i := 1; i < len(v); i++ {
				if r.URL.Path == v[i] {
					return
				}
			}
			t.Errorf("GOT %s EXPECTED %s\n", r.URL.Path, v)
		})
	}

	for _, v := range testTable {
		for i := 1; i < len(v); i++ {
			t.Log("GET " + v[i])
			fn, variables := router.FindRoute("GET", v[i])
			if fn == nil {
				t.Error("Not Found", v[i], variables)
				continue
			}
			req, _ := http.NewRequest("GET", v[i], nil)
			fn(nil, req, variables)
		}
	}
}

var router = New()

var benchRouter = New()
var benchTest = [][]string{
	{"/user/:name3/:userId/*path2", "name3", "userId", "path2"},
	{"/user/:name2/list", "name2"},
	{"/:name", "name"},
	{"/user/:name/*path", "name", "path"},
	{"/user/files/*path3", "path3"},
}

func init() {

	for i := 0; i < len(benchTest); i++ {
		benchRouter.AddRoute("GET", benchTest[i][0], func(w http.ResponseWriter, r *http.Request, vp Parameter) {

		})
		benchTest[i][0] = strings.NewReplacer(":", "", "*", "").Replace(benchTest[i][0])
	}

	for _, v := range testTable {
		v := v
		router.AddRoute("GET", v[0], func(w http.ResponseWriter, r *http.Request, vp Parameter) {
			for i := 1; i < len(v); i++ {
				if r.URL.Path == v[i] {
					return
				}
			}
		})
	}
}

func TestGetParam(t *testing.T) {
	for i := 0; i < len(benchTest); i++ {
		fn, vl := benchRouter.FindRoute("GET", benchTest[i][0])
		if fn == nil {
			t.Errorf("%q was not found \n %s", benchTest[i][0], benchRouter)
		}
		for j := 1; j < len(benchTest[i]); j++ {
			param := vl.Get(benchTest[i][j])
			if param != benchTest[i][j] {
				t.Errorf("%s Expected param %q get %q", benchTest[i][0], benchTest[i][j], param)
			}
		}
	}
}

func BenchmarkGetParam1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchtest := benchTest[0]
		_, vl := benchRouter.FindRoute("GET", benchtest[0])
		param := vl.Get(benchtest[1])
		if param != benchtest[1] {
			b.Errorf("Expected param %q get %q", benchtest[1], param)
		}
	}
}

func BenchmarkGetParams(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for i := 0; i < len(benchTest); i++ {
			fn, vl := benchRouter.FindRoute("GET", benchTest[i][0])
			if fn == nil {
				b.Errorf("%q was not found", benchTest[i][0])
			}
			for j := 1; j < len(benchTest[i]); j++ {
				param := vl.Get(benchTest[i][j])
				if param != benchTest[i][j] {
					b.Errorf("Expected param %q get %q", benchTest[i][j], param)
				}
			}
		}
	}
}

func BenchmarkWildCard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn, _ := router.FindRoute("GET", testTable[0][1])
		if fn == nil {
			b.Error("Not Found", testTable[0][1])
			continue
		}
	}
}

func BenchmarkPart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn, _ := router.FindRoute("GET", testTable[3][1])
		if fn == nil {
			b.Error("Not Found", testTable[3][1])
			continue
		}
	}
}

func BenchmarkManyURLS(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range testTable {
			for i := 1; i < len(v); i++ {
				fn, _ := router.FindRoute("GET", v[i])
				if fn == nil {
					b.Error("Not Found", v[i])
					continue
				}
			}
		}
	}
}
