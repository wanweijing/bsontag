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

	"golang.org/x/sync/errgroup"
)

func getGoPath() string {
	temp := os.Getenv("GOPATH")
	return strings.ReplaceAll(temp, "\\", "/")
}

func autoGenCode(pkgDir string, pkgName string, tabs []string) string {
	workDir := getGoPath() + "/src/bsontagtemp"
	// os.RemoveAll(workDir)
	// os.MkdirAll(workDir, 0644)

	modName := strings.Split(pkgDir, "/")[0]
	gitGroup := "biz"
	if pkgDir == "pkg" {
		gitGroup = "base"
	}

	// go mod 文件
	modBuf := fmt.Sprintf(`module bsontagtemp

	go 1.13
	
	require git.dustess.com/mk-%v/%v latest
	
	replace git.dustess.com/mk-%v/%v => %v/src/%v
	`, gitGroup, modName, gitGroup, modName, getGoPath(), modName)

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
		"os"
	)
	
	type fieldTag struct {
		tagName   string     // tag名
		fieldName string     // 字段名
		subField  []fieldTag // 子字段
	}
	
	// 返回<字段名、tag>
	func parseTag(obj interface{}) []fieldTag {
		return parseTagImpl(reflect.TypeOf(obj))
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
	
	func parseTagImpl(objType reflect.Type) []fieldTag {
		if objType.Kind() == reflect.Ptr {
			return parseTagImpl(objType.Elem())
		}
	
		var fieldTags []fieldTag
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
	
				if temps := parseTagImpl(objType.Field(i).Type); len(temps) > 0 {
	
					if bson == ",inline" {
						fieldTags = append(fieldTags, temps...)
					} else {
						myself := fieldTag{
							tagName:   bson,
							fieldName: objType.Field(i).Name + "_",
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

	func checkFieldTag(tags []fieldTag) bool {
		for i, tag1 := range tags {
			if !checkFieldTag(tag1.subField) {
				return false
			}
	
			for _, tag2 := range tags[i+1:] {
				if tag1.fieldName == tag2.fieldName {
					fmt.Printf("检测到冲突, field name = %v, tag name = %v\n", tag1.fieldName, tag1.tagName)
					return false
				}
			}
		}
	
		return true
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
			if !checkFieldTag(m) {
				fmt.Println("检测到字段冲突，终止生成tag")
				os.Exit(3)
			}
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

	fileBuf = strings.Replace(fileBuf, "PKG-DIR", "git.dustess.com/mk-"+gitGroup+"/"+pkgDir, -1)
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
	if err := ioutil.WriteFile(workDir+"/"+pkgDir+"/main.go", []byte(fileBuf), 0644); err != nil {
		fmt.Println(workDir+"/"+pkgDir+"/main.go写入失败, err = ", err)
	} else {
		fmt.Println(workDir + "/" + pkgDir + "/main.go成功写入")
	}

	return workDir + "/" + pkgDir + "/main.go"

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
			if !strings.HasSuffix(fi.Name(), ".go") {
				continue
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
			if strings.HasPrefix(fi.Name(), ".") {
				continue
			}

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

	var mainFiles []string
	for _, v := range pInfo {
		if len(v.tabs) > 0 {
			temps := strings.Split(v.pkgName, "/")
			if mainFile := autoGenCode(v.pkgName, temps[len(temps)-1], v.tabs); mainFile != "" {
				mainFiles = append(mainFiles, mainFile)
			}
		}
	}

	// 运行源码，反射得到每个字段的tag，运行策略：先单独运行第一个main.go，之后并发运行所有的main.go
	fn := func(mainFile string) error {
		fmt.Printf("开始运行%v...\n", mainFile)
		cmd := exec.Command("go", "run", mainFile)
		cmd.Dir = strings.TrimSuffix(mainFile, "/main.go")
		if err := cmd.Start(); err != nil {
			fmt.Printf("运行main.go失败, err = %v, file = %v\n", err, mainFile)
			return err
		}

		if err := cmd.Wait(); err != nil {
			fmt.Printf("运行main.go失败-2, err = %v, file = %v\n", err, mainFile)
			return err
		}

		fmt.Println(mainFile, "成功运行")

		return nil
	}

	if len(mainFiles) > 0 {
		fn(mainFiles[0])

		var wg errgroup.Group
		for _, v := range mainFiles[1:] {
			func(file string) {
				wg.Go(func() error {
					return fn(file)
				})
			}(v)
		}
		// for _, act := range uaList {
		// 	func(a *vs.GroupWXClueActivity) {
		// 		wg.Go(func() error {
		// 			if err := loadClueSubActivityDetail(ctx.Request.Context(), a, session.Account, session.AccountDB); err != nil {
		// 				return err
		// 			}
		// 			return nil
		// 		})
		// 	}(act)
		// }
		if err := wg.Wait(); err != nil {
			fmt.Println(err)
			return
		}
	}

	// 生成tag源码

	// 格式化源码
}
