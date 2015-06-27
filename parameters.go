package router

type Variables struct {
	key    []string
	values []string
}

func (variables Variables) Empty() bool {
	return variables.key == nil
}

func (variables Variables) Get(name string) string {
	for i := 0; i < len(variables.key); i++ {
		if variables.key[i] == name {
			return variables.values[i]
		}
	}
	return ""
}
