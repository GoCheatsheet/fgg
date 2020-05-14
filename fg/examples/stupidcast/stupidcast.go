//$ go run oopsla20-91/fgg -eval=-1 -v fg/examples/stupidcast/stupidcast.go

package main; type Any interface{}; type ToAny struct { any Any }; type A struct {}; func main() { _ = ToAny{A{}}.any.(ToAny) }
