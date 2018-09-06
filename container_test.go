package di

import (
	"testing"
	"github.com/stretchr/testify/require"
	"errors"
)

type Foo struct {
	Foo string
}

func NewFoo() *Foo {
	return &Foo{
		Foo: "Foo",
	}
}

type Bar struct {
	Foo *Foo
	Bar string
}

func NewBar(foo *Foo) *Bar {
	return &Bar{
		Foo: foo,
		Bar: "Bar",
	}
}

type C struct {
	Foo    *Foo
	Closed bool
}

func NewC(foo *Foo) *C {
	return &C{
		Foo:    foo,
		Closed: false,
	}
}

func (c *C) Close() error {
	c.Closed = true
	return nil
}

type D struct {
	Foo *Foo
}

func NewD(foo *Foo) (*D, error) {
	if foo.Foo == "Foo" {
		return nil, errors.New("invalid")
	}
	return &D{
		Foo: foo,
	}, nil
}

func TestSimpleFunctions(t *testing.T) {
	assert := require.New(t)

	container := New(nil)
	err := container.Provide(NewFoo)
	assert.Nil(err, "%+v", err)
	err = container.Provide(NewBar)
	assert.Nil(err, "%+v", err)

	var bar *Bar
	err = container.Get(&bar)
	assert.Nil(err, "%+v", err)
	assert.Equal(&Bar{Foo: &Foo{Foo: "Foo"}, Bar: "Bar"}, bar)
}

func TestGetMultipleComponents(t *testing.T) {
	assert := require.New(t)

	container := New(nil)
	err := container.Provide(NewFoo)
	assert.Nil(err, "%+v", err)
	err = container.Provide(NewBar)
	assert.Nil(err, "%+v", err)

	var foo *Foo
	var bar *Bar
	err = container.Get(&foo, &bar)
	assert.Nil(err, "%+v", err)
	assert.Equal(&Foo{Foo: "Foo"}, foo)
	assert.Equal(&Bar{Foo: &Foo{Foo: "Foo"}, Bar: "Bar"}, bar)
}

func TestScoped(t *testing.T) {
	assert := require.New(t)

	parent := New(nil)
	err := parent.Provide(NewBar)
	assert.Nil(err, "%+v", err)
	child := New(parent)
	err = child.Provide(NewFoo)
	assert.Nil(err, "%+v", err)

	var foo *Foo
	var bar *Bar
	err = child.Get(&foo, &bar)
	assert.Nil(err, "%+v", err)
	assert.Equal(&Foo{Foo: "Foo"}, foo)
	assert.Equal(&Bar{Foo: &Foo{Foo: "Foo"}, Bar: "Bar"}, bar)
}

func TestCloser(t *testing.T) {
	assert := require.New(t)

	parent := New(nil)
	err := parent.Provide(NewBar)
	assert.Nil(err, "%+v", err)
	err = parent.Provide(NewC)
	assert.Nil(err, "%+v", err)
	child := New(parent)
	err = child.Provide(NewFoo)
	assert.Nil(err, "%+v", err)

	var c *C
	err = child.Get(&c)
	assert.Nil(err, "%+v", err)

	err = child.Close()
	assert.Nil(err, "%+v", err)
	assert.False(c.Closed)

	err = parent.Close()
	assert.Nil(err, "%+v", err)
	assert.True(c.Closed)
}

func TestProvideWithError(t *testing.T) {
	assert := require.New(t)

	parent := New(nil)
	err := parent.Provide(NewFoo)
	assert.Nil(err, "%+v", err)
	err = parent.Provide(NewD)
	assert.Nil(err, "%+v", err)

	var d *D
	err = parent.Get(&d)
	assert.Equal(errors.New("invalid"), err)
}
