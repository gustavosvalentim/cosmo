package cosmo

import (
	"errors"
	"fmt"
	"reflect"
)

type Container struct {
	providers map[reflect.Type]reflect.Value
	instances map[reflect.Type]reflect.Value
}

func New() *Container {
	return &Container{
		providers: make(map[reflect.Type]reflect.Value),
		instances: make(map[reflect.Type]reflect.Value),
	}
}

func (c *Container) Add(constructor any) error {
	v := reflect.ValueOf(constructor)
	t := v.Type()

	if t.Kind() != reflect.Func {
		return errors.New("constructor must be a function")
	}

	if t.NumOut() == 0 || t.NumOut() > 2 {
		return errors.New("constructor must return T or (T, error)")
	}

	resultType := t.Out(0)
	c.providers[resultType] = v

	return nil
}

func (c *Container) Resolve(t reflect.Type) (reflect.Value, error) {
	if inst, ok := c.instances[t]; ok {
		return inst, nil
	}

	provider, ok := c.providers[t]
	if !ok {
		return reflect.Value{}, fmt.Errorf("no provider for type %v", t)
	}

	providerType := provider.Type()
	args := make([]reflect.Value, providerType.NumIn())

	for i := 0; i < providerType.NumIn(); i++ {
		argType := providerType.In(i)
		val, err := c.Resolve(argType)
		if err != nil {
			return reflect.Value{}, err
		}
		args[i] = val
	}

	out := provider.Call(args)

	if len(out) == 2 && !out[1].IsNil() {
		return reflect.Value{}, out[1].Interface().(error)
	}

	result := out[0]

	c.instances[t] = result

	return result, nil
}

func (c *Container) Invoke(fn any) error {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return errors.New("invoke expects a function")
	}

	t := v.Type()
	args := make([]reflect.Value, t.NumIn())

	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		val, err := c.Resolve(argType)
		if err != nil {
			return err
		}
		args[i] = val
	}

	v.Call(args)

	return nil
}

func (c *Container) Bind(out any) error {
	v := reflect.ValueOf(out).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		val, err := c.Resolve(fieldType.Type)
		if err != nil {
			return err
		}

		field.Set(val)
	}

	return nil
}
