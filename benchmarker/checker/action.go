package checker

type Action struct {
	Action func() (int, error)
	Name   string
	Method string
}

type Actions []*Action

func NewAction(method, name string, action func() (int, error)) *Action {
	return &Action{
		Action: action,
		Name:   name,
		Method: method,
	}
}
