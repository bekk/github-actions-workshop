package greeting

import (
	"errors"
	"fmt"
)

// This function takes a slice of strings and returns a greeting string.
func Greet(names []string) (string, error) {
	if len(names) == 0 {
		return "", errors.New("at least one name must be specified")
	}

	greeting := fmt.Sprintf("Hello %s", names[0])
	for i, name := range names[1:] {
		if i == len(names)-2 {
			greeting += fmt.Sprintf(" and %s", name)
		} else {
			greeting += fmt.Sprintf(", %s", name)
		}
	}
	return greeting, nil
}
