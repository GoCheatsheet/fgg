// Pre (1): ANTLR4
// E.g., antlr-4.7.1-complete.jar
// (See go:generate below)

// Pre (2): ANTLR4 Runtime for Go
//$ go get github.com/antlr/antlr4/runtime/Go/antlr
// Optional:
//$ cd $CYGHOME/code/go/src/github.com/antlr/antlr4
//$ git checkout -b antlr-go-runtime tags/4.7.1  // Match antlr-4.7.1-complete.jar -- but unnecessary

//rhu@HZHL4 MINGW64 ~/code/go/src/
//$ go run github.com/rhu1/fgg -eval=10 fg/examples/hello/hello.go
//$ go run github.com/rhu1/fgg -inline="package main; type A struct {}; func main() { _ = A{} }"
// or
//$ go install
//$ /c/Users/rhu/code/go/bin/fgg.exe ...

// N.B. GoInstall installs to $CYGHOME/code/go/bin (not $WINHOME)

// Assuming "antlr4" alias for (e.g.): java -jar ~/code/java/lib/antlr-4.7.1-complete.jar
//go:generate antlr4 -Dlanguage=Go -o parser FG.g4

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"

	"github.com/rhu1/fgg/fg"
)

var _ = reflect.TypeOf
var _ = strconv.Itoa

var verbose bool = false

// N.B. flags (e.g., -internal=true) must be supplied before any non-flag args
func main() {
	evalPtr := flag.Int("eval", -1, "Number of steps to evaluate")
	internalPtr := flag.Bool("internal", false, "Use \"internal\" input as source")
	inlinePtr := flag.String("inline", "", "Use inline input as source")
	strictParsePtr := flag.Bool("strict", true,
		"Set strict parsing (panic on error, no recovery)")
	verbosePtr := flag.Bool("v", false,
		"Set verbose printing")
	flag.Parse()

	verbose = *verbosePtr

	var src string
	if *internalPtr { // First priority
		src = makeInternalSrc()
	} else if *inlinePtr != "" { // Second priority, i.e., -inline overrules src file arg
		src = *inlinePtr
	} else {
		if len(os.Args) < 2 {
			fmt.Println("Input error: need a source .go file (or an -inline program)")
		}
		bs, err := ioutil.ReadFile(os.Args[len(os.Args)-1])
		checkErr(err)
		src = string(bs)
	}

	vPrintln("\nParsing AST:")
	var adptr fg.FGAdaptor
	prog := adptr.Parse(*strictParsePtr, src) // AST (FGProgram root)
	vPrintln(prog.String())

	vPrintln("\nChecking source program OK:")
	allowStupid := false
	prog.Ok(allowStupid)

	if *evalPtr < 0 {
		return
	}
	vPrintln("\nEntering Eval loop:")
	vPrintln("Decls:")
	for _, v := range prog.GetDecls() {
		vPrintln("\t" + v.String() + ";")
	}
	vPrintln("Eval steps:")
	allowStupid = true
	vPrintln(fmt.Sprintf("%6d: %v", 0, prog.GetExpr()))
	for i := 1; i <= *evalPtr; i++ {
		prog = prog.Eval()
		vPrintln(fmt.Sprintf("%6d: %v", i, prog.GetExpr()))
		vPrintln("Checking OK:")
		prog.Ok(allowStupid)
	}
}

// For convenient quick testing -- via flag "-internal=true"
func makeInternalSrc() string {
	Any := "type Any interface {}"
	ToAny := "type ToAny struct { any Any }"
	e := "ToAny{1}"
	return fg.MakeFgProgram(Any, ToAny, e)
}

/* Helpers */

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func vPrintln(x string) {
	if verbose {
		fmt.Println(x)
	}
}

/* TODO
- WF: repeat type decl

	//b.WriteString("type B struct { f t };\n")  // TODO: unknown type
	//b.WriteString("type B struct { b B };\n")  // TODO: recursive struct
*/
