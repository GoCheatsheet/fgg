//$ go run oopsla20-91/fgg -fgg -monomc=tmp/test/fg/monom/abcdef/mono-ok/param-meth-cast.go fgg/examples/monom/abcdef/mono-ok/param-meth-cast.go
//$ go run oopsla20-91/fgg -eval=2 -v tmp/test/fg/monom/abcdef/mono-ok/param-meth-cast.go

// This illustrates the need to preserve the full method sets when monomorphising
// Proposed fix: https://play.golang.org/p/yya08l04Nbg
package main;

import "fmt";

type Any(type ) interface {};

type S(type ) struct {};

type T(type ) struct {};

type Foo(type ) interface { m(type a Any())(x a, y S()) Any() }; // m<Any>(x Any, y S)

type Bar(type ) interface { m(type a Any())(x a, y T()) Any() }; // m<Any>(x Any, y T)

type V(type ) struct {};

func (x V(type )) toAny(type )(y Foo()) Any() {
	return y.(Any())
};

// V <: Foo
func (x V(type )) m(type a Any())(z a, y S()) Any() { // m<Any>(x Any, y S)
	return z.(Any())
};

func main() {
	//_ =
	fmt.Printf("%#v",
		V(){}.toAny()(V(){}).(Bar()) // cast fails
	)
}

