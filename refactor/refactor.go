package refactor

import (
	"bytes"
	"fmt"
	"go/printer"
	"go/token"
	"io/ioutil"
	"strings"

	"github.com/bblfsh/go-driver/driver/golang"
	"github.com/bblfsh/go-driver/driver/golang/convert"
	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/transformer"
	"github.com/bblfsh/sdk/v3/uast/uastyaml"
)

const mainTemplate = `package main

func main() {
	%s
}`

type Refactor struct {
	before string
	after  string
	m      transformer.Transformer
}

func NewRefactor(beforePath, afterPath string) (*Refactor, error) {
	beforeData, err := ioutil.ReadFile(beforePath)
	if err != nil {
		return nil, err
	}

	afterData, err := ioutil.ReadFile(afterPath)
	if err != nil {
		return nil, err
	}

	return &Refactor{
		before: string(beforeData),
		after:  string(afterData),
	}, nil
}

func (r *Refactor) Prepare() error {
	in, err := parseNodeHack(r.before)
	if err != nil {
		return err
	}

	out, err := parseNodeHack(r.after)
	if err != nil {
		return err
	}

	dump(in, "../out/1.yml")
	dump(out, "../out/2.yml")

	r.m = transformer.Mappings(
		transformer.Map(nodeToOp(in), nodeToOp(out)))
	return nil
}

func (r *Refactor) Apply(code string) (string, error) {
	test, err := golang.Parse(code)
	if err != nil {
		return "", err
	}

	test, err = trimPositions(test)
	if err != nil {
		return "", err
	}

	dump(test, "../out/test.yml")

	res, err := r.m.Do(test)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := printer.Fprint(buf, token.NewFileSet(), convert.NodeToAST(res)); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func trimPositions(n nodes.Node) (nodes.Node, error) {
	return transformer.Mappings(transformer.Map(
		transformer.Part("_", transformer.Obj{uast.KeyPos: transformer.Any()}),
		transformer.Part("_", transformer.Obj{}),
	)).Do(n)
}

func nodeToOp(n nodes.Node) transformer.Op {
	switch o := n.(type) {
	case nil:
		return transformer.Is(o)
	case nodes.Value:
		return transformer.Is(o)
	case nodes.Object:
		if uast.TypeOf(o) == "Ident" {
			name := o["Name"]
			str := name.(nodes.String)
			if strings.HasPrefix(string(str), "X") {
				return transformer.Var(string(str))
			}
		}
		// transformer.Fields is extended version of transformer.Obj
		var res transformer.Fields
		for k, v := range o {
			// conditions to drop pos
			var field = transformer.Field{Name: k}
			if k == uast.KeyPos {
				field.Drop = true
				field.Op = transformer.Any()
			} else {
				field.Op = nodeToOp(v)
			}
			res = append(res, field)
		}
		return res
	case nodes.Array:
		var res []transformer.Op
		for _, node := range o {
			res = append(res, nodeToOp(node))
		}
		return transformer.Arr(res...)
	default:
		panic("not supported type " + o.Kind().String())
	}
}

// TODO functions
func parseNodeHack(snippet string) (nodes.Node, error) {
	wrapped, err := golang.Parse(wrapInMain(snippet))
	if err != nil {
		return nil, err
	}

	return trimPositions(wrapped.(nodes.Object)["Decls"].(nodes.Array)[0].(nodes.Object)["Body"].(nodes.Object)["List"].(nodes.Array)[0])
}

// TODO gofmt
func wrapInMain(code string) string {
	return fmt.Sprintf(mainTemplate, code)
}

func dump(n nodes.Node, filePath string) {
	data, err := uastyaml.Marshal(n)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		panic(err)
	}
}
