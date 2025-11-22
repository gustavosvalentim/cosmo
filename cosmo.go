package cosmo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// Scope is a dependency scope
type Scope int

// ContextKeyValue is an alias type for a context key
type ContextKeyValue string

// ContextKey is the key used in the container.Context
const ContextKey ContextKeyValue = "Cosmo:Container"

const (
	ScopeTransient Scope = iota
	ScopeSingleton
)

// Container manages the configurations, providers and instances
type Container struct {
	configurations map[string]reflect.Type
	providers      map[reflect.Type]Spec
	instances      map[reflect.Type]reflect.Value
}

// Spec is a descriptor of the service providers
type Spec struct {
	Type  reflect.Type
	Value reflect.Value
	Scope Scope
}

// New creates a new Container
func New() *Container {
	return &Container{
		configurations: make(map[string]reflect.Type),
		providers:      make(map[reflect.Type]Spec),
		instances:      make(map[reflect.Type]reflect.Value),
	}
}

// Context creates a context that contains this container. Dependencies can later
// be retrived using the cosmo.Context helper function.
func (c *Container) Context() context.Context {
	return context.WithValue(context.Background(), ContextKey, c)
}

// AddWithScope will add the constructor to the providers using the specified scope.
func (c *Container) AddWithScope(scope Scope, constructor any) error {
	t, v, err := spec(constructor)
	if err != nil {
		return err
	}
	c.providers[t] = Spec{
		Type:  t,
		Value: v,
		Scope: scope,
	}
	return nil
}

// Add adds the constructor to the container with ScopeTransient
func (c *Container) Add(constructor any) error {
	if err := c.AddWithScope(ScopeTransient, constructor); err != nil {
		return err
	}
	return nil
}

// AddSingleton adds the constructor to the container with ScopeSingleton
func (c *Container) AddSingleton(constructor any) error {
	if err := c.AddWithScope(ScopeSingleton, constructor); err != nil {
		return err
	}
	return nil
}

// spec uses reflect to identify the type and value of the constructor, also performs
// validation to know if the constructor is a function and has the correct amount of
// output types.
func spec(constructor any) (reflect.Type, reflect.Value, error) {
	v := reflect.ValueOf(constructor)
	t := v.Type()

	if t.Kind() != reflect.Func {
		return t, reflect.Value{}, errors.New("constructor must be a function")
	}

	if t.NumOut() == 0 || t.NumOut() > 2 {
		return t, reflect.Value{}, errors.New("constructor must return T or (T, error)")
	}

	return t.Out(0), v, nil
}

// resolve returns the instance associated with the type passed as argument.
//
// If the dependency was registered with ScopeSingleton, then resolve will first
// check if the instance already exists, if it does, resolve won't call the ctor again.
//
// If the instance was not created before, resolve creates the instance and stores in cache
// to reuse it later.
func (c *Container) resolve(t reflect.Type) (reflect.Value, error) {
	if inst, ok := c.instances[t]; ok {
		return inst, nil
	}

	provider, ok := c.providers[t]
	if !ok {
		return reflect.Value{}, fmt.Errorf("no provider for type %v", t)
	}

	providerType := provider.Value.Type()
	args := make([]reflect.Value, providerType.NumIn())

	for i := 0; i < providerType.NumIn(); i++ {
		argType := providerType.In(i)
		val, err := c.resolve(argType)
		if err != nil {
			return reflect.Value{}, err
		}

		args[i] = val
	}

	out := provider.Value.Call(args)

	if len(out) == 2 && !out[1].IsNil() {
		return reflect.Value{}, out[1].Interface().(error)
	}

	result := out[0]

	if provider.Scope == ScopeSingleton {
		c.instances[t] = result
	}

	return result, nil
}

// Invoke runs a function, injecting the dependencies in the function arguments.
// This method uses reflection to identify the function arguments types, so it can
// know which types to resolve.
func (c *Container) Invoke(fn any) error {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return errors.New("invoke expects a function")
	}

	t := v.Type()
	args := make([]reflect.Value, t.NumIn())

	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		val, err := c.resolve(argType)
		if err != nil {
			return err
		}
		args[i] = val
	}

	v.Call(args)

	return nil
}

// Bind injects dependencies into the `out` struct.
// `out` must be a pointer to a struct.
// All dependencies inside the out struct will be resolved using the
// current cosmo.Container, and will return error if they can't.
func (c *Container) Bind(out any) error {
	v := reflect.ValueOf(out).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		val, err := c.resolve(fieldType.Type)
		if err != nil {
			return err
		}

		field.Set(val)
	}

	return nil
}

// Configure sets the constructor in a configurations map, so it can be retrieved
// later using the associated key
func (c *Container) Configure(key string, constructor any) error {
	t, _, err := spec(constructor)
	if err != nil {
		return err
	}

	if err = c.AddSingleton(constructor); err != nil {
		return err
	}

	c.configurations[key] = t

	return nil
}

// Get returns the resolved type associated with the key
func (c *Container) Get(key string) any {
	t, ok := c.configurations[key]
	if !ok {
		return nil
	}

	v, err := c.resolve(t)
	if err != nil {
		return nil
	}

	return v.Interface()
}

// Context returns the resolved type associated with key. It uses *Container.Get
// after obtaining the container inside the context. This is just a helper function, the
// container can be retrieved by using:
//
// container, ok := Context.Value(cosmo.ContextKey).(*cosmo.Container)
//
// This function only works with dependencies registered with container.Configure.
func Context(ctx context.Context, key string) any {
	container, ok := ctx.Value(ContextKey).(*Container)
	if !ok {
		return nil
	}
	return container.Get(key)
}
