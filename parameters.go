package router

import "strconv"

type Values struct {
	Keys   []string
	Values []string
}

func (variables Values) Empty() bool {
	return variables.Keys == nil
}

func (variables Values) Get(name string) string {
	for i := 0; i < len(variables.Keys); i++ {
		if variables.Keys[i] == name {
			return variables.Values[i]
		}
	}
	return ""
}

func (variables Values) GetIdx(name string) int {
	for i := 0; i < len(variables.Keys); i++ {
		if variables.Keys[i] == name {
			return i
		}
	}
	return -1
}

func (variables Values) Int(name string) (int, bool) {
	var idx = variables.GetIdx(name)
	if idx == -1 {
		return 0, false
	}
	intv, err := strconv.ParseInt(variables.Values[idx], 10, strconv.IntSize)
	return int(intv), err == nil
}
