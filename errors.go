package di

import (
	"reflect"
	"fmt"
	"strings"
)

type missingDependencyError struct {
	dependencyType reflect.Type
}

func (err missingDependencyError) Error() string {
	return fmt.Sprintf("missing dependency for type %s", err.dependencyType.Elem().Name())
}

type combinedError struct {
	errors []error
}

func (ce combinedError) Error() string {
	var errStrings []string
	for _, err := range ce.errors {
		errStrings = append(errStrings, err.Error())
	}
	return fmt.Sprintf("combined errors: [%s]", strings.Join(errStrings, ",\n"))
}

func CombineErrors(left error, right error) error {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	ce, ok := left.(combinedError)
	if ok {
		return combinedError{
			errors: append(ce.errors, right),
		}
	}
	return combinedError{
		errors: []error{left, right},
	}
}
