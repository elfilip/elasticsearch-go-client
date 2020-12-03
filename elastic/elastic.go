package elastic

import (
	"EsClient2/store"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
)

const urlGetIndices string = "/_cat/indices?format=json"
const urlRandomSearch string = "/_search?size="
const urlUpdateDoc string = "/_update/"
const urlFlush string = "/_flush"
const urlAlias string = "/_cat/aliases?format=json"

type Elastic struct {
	Connect store.Connect
	Indices []Index
	Aliases []Alias
	client  *resty.Client
}

type Index struct {
	Name     string `json:"index"`
	DocCount int64  `json:"operation"`
	Size     string `json:"store.size"`
	Health   string `json:"health"`
	Aliases  []string
}

type Alias struct {
	Alias string
	Index string
	IndexInner Index
}

func NewElastic(connect *store.Connect) *Elastic {
	elastic := Elastic{
		Connect: *connect,
		client:  resty.New(),
	}
	return &elastic
}

func (elastic *Elastic) ConnectToES() bool {
	resp, err := elastic.client.R().Get(elastic.Connect.Url + urlGetIndices)
	es := make([]Index, 1)
	if err != nil {
		fmt.Println(err)
		return false
	}
	json.Unmarshal(resp.Body(), &es)
	elastic.Indices = es
	aliasesMap, indexToAliasMap := elastic.GetAllAliases()

	for index, esIndex := range es {
		es[index].Aliases = aliasesMap[esIndex.Name]
		if val, contains := indexToAliasMap[esIndex.Name]; contains {
			val.IndexInner = esIndex
			elastic.Aliases = append(elastic.Aliases, val)
		}
	}
	fmt.Println(string(resp.Body()))
	fmt.Println(es)
	return true
}

func (elastic *Elastic) LoadNFirstDocs(count int, index string, query string) []map[string]interface{} {
	if len(query) == 0 {
		query = "*"
	}
	fmt.Println(elastic.Connect.Url + "/" + index + urlRandomSearch + strconv.Itoa(97) + "&q=" + query)
	resp, err := elastic.client.R().Get(elastic.Connect.Url + "/" + index + urlRandomSearch + strconv.Itoa(97) + "&q=" + query)
	if err != nil {
		fmt.Println(err)
	}
	var res map[string]interface{}
	json.Unmarshal(resp.Body(), &res)

	var finalres []map[string]interface{}
	for _, val := range res["hits"].(map[string]interface{})["hits"].([]interface{}) {
		docSource := val.(map[string]interface{})
		finalres = append(finalres, docSource)
	}
	return finalres
}

func (elastic *Elastic) UpdateDoc(id string, data string, index *Index) {
	updateUrl := elastic.Connect.Url + "/" + index.Name + urlUpdateDoc + id + "?refresh=true"
	fmt.Println(updateUrl)
	data = "{ \"doc\":" + data + "}"
	resp, err := elastic.client.R().SetHeader("content-type", "application/json").SetBody(data).Post(updateUrl)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.String())
	}
	elastic.FlushIndex(index)
}

func (elastic *Elastic) FlushIndex(index *Index) {
	flushUrl := elastic.Connect.Url + "/" + index.Name + urlFlush
	resp, err := elastic.client.R().Post(flushUrl)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.Error())
	}
}

func (elastic *Elastic) GetAllAliases() (map[string][]string, map[string] Alias) {
	aliasUrl := elastic.Connect.Url + urlAlias
	resp, err := elastic.client.R().Get(aliasUrl)
	if err != nil {
		fmt.Println(err)
		return map[string][]string{}, map[string] Alias{}
	}
	var esAliases []Alias
	json.Unmarshal(resp.Body(), &esAliases)
	var res = map[string][]string{}
	var indexToAliasMap = map[string]Alias{}
	for _, alias := range esAliases {
		res[alias.Index] = append(res[alias.Index], alias.Alias)
		indexToAliasMap[alias.Index] = alias
	}
	return res, indexToAliasMap
}
