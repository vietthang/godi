package di

import "reflect"

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func extractInTypes(funcType reflect.Type) []reflect.Type {
	inTypes := make([]reflect.Type, funcType.NumIn())

	for i := 0; i < funcType.NumIn(); i++ {
		inTypes[i] = funcType.In(i)
	}

	return inTypes
}

func extractOutTypes(funcType reflect.Type) ([]reflect.Type, bool) {
	numOut := funcType.NumOut()

	hasError := false

	if numOut > 0 {
		lastOutType := funcType.Out(numOut - 1)

		if lastOutType == errorType {
			hasError = true
			numOut = numOut - 1
		}
	}

	outTypes := make([]reflect.Type, numOut)
	for i := 0; i < numOut; i++ {
		outTypes[i] = funcType.Out(i)
	}

	return outTypes, hasError
}
