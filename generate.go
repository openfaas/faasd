// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(
		http.Dir("./static"),
		vfsgen.Options{
			Filename:     "./pkg/faasd_static.go",
			PackageName:  "pkg",
			VariableName: "Assets",
		})
	if err != nil {
		log.Fatalln(err)
	}
}
