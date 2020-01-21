package kitgen

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/allar/kitgen/assets"
	"github.com/allar/source"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func checkPathDoesNotExist(pathName string) error {
	path := filepath.Join(pathName)
	if _, serr := os.Stat(path); serr != nil {
		if os.IsNotExist(serr) {
			return nil
		}
		return serr
	}

	return fmt.Errorf("path [%s] exists", pathName)
}

func createPath(pathName string) error {
	path := filepath.Join(pathName)
	if _, serr := os.Stat(path); serr != nil {
		if !os.IsNotExist(serr) { // failed read
			return fmt.Errorf("failed read of path [%s]: [%s]", path, serr)
		}
		return os.MkdirAll(path, os.ModeDir|0777)
	} else {
		return fmt.Errorf("path [%s] already exists", path)
	}
}

type MethodParameter struct {
	Name string
	Type string
}

type ServiceMethod struct {
	Name              string
	Parameters        []MethodParameter
	Results           []MethodParameter
	HasImplementation bool
}

type ServiceConfig struct {
	Name          string
	RepoPath      string
	TemplateFuncs template.FuncMap
	Methods       []ServiceMethod
	IsNewService  bool
}

func (sc ServiceConfig) sanitize() ServiceConfig {
	sc.Name = strings.Title(sc.Name)

	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	index := strings.Index(path, "/src/")
	if index != -1 {
		sc.RepoPath = path[index+5:]
	} else {
		log.Fatal("Could not identifiy project directory, is project not in $GOHOME/src?")
	}

	sc.TemplateFuncs = template.FuncMap{
		"lower":                       strings.ToLower,
		"title":                       strings.Title,
		"serviceMethodList":           serviceMethodList,
		"serviceMethodInvokeArgList":  serviceMethodInvokeArgList,
		"serviceMethodResultList":     serviceMethodResultList,
		"serviceMethodBuildLogParams": serviceMethodBuildLogParams,
		"separator": func(s string) func() string {
			i := -1
			return func() string {
				i++
				if i == 0 {
					return ""
				}
				return s
			}
		},
	}

	return sc
}

func (sc ServiceConfig) path() string {
	return strings.ToLower(sc.Name)
}

func serviceMethodList(sc ServiceConfig) string {
	names := []string{}
	for _, n := range sc.Methods {
		names = append(names, "\""+n.Name+"\"")
	}
	return strings.Join(names, ", ")
}

func serviceMethodInvokeArgList(sm ServiceMethod, argPrefix string) string {
	var invokeArgString string
	for _, p := range sm.Parameters {
		actualPrefix := argPrefix
		varName := p.Name
		if p.Name == "ctx" {
			actualPrefix = ""
		} else {
			varName = strings.Title(varName)
		}
		invokeArgString = invokeArgString + actualPrefix + varName + ", "
	}
	return strings.TrimRight(invokeArgString, ", ")
}

func serviceMethodResultList(sm ServiceMethod) string {
	var resultListString string
	for _, r := range sm.Results {
		resultListString = resultListString + r.Name + ", "
	}
	return strings.TrimRight(resultListString, ", ")
}

func serviceMethodBuildLogParams(sm ServiceMethod) string {
	var resultLogString string

	for _, p := range sm.Parameters {
		if p.Name == "ctx" {
			continue
		}
		resultLogString = resultLogString + "\"" + strings.ToLower(p.Name) + "\", " + strings.ToLower(p.Name) + ", "
	}

	for _, r := range sm.Results {
		resultLogString = resultLogString + "\"" + strings.ToLower(r.Name) + "\", " + strings.ToLower(r.Name) + ", "
	}

	return strings.TrimRight(resultLogString, ", ")
}

func BuildServiceConfigFromPath(path string) ServiceConfig {
	sc := ServiceConfig{
		Name: path,
	}.sanitize()

	err := checkPathDoesNotExist(path)
	if err == nil {
		sc.IsNewService = true
		return sc
	}

	sc.IsNewService = false

	serviceFilePath := filepath.Join(sc.path(), "service", "service.go")
	serviceFileBuf, err := ioutil.ReadFile(serviceFilePath)
	check(err)

	serviceSource, err := source.New(string(serviceFileBuf))
	check(err)

	serviceInterface, err := serviceSource.GetInterface(sc.Name + "Service")
	check(err)

	methods := serviceInterface.Methods()
	for _, m := range methods {

		methodParams := []MethodParameter{}

		params := m.Params()
		for i, p := range params {
			if i == 0 {
				if p.Name != "ctx" || p.Type.String() != "context.Context" {
					log.Fatalf("service interface method's [%s] first parameter should be \"ctx context.Context\", got [%s %s]", m.Name(), p.Name, p.Type.String())
				}
			}
			methodParams = append(methodParams, MethodParameter{
				Name: p.Name,
				Type: p.Type.String(),
			})
		}

		methodResults := []MethodParameter{}
		results := m.Results()
		for i, r := range results {
			if i == len(results)-1 {
				if r.Name != "err" || r.Type.String() != "error" {
					log.Fatalf("service interface method's [%s] last return result should be \"err Error\", got [%s %s]", m.Name(), r.Name, r.Type.String())
				}
			}
			methodResults = append(methodResults, MethodParameter{
				Name: r.Name,
				Type: r.Type.String(),
			})
		}

		serviceMethodImplementationFound := false
		serviceFunc, _ := serviceSource.GetFunction(m.Name())

		if serviceFunc != nil {
			matchingName := serviceFunc.Name() == m.Name()
			matchingParams := reflect.DeepEqual(serviceFunc.Params(), params)
			matchingResults := reflect.DeepEqual(serviceFunc.Results(), results)

			serviceMethodImplementationFound = matchingName && matchingParams && matchingResults
			//log.Printf("func is implemented: [%v], matchingName: [%v], matchingParams: [%v], matchingResults: [%v]", serviceMethodImplementationFound, matchingName, matchingParams, matchingResults)
		}

		sc.Methods = append(sc.Methods, ServiceMethod{
			Name:              m.Name(),
			Parameters:        methodParams,
			Results:           methodResults,
			HasImplementation: serviceMethodImplementationFound,
		})
	}

	return sc
}

func CreateService(serviceName string) error {
	serviceConfig := BuildServiceConfigFromPath(serviceName)
	err := createServiceCode(serviceConfig)
	if err != nil {
		return err
	}

	// If newly created, re-run create service to finish generating files
	if serviceConfig.IsNewService {
		return CreateService(serviceName)
	}

	return nil
}

func CompileTemplateToFile(templatePath string, outputFilePath string, serviceConfig ServiceConfig) error {
	if !strings.HasSuffix(templatePath, ".tmpl") {
		return nil
	}

	if !serviceConfig.IsNewService && strings.Contains(templatePath, "/service/service.go.tmpl") {
		log.Printf("skipping service template")
		return nil
	}

	log.Printf("compiling template [%s]\n", templatePath)

	data, ok := assets.FS.String(templatePath)
	if !ok {
		return fmt.Errorf("failed to read template [%s]", templatePath)
	}

	newTemplate, err := template.New("template").Funcs(serviceConfig.TemplateFuncs).Parse(data)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(outputFilePath), os.ModeDir|0755)
	if err != nil {
		return err
	}

	f, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = newTemplate.Execute(f, serviceConfig)
	if err != nil {
		return err
	}

	return nil

}

func CompileTemplatesToPath(templateFolder string, outputPath string, serviceConfig ServiceConfig) error {
	dirs, err := assets.FS.GetSubDirs(filepath.Join("/assets/files/templates/", templateFolder))
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		err := CompileTemplatesToPath(path.Join(templateFolder, dir), path.Join(outputPath, dir), serviceConfig)
		if err != nil {
			return err
		}
	}

	currentTemplatePath := filepath.Join("/assets/files/templates/", templateFolder)
	files, err := assets.FS.GetSubFiles(currentTemplatePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		templateFilePath := filepath.Join(currentTemplatePath, file)
		outputFilePath := filepath.Join(outputPath, strings.TrimRight(file, ".tmpl"))
		err := CompileTemplateToFile(templateFilePath, outputFilePath, serviceConfig)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return err
}

func createServiceCode(serviceConfig ServiceConfig) error {

	err := CompileTemplatesToPath("service_gen", serviceConfig.path(), serviceConfig)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}
