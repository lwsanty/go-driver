package main

import (
	"testing"

	"github.com/bblfsh/go-driver/driver/golang"
	"github.com/bblfsh/sdk/v3/uast/transformer"
	"github.com/stretchr/testify/require"
)

func TestA(t *testing.T) {
	in, err := golang.Parse(getCode("Johnny"))
	require.NoError(t, err)

	exp, err := golang.Parse(getCode("Bravoo"))
	require.NoError(t, err)

	m := transformer.Mappings(
		transformer.Map(
			transformer.String("\"Johnny\""),
			transformer.String("\"Bravoo\""),
		),
	)

	act, err := m.Do(in)
	require.NoError(t, err)

	require.Equal(t, exp, act)
}

func getCode(name string) string {
	return `package main

import "fmt"

func main() {
	var a string = "` + name + `"
	
	fmt.Println(a)
}
`
}
