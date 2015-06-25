package router

type Variables struct {
	Keys   []string
	Values []string
}

func (variables Variables) Get(name string) string {
	for i := 0; i < len(variables.Keys); i++ {
		if variables.Keys[i] == name {
			return variables.Values[i]
		}
	}
	return ""
}
