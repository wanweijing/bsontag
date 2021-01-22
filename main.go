package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func autoGenCode(pkgDir string, pkgName string, tabs []string) {
	os.RemoveAll("./temp2")
	os.MkdirAll("./temp2", 0644)

	fileBuf := `package main

	import (
		// 路径自动替换
		model "PKG-DIR"
		"fmt"
		"io/ioutil"
		"reflect"
	)
	
	type fieldTag struct {
		tagName  string
		subField map[string]fieldTag
	}
	
	// 返回<字段名、tag>
	func parseTag(obj interface{}) map[string]fieldTag {
		return parseTagImpl(reflect.TypeOf(obj), reflect.ValueOf(obj))
	}
	
	func parseTagImpl(objType reflect.Type, objValue reflect.Value) map[string]fieldTag {
		m := make(map[string]fieldTag)
		if objValue.Kind() == reflect.Struct {
			for i := 0; i < objType.NumField(); i++ {
				bson := objType.Field(i).Tag.Get("bson")
				if bson == "" {
					panic(objType.Field(i).Name + "bson非法定义，终止生成tag")
				}
	
				if objType.Field(i).Type.Kind() != reflect.Struct {
					fmt.Println(objType.Field(i).Name, objType.Field(i).Tag.Get("bson"))
	
					m[objType.Field(i).Name] = fieldTag{tagName: bson}
	
				} else {
					subField := parseTagImpl(objType.Field(i).Type, objValue.Field(i))
					m[objType.Field(i).Name] = fieldTag{tagName: bson, subField: subField}
				}
			}
		}
	
		return m
	}
	
	// 返回：(结构体结构、赋值)
	func format(tabName string, tabPrefix string, tagPrefix string, m map[string]fieldTag) (string, string) {
		structStr := fmt.Sprintf(tabName + " struct {\n")
		tagStr := ""
	
		for name, tags := range m {
	
			if tags.subField != nil {
	
				newPrefix := tabName
				if tabPrefix != "" {
					newPrefix = tabPrefix + "." + newPrefix
				}
				newTagPrefix := tags.tagName
				if tagPrefix != "" {
					newTagPrefix = tagPrefix + "." + newTagPrefix
				}
				temp1, temp2 := format(name, newPrefix, newTagPrefix, tags.subField)
				structStr += temp1
				tagStr += temp2
			} else {
				structStr += fmt.Sprintf("%v string\n", name)
				fullTags := tags.tagName
				if tagPrefix != "" {
					fullTags = tagPrefix + "." + fullTags
				}
				if "" == tabPrefix {
					tagStr += fmt.Sprintf("FN.%v.%v = \"%v\"\n", tabName, name, fullTags)
				} else {
					tagStr += fmt.Sprintf("FN.%v.%v.%v = \"%v\"\n", tabPrefix, tabName, name, fullTags)
				}
			}
		}
	
		structStr += "}\n"
	
		return structStr, tagStr
	}
	
	func main() {
		// 包名自动替换
		fileBuf := "package PKG-NAME\n"
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
	}
	`

	fileBuf = strings.Replace(fileBuf, "PKG-DIR", pkgDir, -1)
	fileBuf = strings.Replace(fileBuf, "PKG-NAME", pkgName, -1)
	for k, _ := range tabs {
		tabs[k] = "model." + tabs[k] + "{}"
	}
	fileBuf = strings.Replace(fileBuf, "TABS-NAME", strings.Join(tabs, ", "), -1)
	gopath := os.Getenv("GOPATH")
	gopath = strings.Replace(gopath, "\\", "\\\\", -1)
	fileDir := gopath + "/src/" + pkgDir + "/auto_tag.go"
	fileBuf = strings.Replace(fileBuf, "FILE-DIR", fileDir, -1)
	err := ioutil.WriteFile("./temp2/main.go", []byte(fileBuf), 0644)
	fmt.Println(err)
}

func parseTabImpl(text string) []string {
	eee := strings.Replace(text, "\r\n", "abcdefg", -1)
	reg1 := regexp.MustCompile(`// @table:(\w*)abcdefgtype (\w+) struct {`)
	kkk := reg1.FindAllString(eee, -1)

	var tabs []string
	for _, v := range kkk {
		a := regexp.MustCompile(`type\s+(\w+)\s+struct`)
		f := a.FindAllStringSubmatch(v, -1)
		for _, x := range f {
			tabs = append(tabs, x[1])
		}
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

func main() {
	// t := reflect.TypeOf(nil)
	os.Args = append(os.Args, "mk-pay-svc/pkg")

	// 扫描model目录(model目录 先通过命令行参数传入，后面再做自动扫描)
	if len(os.Args) < 2 {
		fmt.Println("请输入model目录(相对于GOPATH)")
		os.Exit(1)
	}

	modelDir := os.Args[1]
	// 分析指定包中有几张表，返回表名和包名
	// tabs, pkgName := parseTabs(modelDir)
	modelDir = os.Getenv("GOPATH") + "/src/" + modelDir
	pInfo := parseTabs(modelDir)
	for _, v := range pInfo {
		if len(v.tabs) > 0 {
			fmt.Println(v.pkgName, v.tabs)
		}
	}
	// 自动生成源码
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
