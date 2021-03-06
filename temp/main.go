package main

import (
	// 路径自动替换
	"fmt"
	"io/ioutil"
	"reflect"

	model "github.com/wanweijing/bsontag/exam"
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
	fileBuf := "package exam\n"
	// 表名自动替换
	var tabs = []interface{}{model.AAA{}, model.TTTT{}}
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
	ioutil.WriteFile("../exam/auto_tag.go", []byte(fileBuf), 0644)
}
