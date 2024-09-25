package internal

import (
	"errors"
	"go/ast"
	"strings"
)

type ModelField struct {
	Name string
	Type ast.Expr
}

type Model struct {
	Name   string
	Fields []*ModelField
}

type TypeConst struct {
	Type ast.Expr
	Name string
}

type EmbedModels struct {
	Models []*Model
	Types  []*TypeConst
	Consts []*Const
	Enums  map[string]*Enum
}

type ModelContainer struct {
	Imports     []string
	Models      []*Model
	EmbedModels *EmbedModels
}

func ParseModelFromStruct(node *Node) (*ModelContainer, error) {
	container := &ModelContainer{
		Imports: node.Imports,
		EmbedModels: &EmbedModels{
			Enums: make(map[string]*Enum),
		},
	}
	for _, t := range node.Types {
		suffix := strings.ToLower(t.Comment)
		if strings.HasPrefix(suffix, "model") {
			err := parseModel(t, container)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(suffix, "embed") {
			err := parseEmbedModel(t, container)
			if err != nil {
				return nil, err
			}
		}
	}

	container.EmbedModels.Enums = node.Enums
	//for enumName, enum := range node.Enums {
	//	for enumKey, enumValue := range enum.Values {
	//	}
	//}
	return container, nil
}

func parseModelFromStruct(t *Type) (*Model, error) {
	//fmt.Println(t.Type)
	if _, ok := t.Type.(*ast.StructType); ok {
		model := &Model{
			Name: t.Name,
		}
		fields := make([]*ModelField, 0)
		for _, f := range t.Fields {
			fields = append(fields, &ModelField{
				Name: f.Name,
				Type: f.Type,
			})
		}
		model.Fields = fields
		return model, nil
	}
	return nil, errors.New("not a struct " + t.Name)
}

func parseType(t *Type) (*TypeConst, error) {
	if _, ok := t.Type.(*ast.Ident); ok {
		typeConst := &TypeConst{
			Type: t.Type,
			Name: t.Name,
		}
		return typeConst, nil
	}
	return nil, errors.New("have to 원시타입")
}

func parseModel(t *Type, c *ModelContainer) error {
	model, err := parseModelFromStruct(t)
	if err != nil {
		return err
	}
	c.Models = append(c.Models, model)
	//fmt.Println(t.Name)
	return nil
}

func parseEmbedModel(t *Type, c *ModelContainer) error {
	model, err := parseModelFromStruct(t)
	if err != nil {
		typeConst, err := parseType(t)
		if err != nil {
			return err
		}
		c.EmbedModels.Types = append(c.EmbedModels.Types, typeConst)
	} else {
		//fmt.Println("aaaa", c.EmbedModels)
		c.EmbedModels.Models = append(c.EmbedModels.Models, model)
	}
	return nil
}
