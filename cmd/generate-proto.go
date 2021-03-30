package main

import (
	"fmt"
	"os"

	pgs "github.com/vchitai/protoc-gen-star"
)

func main() {
	f, err := os.Open("./descriptors.protoset")
	if err != nil {
		panic(err)
	}
	g := pgs.Init(pgs.ProtocInput(f))
	for _, f := range g.AST().Packages() {
		if f == nil {
			continue
		}
		if f.ProtoName() == "pb" {
			for _, k := range f.Files() {
				//pgs.ExtensibleFile{File: k}.AddMethod("ProtoNameTitle", pgs.Function{
				//	Method: "post",
				//	Path:   "/extra",
				//	Name:   "Extra",
				//	Extra:  `body:"*"`,
				//})
				fmt.Println(pgs.XFile{k}.DescribeSelf())
			}
		}
	}
}
