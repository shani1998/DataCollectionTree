package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Metric struct {
	WebReq    int `json:"web_req"`
	TimeSpent int `json:"time_spent"`
}

type Dimension struct {
	Country string `json:"country"`
	Device  string `json:"device"`
}

//using hash map to construct nary tree!
var Tree = make(map[string]map[string]*Metric) //countries---->devices---->metrics
var TotalWebReq int
var TotalTimeSpent int

//------------------------------Fetch metrics details by given country--------------
func getTotalDataByCountry(country string) (int, int) {
	var totalWebReq, totalTimeSpent = 0, 0
	for _, metric := range Tree[country] {
		totalTimeSpent += metric.TimeSpent
		totalWebReq += metric.WebReq
	}
	log.Printf("Metrics data of country %s: totalWebReq:%v,totaTimeSpent:%v", country, TotalWebReq, totalTimeSpent)
	return totalWebReq, totalTimeSpent
}

//------------------------------Query data from tree--------------------------------
func getDataFromTree(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var dataNode map[string][]map[string]string
	var dim Dimension

	// get the body of our POST request
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("unable to read body, Reason:", err)
		return
	}
	//convert body into our desire struct
	err = json.Unmarshal(reqBody, &dataNode)
	if err != nil {
		log.Println("unable to unmarshal body, Reason:", err)
		return
	}
	log.Println(dataNode)

	//check whether dim present or not
	if _, isDimPresent := dataNode["dim"]; !isDimPresent {
		writeResponse(w, "please check input format, must contains key `dim`")
		return
	}

	//fill dimension struct
	for _, d := range dataNode["dim"] {
		if d["key"] == "device" {
			dim.Device = d["val"]
		}
		if d["key"] == "country" {
			dim.Country = d["val"]
		}
	}

	//check whether country, which record we want is already there in tree or not
	if _, isCountryPresent := Tree[dim.Country]; !isCountryPresent {
		msg := fmt.Sprintf("Record not found for country %s", dim.Country)
		writeResponse(w, msg)
		return
	}

	//get sum of metrics generated by all devices in given country
	webReq, timeSpent := getTotalDataByCountry(dim.Country)
	var webReqData = map[string]string{"key": "webreq", "val": strconv.Itoa(webReq)}
	var timeSpentData = map[string]string{"key": "timespent", "val": strconv.Itoa(timeSpent)}
	dataNode["metrics"] = []map[string]string{webReqData, timeSpentData}

	//convert desire result into json and send it as response
	responseBody, err := json.Marshal(dataNode)
	if err != nil {
		log.Println("unable to make response body, Reason: ", err)
	}
	_, err = w.Write(responseBody)
	if err != nil {
		log.Println("error occurred while writing response, Reason: ", err)
	}

}

//------------------------------Insert data into tree-------------------------------
func insertHelper(dimensions, metrics []map[string]interface{}) error {
	var dim Dimension
	var metric Metric

	// check whether dimension array consist of country & devices
	if len(dimensions) != 2 {
		return fmt.Errorf("dimension lenght must be of 2, eg: [{\n\t\t“key”: “device”,\n\t\t“val”: “mobile”\n},\n{\n\t“key”: “country”,\n\t“val”: “IN”\n}]\n ")
	}

	//fill dimension struct
	for _, d := range dimensions {
		if d["key"] == "device" {
			dim.Device = d["val"].(string)
		}
		if d["key"] == "country" {
			dim.Country = d["val"].(string)
		}
	}

	//fill metric struct
	for _, m := range metrics {
		if m["key"] == "webreq" {
			metric.WebReq = int(m["val"].(float64))
		}
		if m["key"] == "timespent" {
			metric.TimeSpent = int(m["val"].(float64))
		}
	}

	//insert country with initial metrics data into tree if not present.
	if _, isCountryPresent := Tree[dim.Country]; !isCountryPresent {
		deviceMetric := make(map[string]*Metric)
		deviceMetric[dim.Device] = &metric
		Tree[dim.Country] = deviceMetric
		log.Printf("Country %s added succesfully in the tree! %v\n", dim.Country, Tree)
		//update totalWebReq and totalTimeSpent
		TotalWebReq += metric.WebReq
		TotalTimeSpent += metric.TimeSpent
		log.Println("total timeWebReq and timeSpent on the earth are:", TotalWebReq, TotalTimeSpent)
		return nil
	}

	//if country node already there in tree
	//check if request device metric data available or not in given country, if not then add new node to child of corresponding country.
	if _, isDevicePresent := Tree[dim.Country][dim.Device]; !isDevicePresent {
		Tree[dim.Country][dim.Device] = &metric
		log.Printf("device %s succesfully added as child of country %s. %v\n", dim.Device, dim.Country, Tree)
		//update totalWebReq and totalTimeSpent
		TotalWebReq += metric.WebReq
		TotalTimeSpent += metric.TimeSpent
		log.Println("total timeWebReq and timeSpent on the earth are:", TotalWebReq, TotalTimeSpent)
		return nil
	}

	//check whether dimension is same, ie, device and country data already present then add to existing data.
	if _, isDevicePresent := Tree[dim.Country][dim.Device]; isDevicePresent {
		//add metrics to existing metric data
		Tree[dim.Country][dim.Device].WebReq += metric.WebReq
		Tree[dim.Country][dim.Device].TimeSpent += metric.TimeSpent
		log.Printf("metrics of country %s and device %s has been added successfully! \n", dim.Device, dim.Country)
		//update totalWebReq and totalTimeSpent
		TotalWebReq += metric.WebReq
		TotalTimeSpent += metric.TimeSpent
		log.Println("total timeWebReq and timeSpent on the earth are:", TotalWebReq, TotalTimeSpent)
	}
	return nil

}

//------------------------------Handle insert api call------------------------------
func insertDataIntoTree(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var dataNode map[string][]map[string]interface{}

	// get the body of our POST request
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("unable to read body, Reason:", err)
		return
	}

	//convert body into our desire struct
	err = json.Unmarshal(reqBody, &dataNode)
	if err != nil {
		log.Println("unable to unmarshal body, Reason:", err)
		return
	}
	log.Println(dataNode)

	//validate and process input data
	//check whether dim present or not
	if _, isDimPresent := dataNode["dim"]; !isDimPresent {
		writeResponse(w, "please check input format, must contains key `dim`")
		return
	}
	//check whether metrics present or not
	if _, isDimPresent := dataNode["metrics"]; !isDimPresent {
		writeResponse(w, "please check input format, must contains key `metrics`")
		return
	}
	//construct nary-tree using hash data structure
	_ = insertHelper(dataNode["dim"], dataNode["metrics"])

}

//---------------------write response back to api server--------------------------
func writeResponse(w http.ResponseWriter, msg string) {
	_, err := w.Write([]byte(msg))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("unable to write response, Reason:", err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	mux := httprouter.New()
	mux.POST("/v1/insert", insertDataIntoTree)
	mux.GET("/v1/query", getDataFromTree)

	log.Fatalln(http.ListenAndServe(":8080", mux))

}
