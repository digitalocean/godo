// Package stubtest checks that every method of a mirrored type that is not
// implemented returns *unsupported.NotImplementedError.
package stubtest

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errType = reflect.TypeOf((*error)(nil)).Elem()
)

// Check calls every exported method of recv that is not in skip, with zero
// arguments, and requires a *NotImplementedError that names the method:
// returned as the last result, or panicked when the method has no error result.
// It returns how many methods it checked.
func Check(t *testing.T, platform string, recv any, skip map[string]bool) int {
	t.Helper()
	v := reflect.ValueOf(recv)
	typ := v.Type()
	checked := 0
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		if skip[m.Name] || m.Name == "Error" || m.Name == "Unwrap" {
			continue
		}
		checked++
		name := typ.String() + "." + m.Name
		func() {
			defer func() {
				if r := recover(); r != nil {
					err, _ := r.(error)
					assertNotImplemented(t, platform, name, m.Name, err, r)
				}
			}()
			fn := v.Method(i)
			ft := fn.Type()
			n := ft.NumIn()
			if ft.IsVariadic() {
				n--
			}
			args := make([]reflect.Value, n)
			for j := range args {
				if in := ft.In(j); in.Implements(ctxType) && in.Kind() == reflect.Interface {
					args[j] = reflect.ValueOf(context.Background())
				} else {
					args[j] = reflect.Zero(in)
				}
			}
			out := fn.Call(args)
			if len(out) == 0 {
				t.Errorf("%s returned without an error", name)
				return
			}
			last := out[len(out)-1]
			if last.Type() == errType {
				err, _ := last.Interface().(error)
				assertNotImplemented(t, platform, name, m.Name, err, nil)
				return
			}
			if last.Kind() == reflect.Func && last.Type().NumIn() == 1 {
				// An iterator: the stub yields the error.
				var got error
				yield := reflect.MakeFunc(last.Type().In(0), func(a []reflect.Value) []reflect.Value {
					if e, ok := a[len(a)-1].Interface().(error); ok {
						got = e
					}
					return []reflect.Value{reflect.ValueOf(false)}
				})
				last.Call([]reflect.Value{yield})
				assertNotImplemented(t, platform, name, m.Name, got, nil)
				return
			}
			t.Errorf("%s returned %v, want a not-implemented error or panic", name, out)
		}()
	}
	return checked
}

func assertNotImplemented(t *testing.T, platform, name, method string, err error, panicked any) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: no error (panic value %v)", name, panicked)
		return
	}
	var nie *unsupported.NotImplementedError
	if !errors.As(err, &nie) || !errors.Is(err, unsupported.ErrNotImplemented) {
		t.Errorf("%s: error %v is not a *NotImplementedError", name, err)
		return
	}
	if nie.Platform != platform || !strings.Contains(err.Error(), method) {
		t.Errorf("%s: message %q does not name %s.%s", name, err.Error(), platform, method)
	}
}
