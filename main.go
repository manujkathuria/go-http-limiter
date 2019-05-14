package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 5,
}

func main() {
	file, err := os.Open("activate_after_feb_with_fixed_speed.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var cmxTuple CMXTuple
	requests := make(chan CMXTuple)
	defer close(requests)

	// 100 TPS : 1 request per 10 ms
	limiter := time.Tick(10 * time.Millisecond)
	t1 := time.Now()
	for scanner.Scan() {
		<-limiter
		cmxTuple.build(scanner.Text())
		resp, err := cmxTuple.PrepareAndPost()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(resp.ResultCode)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	t2 := time.Now()
	diff := t2.Sub(t1)
	fmt.Println(diff)
}

type CMXTuple struct {
	MSISDN        string
	CampaignCode  string
	OfferCode     string
	RuleId        string
	Param1        string
	Param2        string
	Param3        string
	Param4        string
	Param5        string
	PriceplanCode string
	ExpDate       string
}

func (cmx *CMXTuple) build(msisdn string) {
	cmx.MSISDN = msisdn
	cmx.CampaignCode = "C000000572"
	cmx.OfferCode = "RO00002488"
	cmx.RuleId = "FixSpeedPack"
	cmx.Param1 = "0"
	cmx.Param2 = "0"
	cmx.Param3 = "0"
	cmx.Param4 = "THB"
	cmx.Param5 = "MUL"
	cmx.PriceplanCode = "IPP_000O0_00_NOTALLOWSTD"
	cmx.ExpDate = "2020-01-01 23-59-59"

}

func (cmx *CMXTuple) PrepareAndPost() (Response, error) {
	b, _ := json.Marshal(cmx)
	body := bytes.NewBuffer([]byte(b))
	req, err := http.NewRequest("POST", "http://10.50.77.32:8090/BSOMgateway/AddprepaidPack/", body)
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println(err)

	}

	return response, nil
}

type Response struct {
	ResultCode        int `json:"ResultCode"`
	ResultDiscription struct {
		PARAMVALUE string `json:"PARAM_VALUE"`
	} `json:"ResultDiscription"`
}
