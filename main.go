package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"gopkg.in/yaml.v2"
)

type RequestModel struct {
	Host        string
	Port        string
	Path        string
	Method      string
	Concurrency int
	Total       int
	Message     string
}

type Request struct {
	Url    string
	Method string
	Data   string
}

var requestChannel chan Request
var waitChannel chan int
var totalNum = 0

func main() {
	data, err := ioutil.ReadFile("./request.yaml")
	if err != nil {
		panic(err)
	}

	requestModel := RequestModel{}
	err = yaml.Unmarshal(data, &requestModel)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(requestModel)
	}

	requestChannel = make(chan Request, requestModel.Concurrency)

	go func() {
		request := getRequestFormModel(requestModel)
		for i := 0; i < requestModel.Total; i++ {
			// fmt.Println("add to channel")
			requestChannel <- request
		}
	}()

	for i := 0; i < requestModel.Concurrency; i++ {
		go handleRequestWorker(i)
	}

	<-waitChannel
}

func handleRequestWorker(channelIndex int) {
	client := http.Client{}
	for i := 0; ; i++ {
		request := <-requestChannel

		var httpRequest *http.Request
		var err error

		if request.Method == "post" {
			httpRequest, err = http.NewRequest("POST", request.Url, bytes.NewBuffer([]byte(request.Data)))
			httpRequest.Header.Add("Content-Type", "application/json")
		} else {
			url := request.Url + "?" + request.Data
			httpRequest, err = http.NewRequest("GET", url, nil)
		}

		if err != nil {
			panic(err)
		}

		fmt.Println("start sending request num:" + strconv.Itoa(i) + " for channel:" + strconv.Itoa(channelIndex))
		response, err := client.Do(httpRequest)
		fmt.Println("get response for request num:" + strconv.Itoa(i) + " for channel:" + strconv.Itoa(channelIndex))
		if err != nil {
			panic(err)
		}

		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(body))
		// fmt.Println(response.StatusCode)
		fmt.Println("total num:" + strconv.Itoa(totalNum))
		totalNum++
	}
}

func getRequestFormModel(model RequestModel) Request {
	var request = Request{}

	request.Url = model.Host + ":" + model.Port + model.Path
	request.Data = model.Message
	request.Method = model.Method

	return request
}
