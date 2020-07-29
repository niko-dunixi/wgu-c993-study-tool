// +build ignore

package main

import (
	"github.com/shurcooL/vfsgen"
	"net/http"
)

func main() {
	err := vfsgen.Generate(http.Dir("raw-templates"), vfsgen.Options{
		Filename:     "myTemplates/assets.go",
		PackageName:  "myTemplates",
		VariableName: "assets",
	})
	if err != nil {
		panic(err)
	}
}
