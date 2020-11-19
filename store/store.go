package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"github.com/google/uuid"
	"strings"
)

var pathToSettings = "./settings.json"

type Config struct {
	Connections []Connect
}

type Connect struct {
	Id        string
	Name      string
	Url       string
	Preferred []PreferredField
}

type PreferredField struct {
	IndexPrefix string
	Fields      []string
}

func (pref *PreferredField) GetFieldsAsText() string {
	var sb strings.Builder
	for index, field := range pref.Fields {
		sb.WriteString(field)
		if index != len(pref.Fields)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func (connect *Connect) AddPreferedField(indexPrefix string, fields string) {
	field := PreferredField{
		IndexPrefix: indexPrefix,
	}
	field.SetFieldsFromText(fields)
	connect.Preferred = append(connect.Preferred, field)

}

func (pref *PreferredField) SetFieldsFromText(str string) {
	var res []string
	str = strings.ReplaceAll(str, " ", "")
	res = strings.Split(str, ",")
	pref.Fields = res
}

func (connect *Connect) DeletePrefFieldByIndex(index int) {
	connect.Preferred = append(connect.Preferred[:index], connect.Preferred[index+1:]...)
}

func (config *Config) AddConnection(connect *Connect) {
	connect.Id = uuid.New().String()
	config.Connections = append(config.Connections, *connect)
}

func (config *Config) RemoveConnection(index int) {
	config.Connections = append(config.Connections[:index], config.Connections[index+1:]...)
	config.Save()
}

func (config *Config) Save() {
	file, _ := json.MarshalIndent(config, "", " ")
	err := ioutil.WriteFile(pathToSettings, file, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func NewStore() Config {
	configFile, err := os.Open(pathToSettings)
	if err != nil {
		fmt.Println("error: ", err)
	}

	config := Config{}

	jsonParser := json.NewDecoder(configFile)
	err1 := jsonParser.Decode(&config)
	if err1 != nil {
		fmt.Println(err1)
	}

	fmt.Println(config, err1)
	return config
}
