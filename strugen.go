package strugen

import (
	"bytes"
	"fmt"
	"go/ast"

	"github.com/fatih/structtag"
	"go/printer"
	"go/token"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/tools/go/packages"
)

type Generator struct {
	Types   []string
	TagName string
}

type Struct struct {
	Name   string
	Fields map[string]StructField
}

type StructField struct {
	Name     string
	Exported string
	Tagged   bool
	TagValue string
	Type     string
}

func (g *Generator) FindStructs() (map[string]Struct, string, error) {
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return map[string]Struct{}, "", fmt.Errorf("error loading package: %v", err)
	}

	if len(pkgs) != 1 {
		return map[string]Struct{}, "", fmt.Errorf("error: %d packages found, expected 1", len(pkgs))
	}

	structs := map[string]Struct{}

	for _, f := range pkgs[0].Syntax {
		smap, err := g.parseStruct(f, pkgs[0].Fset)
		if err != nil {
			return map[string]Struct{}, "", err
		}

		structs = lo.Assign(structs, smap)
	}

	return structs, pkgs[0].Name, nil
}

func (g *Generator) parseStruct(file *ast.File, fileSet *token.FileSet) (map[string]Struct, error) {
	sMap := map[string]Struct{}

	var inspectError error

	ast.Inspect(file, func(n ast.Node) bool {
		// If we've errored on a previous node, then just stop trying
		if inspectError != nil {
			return false
		}

		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || typeSpec.Type == nil {
			return true
		}

		s, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structName := typeSpec.Name.Name

		if !lo.Contains(g.Types, structName) {
			return false
		}

		struct_ := Struct{
			Name: structName,
			Fields: map[string]StructField{},
		}
		for _, field := range s.Fields.List {
			sf := StructField{
				Name: field.Names[0].Name,
			}
			typeNameBuf := new(bytes.Buffer)
			err := printer.Fprint(typeNameBuf, fileSet, field.Type)
			if err != nil {
				inspectError = fmt.Errorf("error printing type name: %v", err)
				return false
			}

			sf.Type = typeNameBuf.String()

			if field.Tag != nil {
				tag := field.Tag.Value
				tag = strings.Trim(tag, "`")
				tags, err := structtag.Parse(tag)
				if err != nil {
					inspectError = fmt.Errorf("error extracting tags: %v", err)
					return false
				}

				userTag, err := tags.Get(g.TagName)
				if err == nil {
					sf.Tagged = true
					sf.TagValue = strings.Join(append([]string{userTag.Name}, userTag.Options...), ",")
				} else {
					sf.Tagged = false
				}
			} else {
				sf.Tagged = false
			}

			struct_.Fields[sf.Name] = sf
		}

		sMap[struct_.Name] = struct_
		return false
	})

	return sMap, inspectError
}
