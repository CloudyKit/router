package router

import (
	"net/http"
	"testing"
)

func stringsEq(one, two []string) bool {

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

var table = map[string][2][]string{
	"/*name":                    {{"/", "*"}, {"name"}},
	"/users/:name":              {{"/users/", ":"}, {"name"}},
	"/users/:name/put":          {{"/users/", ":", "/put"}, {"name"}},
	"/users/:name/put/:section": {{"/users/", ":", "/put/", ":"}, {"name", "section"}},

	"/customers/:name/put/:section":        {{"/customers/", ":", "/put/", ":"}, {"name", "section"}},
	"/customers/groups/:name/put/:section": {{"/customers/groups/", ":", "/put/", ":"}, {"name", "section"}},
}

func TestSplit(t *testing.T) {

	for path, result := range table {
		parts, names := split(path)
		if !stringsEq(parts, result[0]) || !stringsEq(names, result[1]) {
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
}

func TestTreeLookupSimple(t *testing.T) {
	router := New()
	for _, v := range testTable {
		v := v
		router.Handle("GET", v[0], func(w http.ResponseWriter, r *http.Request, vp Variables) {
			for i := 1; i < len(v); i++ {
				if r.URL.Path == v[i] {
					return
				}
			}
			t.Errorf("GOT %s EXPECTED %s\n", r.URL.Path, v[1:])
		})
	}

	t.Log(router.String())

	for _, v := range testTable {
		for i := 1; i < len(v); i++ {
			t.Log("GET " + v[i])
			fn, variables := router.Lookup("GET", v[i])
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

func TestSetup(t *testing.T) {
	for _, v := range testTable {
		v := v
		router.Handle("GET", v[0], func(w http.ResponseWriter, r *http.Request, vp Variables) {
			for i := 1; i < len(v); i++ {
				if r.URL.Path == v[i] {
					return
				}
			}
		})
	}
}

func BenchmarkWildCard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn, _ := router.Lookup("GET", testTable[0][1])
		if fn == nil {
			b.Error("Not Found", testTable[0][1])
			continue
		}
	}
}

func BenchmarkPart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fn, _ := router.Lookup("GET", testTable[3][1])
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
				fn, _ := router.Lookup("GET", v[i])
				if fn == nil {
					b.Error("Not Found", v[i])
					continue
				}
			}
		}
	}
}
