package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Node struct {
	Package string
	Imports []string
	Consts  []*ConstWrap
	Types   []*Type
	Enums   map[string]*Enum
}

type Field struct {
	Name    string
	Type    ast.Expr
	Tag     string
	Comment string
}

type Type struct {
	Type    ast.Expr
	Name    string
	Fields  []*Field
	Comment string
}

type ConstWrap struct {
	Type    ast.Expr
	Consts  []*Const
	Comment string
}

type Const struct {
	Type  ast.Expr
	Name  string
	Value string
}

type Enum struct {
	Values map[string]int
}

func ParseStruct(fileName string) (*Node, error) {
	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, fileName, nil, parser.ParseComments)

	if err != nil {
		return nil, err
	}

	nodeInfo := &Node{}
	nodeInfo.Package = node.Name.Name

	ast.Inspect(node, func(n ast.Node) bool {
		if gd, ok := n.(*ast.GenDecl); ok {
			for _, sp := range gd.Specs {
				switch s := sp.(type) {
				case *ast.TypeSpec:
					anyType := &Type{
						Type:    s.Type,
						Name:    s.Name.Name,
						Comment: gd.Doc.Text(),
					}
					if structType, ok := s.Type.(*ast.StructType); ok {
						var fields []*Field

						for _, f := range structType.Fields.List {

							for _, id := range f.Names {
								fields = append(fields, &Field{
									Name:    id.Name,
									Type:    f.Type,
									Comment: f.Comment.Text(),
								})

							}
						}

						anyType.Fields = fields
					}
					nodeInfo.Types = append(nodeInfo.Types, anyType)
				case *ast.ImportSpec:
					nodeInfo.Imports = append(nodeInfo.Imports, strings.ReplaceAll(s.Path.Value, "\"", ""))
				case *ast.ValueSpec:
					if gd.Tok == token.CONST {
						constsWrap := &ConstWrap{}
						constsWrap.Comment = gd.Doc.Text()
						consts := make([]*Const, 0)
						constsWrap.Type = s.Type

						for i, id := range s.Names {
							var value string
							if len(s.Values) > i {
								if e, ok := s.Type.(*ast.Ident); ok {
									if bl, ok := s.Values[i].(*ast.Ident); ok {
										if bl.Name == "iota" && strings.HasPrefix(id.Name, e.Name) {
											//fmt.Println(bl.Name)
											if nodeInfo.Enums == nil {
												nodeInfo.Enums = make(map[string]*Enum)
											}
											nodeInfo.Enums[e.Name] = &Enum{
												Values: map[string]int{
													id.Name: 0,
												},
											}
										}
									}
								}
								value = fmt.Sprint(id.Obj.Decl.(*ast.ValueSpec).Values[0])
							} else {
								for k, v := range nodeInfo.Enums {
									res := strings.Split(id.Name, "_")
									if res[0] == k {
										var enumMax int
										for _, _v := range v.Values {
											enumMax = max(_v, enumMax)
										}
										nodeInfo.Enums[k].Values[id.Name] = enumMax + 1
										break
									}
								}
							}
							consts = append(consts, &Const{
								Type:  s.Type,
								Name:  id.Name,
								Value: value,
							})
						}
						constsWrap.Consts = consts
						nodeInfo.Consts = append(nodeInfo.Consts, constsWrap)
					}

				}

			}

		}
		return true
	})
	//fmt.Println(nodeInfo.Enums)

	return nodeInfo, nil
}
