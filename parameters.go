package router

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
