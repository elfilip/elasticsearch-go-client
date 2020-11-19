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

type Elastic struct {
	Connect store.Connect
	Indices []Index
	client  *resty.Client
}

type Index struct {
	Name     string `json:"index"`
	DocCount int64  `json:"operation"`
	Size     string `json:"store.size"`
	Health   string `json:"health"`
}

type Es struct {
	Indices []Index
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
	fmt.Println(string(resp.Body()))
	fmt.Println(es)
	return true
}

func (elastic *Elastic) LoadNFirstDocs(count int, index string, query string)  []map[string]interface{}{
	if len(query) == 0 {
		query = "*"
	}
	fmt.Println(elastic.Connect.Url + "/" + index + urlRandomSearch + strconv.Itoa(97) + "&q="+query)
	resp, err := elastic.client.R().Get(elastic.Connect.Url + "/" + index + urlRandomSearch + strconv.Itoa(97) + "&q="+query)
	if err != nil {
		fmt.Println(err)
	}
	var res map[string]interface{}
	json.Unmarshal(resp.Body(), &res)

	var finalres []map[string]interface{}
	for _,val := range res["hits"].(map[string]interface{})["hits"].([]interface{}){
		finalres = append(finalres, val.(map[string]interface{})["_source"].(map[string]interface{}))
	}
	return finalres

}
