package httperror

import "fmt"

type Error struct {
	Status int
	Code   int
	Value  interface{}
}

func (h Error) Error() string {
	return fmt.Sprintf("%v", h.Value)
}
