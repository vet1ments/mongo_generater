package main

import (
	_ "embed"
	"flag"
	"fmt"
	"github.com/vet1ments/mongo_generater/internal"
	"os"
	"text/template"
)

var (
	packageNameFlag = flag.String("p", "", "package name")
	createPathFlag  = flag.String("o", "", "output path")
	inputPathFlag   = flag.String("i", "", "input path")
	modulePathFlag  = flag.String("m", "", "module path")
)

func exitProgram(s string) {
	fmt.Println(s)
	os.Exit(0)
}

//go:embed templates/embed_model.go.tmpl
var embedTemplate string

//go:embed templates/model.go.tmpl
var modelTemplate string

func main() {
	flag.Parse()

	if *packageNameFlag == "" {
		exitProgram("path 입력 없음")
	}
	packageName := *packageNameFlag

	if *inputPathFlag == "" {
		exitProgram("인풋없음")
	}
	inputPath := *inputPathFlag

	if *createPathFlag == "" {
		exitProgram("아웃풋없음")
	}
	createPath := *createPathFlag

	if *modulePathFlag == "" {
		exitProgram("모듈 패스 없음")
	}
	modulePath := *modulePathFlag

	//packageName := "test"
	//inputPath := "./test"
	//createPath := "./gen"
	//modulePath := "test"

	files, err := os.ReadDir(inputPath)
	if err != nil {
		exitProgram("디렉토리 읽기 오류")
	}

	fileNames := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, inputPath+"/"+file.Name())
		}
	}

	//var wg sync.WaitGroup
	embedTmplDatas := make([]*internal.EmbedModelTemplate, 0)
	for _, fileName := range fileNames {

		res, e1 := internal.ParseStruct(fileName)
		if e1 != nil {
			exitProgram(e1.Error())
			return
		}
		//fmt.Println(res)
		mc, e2 := internal.ParseModelFromStruct(res)
		if e2 != nil {
			exitProgram(e2.Error())
			return
		}

		embedTmplData, err := internal.CreateTemplateFromEmbedModels(mc.EmbedModels, packageName, createPath, modulePath, mc.Imports...)

		embedTmplDatas = append(embedTmplDatas, embedTmplData)

		if err != nil {
			exitProgram(err.Error())
		}
		err = internal.CreateTemplateFromModels(mc.Models, packageName, createPath, modulePath, mc.Imports...)
		if err != nil {
			exitProgram(err.Error())
		}
	}
	tmpl, err := template.New("embed").Parse(embedTemplate)
	if err != nil {
		exitProgram("embed model템플릿 파일 파싱 실패")
	}

	mergeData := &internal.EmbedModelTemplate{}

	for _, v := range embedTmplDatas {
		mergeData.Enums = append(mergeData.Enums, v.Enums...)
		mergeData.Models = append(mergeData.Models, v.Models...)
		mergeData.Imports = append(mergeData.Imports, v.Imports...)
		mergeData.PackageName = packageName
		mergeData.Types = append(mergeData.Types, v.Types...)
		mergeData.Consts = append(mergeData.Consts, v.Consts...)
	}
	err = internal.MkDir(internal.PathSuffixCheck(createPath))
	makePath := internal.PathSuffixCheck(createPath)
	if err != nil {
		exitProgram(err.Error())
	}
	f, err := os.Create(makePath + "mg_embed.go")
	err = tmpl.Execute(f, mergeData)
	if err != nil {
		exitProgram(err.Error())
	}
	fmt.Println("완료 " + createPath)
	//wg.Wait()

}
