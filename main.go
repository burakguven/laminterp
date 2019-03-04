package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

var formatFlag = flag.Bool("format", false, "print a formatted version of the program instead of evaluating it")

func main() {
	log.SetFlags(0)
	log.SetPrefix("laminterp: ")

	flag.Parse()

	if flag.NArg() > 1 {
		log.Fatal("too many arguments")
	}

	if flag.NArg() == 1 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		scriptMode(f)
	} else if readline.DefaultIsTerminal() {
		interactiveMode()
	} else {
		scriptMode(os.Stdin)
	}
}

func interactiveMode() {
	rl, err := readline.New("")
	if err != nil {
		log.Fatal(err)
	}

	for {
	ReadNew:
		program := ""
	ReadMore:
		if program == "" {
			rl.SetPrompt(">> ")
		} else {
			rl.SetPrompt(".. ")
		}
		line, err := rl.Readline()
		if err != nil && (err == io.EOF || err == readline.ErrInterrupt) {
			// If the user interrupts with no prior input, they're
			// probably trying to quit the interpreter. Otherwise,
			// they're probably just trying to start another
			// program.
			if program == "" && line == "" {
				break
			} else {
				goto ReadNew
			}
		} else if err != nil {
			log.Fatal(err)
		}
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			program += trimmed + "\n"
		}

		// Since this language has a simple grammar, we can just
		// attempt to parse the program entered so far to see if
		// it's valid and then assume there's more if we get an
		// unexpected EOF error.
		node := parseString(program)
		if node.typ == nodeError && isUnexpectedEOFError(node) {
			goto ReadMore
		} else if node.typ == nodeError {
			fmt.Println("parse error:", node.val)
		} else if *formatFlag {
			format(node, "")
			fmt.Println()
		} else {
			fmt.Println(eval(node))
		}
	}
}

func scriptMode(r io.Reader) {
	program, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	node := parseString(string(program))
	if node.typ == nodeError {
		log.Fatalln("parse error:", node.val)
		return
	}
	if *formatFlag {
		format(node, "")
		fmt.Println()
	} else {
		obj := eval(node)
		if obj.typ == objectError {
			log.Fatalln("runtime error:", obj)
		}
		fmt.Println(obj)
	}
}

func isSimpleNode(n *node) bool {
	switch n.typ {
	case nodeIdentifier, nodeNumber, nodeBool:
		return true
	default:
		return false
	}
}

const formatIndent = "    "

func format(n *node, indent string) {
	switch {
	case isSimpleNode(n):
		fmt.Printf("%s%v", indent, n.val)
	case n.typ == nodeLam:
		lam := n.val.(*lamNode)
		fmt.Printf("%slam %v ", indent, lam.param)
		if isSimpleNode(lam.body) {
			fmt.Print(lam.body.val)
		} else {
			fmt.Println()
			format(lam.body, indent+formatIndent)
		}
	case n.typ == nodeApp:
		app := n.val.(*appNode)
		fmt.Printf("%sapp", indent)
		if isSimpleNode(app.fn) && isSimpleNode(app.arg) {
			fmt.Printf(" %v %v", app.fn.val, app.arg.val)
		} else {
			fmt.Println()
			format(app.fn, indent+formatIndent)
			fmt.Println()
			format(app.arg, indent+formatIndent)
		}
	}
}
