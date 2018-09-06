package di

import (
	"reflect"
	"io"
	"fmt"
	"errors"
)

type node struct {
	container     *Container
	closed        bool
	valueType     reflect.Type
	providerValue reflect.Value
	cache         interface{}
}

func (n *node) ResolveValue(container *Container) (interface{}, error) {
	if n.cache != nil {
		return n.cache, nil
	}

	providerType := n.providerValue.Type()
	inTypes := extractInTypes(providerType)
	var inValues []reflect.Value
	for _, inType := range inTypes {
		component, err := container.resolveForType(container, inType)
		if err != nil {
			return nil, err
		}
		inValues = append(inValues, reflect.ValueOf(component))
	}
	outValues := n.providerValue.Call(inValues)
	outTypes, hasError := extractOutTypes(providerType)
	if hasError {
		errValue := outValues[len(outValues)-1]
		if !errValue.IsNil() {
			return nil, errValue.Interface().(error)
		}
	}

	for i, outType := range outTypes {
		currentNode := n.container.nodes[outType]
		if currentNode == nil {
			currentNode = &node{
				container:     n.container,
				closed:        false,
				valueType:     outType,
				providerValue: n.providerValue,
				cache:         outValues[i].Interface(),
			}
			n.container.nodes[outType] = currentNode
		}
		currentNode.cache = outValues[i].Interface()
	}

	return n.cache, nil
}

func (n *node) Close() error {
	if n.closed {
		return nil
	}

	var combinedErr error

	if closer, ok := n.cache.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			combinedErr = CombineErrors(combinedErr, err)
		}
	}

	n.closed = true

	return nil
}

type Container struct {
	parent *Container
	nodes  map[reflect.Type]*node
}

func New(parent *Container) *Container {
	return &Container{
		parent: parent,
		nodes:  make(map[reflect.Type]*node),
	}
}

func (container *Container) resolveForType(ctxContainer *Container, inType reflect.Type) (interface{}, error) {
	if container.parent != nil {
		value, err := container.parent.resolveForType(ctxContainer, inType)
		if err != nil {
			if _, ok := err.(missingDependencyError); !ok {
				return nil, err
			}
		} else {
			return value, nil
		}
	}
	inNode, ok := container.nodes[inType]
	if !ok {
		return nil, missingDependencyError{
			dependencyType: inType,
		}
	}
	value, err := inNode.ResolveValue(ctxContainer)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (container *Container) Provide(constructor interface{}) error {
	constructorValue := reflect.ValueOf(constructor)
	if constructorValue.Kind() != reflect.Func {
		return errors.New("constructor is not a function")
	}

	outTypes, _ := extractOutTypes(constructorValue.Type())

	for _, outType := range outTypes {
		if _, ok := container.nodes[outType]; ok {
			return fmt.Errorf("type %s already provided", outType.Name())
		}

		container.nodes[outType] = &node{
			container:     container,
			closed:        false,
			valueType:     outType,
			providerValue: constructorValue,
			cache:         nil,
		}
	}

	return nil
}

func (container *Container) Get(outPtrs ...interface{}) error {
	for _, outPtr := range outPtrs {
		outPtrValue := reflect.ValueOf(outPtr)
		if outPtrValue.Kind() != reflect.Ptr {
			return fmt.Errorf("out has to be pointer")
		}
		if outPtrValue.IsNil() {
			return fmt.Errorf("out has not to be nil")
		}
		outType := outPtrValue.Type().Elem()
		component, err := container.resolveForType(container, outType)
		if err != nil {
			return err
		}
		outPtrValue.Elem().Set(reflect.ValueOf(component))
	}

	return nil
}

func (container *Container) Close() error {
	var combinedErr error
	for _, node := range container.nodes {
		if err := node.Close(); err != nil {
			combinedErr = CombineErrors(combinedErr, err)
		}
	}
	return combinedErr
}
