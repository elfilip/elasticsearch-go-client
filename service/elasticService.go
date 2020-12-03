package service

import (
	"EsClient2/elastic"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var Num int = 1

type EsData struct {
	Elastic *elastic.Elastic
	Index   *elastic.Index
	Data    []map[string]interface{}
	Query   string
}

func NewEsData(elastic *elastic.Elastic, index *elastic.Index, data []map[string]interface{}, query string) *EsData {
	return &EsData{Elastic: elastic, Index: index, Data: data, Query: query}
}

func (esData *EsData) GetDocId(index int) string {
	return esData.Data[index]["_id"].(string)
}

func (esData *EsData) UpdatePath(docIndex int, path string, data string) string{
	var curField interface{} = esData.Data[docIndex]["_source"]
	pathIndex := 0
	pathArr := strings.Split(path, ".")

	for pathIndex < len(pathArr)-1 {
		switch curField.(type) {
		case map[string]interface{}:
			curField = curField.(map[string]interface{})[pathArr[pathIndex]]
			pathIndex = pathIndex + 1
		default:
			panic("Updating not updatable fields" + path)
		}
	}
	var convertedData interface{}
	switch curField.(map[string]interface{})[pathArr[pathIndex]].(type) {
	case string:
		convertedData = data
	case float32:
		convertedData, _ = strconv.ParseFloat(data, 32)
	case float64:
		convertedData, _ = strconv.ParseFloat(data, 64)
	case int:
		convertedData, _ = strconv.Atoi(data)
	case bool:
		convertedData, _ = strconv.ParseBool(data)
	default:
		json.Unmarshal([]byte(data), &convertedData)
	}

	curField.(map[string]interface{})[pathArr[pathIndex]] = convertedData
	updatedData,_ := json.MarshalIndent(esData.Data[docIndex]["_source"], "", " ")
	fmt.Println(string(updatedData))
	return string(updatedData)
}

func (esData *EsData) RefreshData() {
	esData.Data = esData.Elastic.LoadNFirstDocs(20, esData.Index.Name, esData.Query)
}

func (esData *EsData) CanUpdate(index int, path string) {
	canUpdate(esData.Data[index]["_source"], 0, strings.Split(path, "."))
}

func canUpdate(data interface{}, pathIndex int, path []string) bool {
	if pathIndex == len(path) {
		return true
	}
	if data == nil {
		return false
	}
	switch data.(type) {
	case map[string]interface{}:
		return canUpdate(data.(map[string]interface{})[path[pathIndex]], pathIndex+1, path)
	default:
		return false
	}
}

func (esData *EsData) GetStringFromESFirstLevelField(field string, docIndex int, format bool) string {
	if len(field) > 0 {
		return getStringFromEsPath(strings.Split(field, "."), esData.Data[docIndex]["_source"], 0, format)
	} else {
		return getStringFromEsPath([]string{}, esData.Data[docIndex]["_source"], 0, format)
	}
}

func getStringFromEsPath(path []string, data interface{}, index int, format bool) string {
	if index == len(path) {
		return ConvertAnyToString(data, format)
	}
	if data == nil {
		return ""
	}
	switch data.(type) {
	case map[string]interface{}:
		return getStringFromEsPath(path, data.(map[string]interface{})[path[index]], index+1, format)
	case []interface{}:
		return getStringFromEsPath(path, data.([]interface{})[0], index, format)
	default:
		return ""
	}
}

func ConvertAnyToString(field interface{}, format bool) string {
	var text string
	switch field.(type) {
	case map[string]interface{}:
		if format {
			val, _ := json.MarshalIndent(field.(map[string]interface{}), "", "  ")
			text = string(val)
		} else {
			val, _ := json.Marshal(field.(map[string]interface{}))
			text = string(val)
		}
	case []interface{}:
		if format {
			val, _ := json.MarshalIndent(field.([]interface{}), "", "  ")
			text = string(val)
		} else {
			val, _ := json.Marshal(field.([]interface{}))
			text = string(val)
		}
	case string:
		text = fmt.Sprintf("%s", field)
	case float64:
		text = fmt.Sprintf("%f", field)
		num := field.(float64)
		if float64(int64(num)) == num {
			text = fmt.Sprintf("%d", int64(num))
		}
	case float32:
		text = fmt.Sprintf("%f", field)
		num := field.(float32)
		if float32(int64(num)) == num {
			text = fmt.Sprintf("%d", int64(num))
		}
	default:
		text = fmt.Sprintf("%#v", field)
	}
	return text
}
