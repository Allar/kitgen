package kitgen

import (
	"fmt"
	"text/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

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

type ServiceMethod struct {
	Name string
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
		"lower":             strings.ToLower,
		"serviceMethodList": serviceMethodList,
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
		sc.Methods = append(sc.Methods, ServiceMethod{
			Name: m.Name(),
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

	log.Printf("compiling template [%s]", templatePath)

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
