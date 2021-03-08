package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func getGoPath() string {
	temp := os.Getenv("GOPATH")
	return strings.ReplaceAll(temp, "\\", "/")
}

func autoGenCode(pkgDir string, pkgName string, tabs []string) {
	workDir := getGoPath() + "/src/bsontagtemp"
	// os.RemoveAll(workDir)
	// os.MkdirAll(workDir, 0644)

	modName := strings.Split(pkgDir, "/")[0]

	// go mod 文件
	modBuf := fmt.Sprintf(`module bsontagtemp

	go 1.13
	
	require git.dustess.com/mk-biz/%v latest
	
	replace git.dustess.com/mk-biz/%v => %v/src/%v
	`, modName, modName, getGoPath(), modName)

	ioutil.WriteFile(workDir+"/go.mod", []byte(modBuf), 0644)
	// os.RemoveAll("./temp2")
	// os.MkdirAll("./temp2", 0644)

	fileBuf := `package main

	import (
		// 路径自动替换
		model "PKG-DIR"
		"fmt"
		"os/exec"
		"io/ioutil"
		"reflect"
		"strings"
	)
	
	type fieldTag struct {
		tagName   string     // tag名
		fieldName string     // 字段名
		subField  []fieldTag // 子字段
	}
	
	// 返回<字段名、tag>
	func parseTag(obj interface{}) []fieldTag {
		return parseTagImpl(reflect.TypeOf(obj), reflect.ValueOf(obj))
	}

	func isBaseType(objType reflect.Type) bool {
		if objType.Kind() == reflect.Ptr {
			return isBaseType(objType.Elem())
		}
	
		switch objType.Kind() {
		case reflect.Struct /*, reflect.Map, reflect.Array*/ :
			return false
	
		default:
			return true
		}
	}
	
	func parseTagImpl(objType reflect.Type, objValue reflect.Value) []fieldTag {
		if objType.Kind() == reflect.Ptr {
			return parseTagImpl(objType.Elem(), objValue)
		}
	
		var fieldTags []fieldTag
		fmt.Println(objType.Name(), objType.Kind(), objValue.Kind())
		if objType.Kind() == reflect.Struct {
			for i := 0; i < objType.NumField(); i++ {
				bson := objType.Field(i).Tag.Get("bson")
				if bson == "" || bson == "-" {
					continue
				}

				bson = strings.TrimSuffix(bson, ",omitempty")
	
				if isBaseType(objType.Field(i).Type) {
					temp := fieldTag{
						tagName:   bson,
						fieldName: objType.Field(i).Name,
					}
	
					fieldTags = append(fieldTags, temp)
					continue
				}
	
				if temps := parseTagImpl(objType.Field(i).Type, objValue.Field(i)); len(temps) > 0 {
	
					if bson == ",inline" {
						fieldTags = append(fieldTags, temps...)
					} else {
						myself := fieldTag{
							tagName:   bson,
							fieldName: "_" + objType.Field(i).Name,
						}
						fieldTags = append(fieldTags, myself)
	
						temp := fieldTag{
							tagName:   bson,
							fieldName: objType.Field(i).Name,
							subField:  temps,
						}
						fieldTags = append(fieldTags, temp)
					}
				}
			}
		}
	
		return fieldTags
	}
	
	// 返回：(结构体结构、赋值)
	func format(tabName string, tabPrefix string, tagPrefix string, fieldTags []fieldTag) (string, string) {
		structStr := fmt.Sprintf(tabName + " struct {\n")
	tagStr := ""

	for _, v := range fieldTags {
		if len(v.subField) > 0 {
			newPrefix := tabName
			if tabPrefix != "" {
				newPrefix = tabPrefix + "." + newPrefix
			}

			newTagPrefix := v.tagName
			if tagPrefix != "" {
				newTagPrefix = tagPrefix + "." + newTagPrefix
			}

			temp1, temp2 := format(v.fieldName, newPrefix, newTagPrefix, v.subField)
			structStr += temp1
			tagStr += temp2
		} else {
			structStr += fmt.Sprintf("%v string\n", v.fieldName)
			fullTags := v.tagName
			if tagPrefix != "" {
				fullTags = tagPrefix + "." + fullTags
			}
			if tabPrefix == "" {
				tagStr += fmt.Sprintf("FN.%v.%v = \"%v\"\n", tabName, v.fieldName, fullTags)
			} else {
				tagStr += fmt.Sprintf("FN.%v.%v.%v = \"%v\"\n", tabPrefix, tabName, v.fieldName, fullTags)
			}
		}
	}

	structStr += "}\n"

	return structStr, tagStr
	}
	
	func main() {
		// 包名自动替换
		fileBuf := "package PKG-NAME\n"

		fileBuf += "/* ------此文件为自动生成，不要更改---------\n---------此文件为自动生成，不要更改----------\n---------此文件为自动生成，不要更改----------*/"
		fileBuf += "\n\n"

		// 表名自动替换
		var tabs = []interface{}{TABS-NAME}
		allTagStr := "func init() {\n"
		fnBuf := fmt.Sprintln("var FN = struct {")
		for _, tab := range tabs {
			tabName := reflect.TypeOf(tab).Name()
			fileBuf += "type tag"
			m := parseTag(tab)
			structStr, tagStr := format(tabName, "", "", m)
			fileBuf += structStr
			allTagStr += tagStr
	
			fnBuf += fmt.Sprintf("%v %v\n", tabName, "tag"+tabName)
		}
		fnBuf += "}{}\n"
		allTagStr += "}\n"
	
		fileBuf += fnBuf
		fileBuf += allTagStr
	
		// 路径自动替换
		ioutil.WriteFile("FILE-DIR", []byte(fileBuf), 0644)

		// 格式化
		cmd := exec.Command("go", "fmt", "FILE-DIR")
		cmd.Start()
		cmd.Wait()
	}
	`

	fileBuf = strings.Replace(fileBuf, "PKG-DIR", "git.dustess.com/mk-biz/"+pkgDir, -1)
	fileBuf = strings.Replace(fileBuf, "PKG-NAME", pkgName, -1)
	for k, _ := range tabs {
		tabs[k] = "model." + tabs[k] + "{}"
	}
	fileBuf = strings.Replace(fileBuf, "TABS-NAME", strings.Join(tabs, ", "), -1)
	// gopath := os.Getenv("GOPATH")
	// gopath = strings.Replace(gopath, "\\", "\\\\", -1)
	fileDir := getGoPath() + "/src/" + pkgDir + "/auto_tag.go"
	fileBuf = strings.Replace(fileBuf, "FILE-DIR", fileDir, -1)
	os.MkdirAll(workDir+"/"+pkgDir, 0644)
	err := ioutil.WriteFile(workDir+"/"+pkgDir+"/main.go", []byte(fileBuf), 0644)
	fmt.Println(err)

	// 运行生成的main.go
	cmd := exec.Command("go", "run", workDir+"/"+pkgDir+"/main.go")
	cmd.Dir = workDir + "/" + pkgDir
	if err := cmd.Start(); err != nil {
		fmt.Printf("运行main.go失败, err = %v, file = %v\n", err, workDir+"/"+pkgDir+"/main.go")
		return
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("运行main.go失败-2, err = %v, file = %v\n", err, workDir+"/"+pkgDir+"/main.go")
		return
	}
}

func parseTabImpl(text string) []string {

	t := make(map[string]bool)
	flag := []string{"\n", "\r\n"}
	for _, v := range flag {
		eee := strings.Replace(text, v, "abcdefg", -1)
		reg1 := regexp.MustCompile(`// @table:(\w*)abcdefgtype (\w+) struct {`)
		kkk := reg1.FindAllString(eee, -1)

		// var tabs []string
		for _, v := range kkk {
			a := regexp.MustCompile(`type\s+(\w+)\s+struct`)
			f := a.FindAllStringSubmatch(v, -1)
			for _, x := range f {
				// tabs = append(tabs, x[1])
				t[x[1]] = true
			}
		}
	}

	var tabs []string
	for k := range t {
		tabs = append(tabs, k)
	}

	return tabs
}

type packInfo struct {
	pkgName string
	tabs    []string
}

func parseTabs(parent string) []packInfo {
	gopath := os.Getenv("GOPATH")
	gopath = strings.Replace(gopath, "\\", "\\\\", -1)
	// dir = gopath + "/src/" + dir
	fileinfo, err := ioutil.ReadDir(parent)
	if err != nil {
		panic(err)
	}

	var pInfo []packInfo

	var tabs []string
	pkgName := ""
	for _, fi := range fileinfo {
		if !fi.IsDir() {
			// files = append(files, fi.Name())
			if fi.Name() == "order.go" {
				fmt.Println(1)
			}
			buf, err := ioutil.ReadFile(parent + "/" + fi.Name())
			if err != nil {
				panic("读取" + fi.Name() + "文件失败，终止生成tag")
			}
			tabs = append(tabs, parseTabImpl(string(buf))...)
			if pkgName == "" {
				a := regexp.MustCompile(`package\s+(\w+)\s+`)
				f := a.FindAllStringSubmatch(string(buf), -1)
				pkgName = f[0][1]
				pkgName = strings.TrimPrefix(parent, os.Getenv("GOPATH")+"/src/")
			}
		} else {
			if temp := parseTabs(parent + "/" + fi.Name()); len(temp) > 0 {
				pInfo = append(pInfo, temp...)
			}
		}
	}

	pInfo = append(pInfo, packInfo{
		pkgName: pkgName,
		tabs:    tabs,
	})

	return pInfo
}

func env(dir string) string {
	workDir, _ := os.Getwd()
	if dir != "" {
		workDir = workDir + "/" + dir
	}

	workDir, err := filepath.Abs(workDir)
	if err != nil {
		fmt.Println("目录不存在, err = ", err)
		return ""
	}

	workDir = strings.ReplaceAll(workDir, "\\", "/")
	// cc, err2 := filepath.Rel(workDir+"/exam", "C:/work/go/bin")
	// fmt.Println(err2, cc)

	fmt.Printf("开始搜寻当前目录(%v)...\n", workDir)
	// if !strings.HasPrefix(workDir, getGoPath()) {
	// 	fmt.Println("被扫描目录不在GOPATH下，终止运行")
	// 	return ""
	// }

	if !strings.HasSuffix(workDir, "/pkg") {
		fmt.Println("当前目录不在pkg目录，终止运行")
		return ""
	}

	diff, err := filepath.Rel(getGoPath()+"/src", workDir)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return strings.ReplaceAll(diff, "\\", "/")
}

func main() {
	dir := flag.String("d", "", "请输入pkg目录，例如mk-pay-svc/pkg、 ./pkg")
	flag.Parse()
	pkg := env(*dir)
	if pkg == "" {
		return
	}

	fmt.Printf("开始解析%v目录下的model子目录\n", pkg)
	modelDir := pkg
	// 分析指定包中有几张表，返回表名和包名
	// tabs, pkgName := parseTabs(modelDir)
	modelDir = os.Getenv("GOPATH") + "/src/" + modelDir
	pInfo := parseTabs(modelDir)
	hasTabs := false
	for _, v := range pInfo {
		if len(v.tabs) > 0 {
			hasTabs = true
			fmt.Println(v.pkgName, v.tabs)
		}
	}

	if !hasTabs {
		return
	}

	// 自动生成源码
	workDir := getGoPath() + "/src/bsontagtemp"
	os.RemoveAll(workDir)
	os.MkdirAll(workDir, 0644)

	for _, v := range pInfo {
		if len(v.tabs) > 0 {
			temps := strings.Split(v.pkgName, "/")
			autoGenCode(v.pkgName, temps[len(temps)-1], v.tabs)
		}
	}

	// 运行源码，反射得到每个字段的tag

	// 生成tag源码

	// 格式化源码
}
