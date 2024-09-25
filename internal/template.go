package internal

import (
	_ "embed"
	"errors"
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"strings"
	"unicode"
)

import (
	"html/template"
)

type TModel struct {
	Name   string
	Fields []*TModelField
}

type TModelField struct {
	Type  string
	Upper string
	Snake string
}

type TConst struct {
	Type     string
	Name     string
	Value    string
	HasValue bool
}

type TEnum struct {
	Type   string
	Values map[string]int
}

type TType struct {
	Name string
	Type string
}

type EmbedModelTemplate struct {
	PackageName string
	Imports     []string
	Models      []*TModel
	Consts      []*TConst
	Types       []*TType
	Enums       []*TEnum
}

type ModelTemplate struct {
	PackageName string
	Imports     []string
	*TModel
}

func CreateTemplateFromModelContainer(c *ModelContainer) {

}

//go:embed templates/embed_model.go.tmpl
var embedTemplate string

//go:embed templates/model.go.tmpl
var modelTemplate string

func CreateTemplateFromEmbedModels(e *EmbedModels, packageName, createPath string, modulePath string, imports ...string) (*EmbedModelTemplate, error) {
	//tmpl, err := template.New("embed").Parse(embedTemplate)
	//if err != nil {
	//	return nil, errors.New("embed model템플릿 파일 파싱 실패")
	//}
	fmt.Println(e.Enums, "wwww")

	//err = MkDir(PathSuffixCheck(createPath))
	//makePath := PathSuffixCheck(createPath)
	//if err != nil {
	//return nil, err
	//}
	//f, err := os.Create(makePath + "mg_embed.go")

	tmplData := &EmbedModelTemplate{
		Imports:     imports,
		PackageName: packageName,
	}

	for _, m := range e.Models {
		tmodel := &TModel{
			Name: m.Name,
		}

		for _, f := range m.Fields {
			//fmt.Println(reflect.TypeOf(f.Type))
			field := &TModelField{
				Upper: f.Name,
				Snake: ToSnakeCase(f.Name),
				Type:  parseFieldType(f.Type),
			}
			tmodel.Fields = append(tmodel.Fields, field)
		}
		tmplData.Models = append(tmplData.Models, tmodel)
	}

	for enumName, enum := range e.Enums {
		//fmt.Println("wwwwwww")
		tmplData.Enums = append(tmplData.Enums, &TEnum{
			Type:   enumName,
			Values: enum.Values,
		})
	}

	//for _, c := range e.Consts {
	//	hasValue := c.Value != ""
	//	fmt.Println(reflect.TypeOf(c.Type), "ww", c.Name)
	//	if a, ok := c.Type.(*ast.Ident); ok {
	//		fmt.Println("awwffffffffff", c.Value, "wwwwwwwwww")
	//		tmplData.Consts = append(tmplData.Consts, &TConst{
	//			Type:     a.Name,
	//			Name:     c.Name,
	//			Value:    c.Value,
	//			HasValue: hasValue,
	//		})
	//	}
	//}

	//for _, t := range e.Types {
	//	tmplData.Types = append(tmplData.Types, &TType{
	//		Name: t.Name,
	//		Type: t.Type.(*ast.Ident).Name,
	//	})
	//}

	//err = tmpl.Execute(f, tmplData)
	return tmplData, nil
}

func CreateTemplateFromModels(m []*Model, packageName string, createPath string, modulePath string, imports_ ...string) error {
	tmpl, err := template.New("model").Parse(modelTemplate)
	if err != nil {
		return errors.New("embed model템플릿 파일 파싱 실패")
	}

	imports := make([]string, 0)
	for _, i := range imports_ {
		if i != "go.mongodb.org/mongo-driver/v2/bson" && strings.HasPrefix(i, "modulePath") {
			imports = append(imports, i)
		}

	}

	models := m
	for _, model := range models {
		tmplData := &ModelTemplate{
			Imports:     imports,
			PackageName: packageName,
		}
		fileName := "mg_" + strings.ToLower(model.Name)
		err = MkDir(PathSuffixCheck(createPath))
		makePath := PathSuffixCheck(createPath)
		if err != nil {
			return err
		}

		f, err := os.Create(makePath + fileName + ".go")

		tmodel := &TModel{
			Name: model.Name,
		}

		for _, field := range model.Fields {
			modelField := &TModelField{
				Type:  parseFieldType(field.Type),
				Upper: field.Name,
				Snake: ToSnakeCase(field.Name),
			}
			tmodel.Fields = append(tmodel.Fields, modelField)
		}
		tmplData.TModel = tmodel
		err = tmpl.Execute(f, tmplData)
		if err != nil {
			return err
		}
	}
	return nil
}

func MkDir(d string) error {
	_, err := os.Stat(d)
	if os.IsNotExist(err) {
		err = os.MkdirAll(d, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
func PathSuffixCheck(s string) string {
	if strings.HasSuffix(s, "/") {
		return s
	}
	return s + "/"
}

func ToSnakeCase(s string) string {
	var result strings.Builder

	for i, char := range s {
		if unicode.IsUpper(char) {
			// 첫 번째 글자이거나 이전 글자가 대문자가 아닐 경우 언더스코어 추가
			if i > 0 && !unicode.IsUpper(rune(s[i-1])) {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(char))
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

func parseFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		// 배열 타입
		return fmt.Sprintf("[]%s", parseFieldType(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", parseFieldType(t.Key), parseFieldType(t.Value))
	case *ast.StructType:
		// 구조체 타입
		return "struct"
	case *ast.FuncType:
		// 함수 타입
		return "func"
	case *ast.SelectorExpr:
		return t.X.(*ast.Ident).Name + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + parseFieldType(t.X)
	default:
		return ""
	}
}

func IsValidType(typeStr string) bool {
	// 기본 타입 목록
	validTypes := map[string]reflect.Kind{
		"string":  reflect.String,
		"int":     reflect.Int,
		"int8":    reflect.Int8,
		"int16":   reflect.Int16,
		"int32":   reflect.Int32,
		"int64":   reflect.Int64,
		"uint":    reflect.Uint,
		"uint8":   reflect.Uint8,
		"uint16":  reflect.Uint16,
		"uint32":  reflect.Uint32,
		"uint64":  reflect.Uint64,
		"float32": reflect.Float32,
		"float64": reflect.Float64,
		"bool":    reflect.Bool,
		"rune":    reflect.Int32, // rune은 int32의 별칭
		"byte":    reflect.Uint8, // byte는 uint8의 별칭
	}

	// 해당 타입이 유효한지 확인
	_, exists := validTypes[typeStr]
	return exists
}
