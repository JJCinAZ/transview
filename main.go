package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
)

type Event struct {
	KeyID         string `json:"KeyID"`
	Latitude      string `json:"Latitude"`
	Longitude     string `json:"Longitude"`
	Description   string `json:"description"`
	Headline      string `json:"headline"`
	Jurisdiction  string `json:"jurisdiction"`
	CreatedString string `json:"created"`
	Created       time.Time
	UpdatedString string `json:"updated"`
	Updated       time.Time
}

func main() {
	jsonData, err := readData("https://www.transview.org/Incidents.aspx/GetEvents")
	if err != nil {
		log.Fatal(err)
	}
	events, err := parseData(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	for i := range events {
		spew.Dump(events[i])
	}
}

func readData(url string) ([]byte, error) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)
	client := &http.Client{Timeout: time.Second * 30}
	req, err = http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"theEventType":"#Incidents"}`)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Accept", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func parseData(jsondata []byte) ([]Event, error) {
	var (
		result1 struct {
			EscapedData string `json:"d"`
		}
		result2 struct {
			Wrapper struct {
				Events []Event `json:"Event"`
			} `json:"Events"`
		}
		reTime1 = regexp.MustCompile(`(\d\d:\d\d:\d\d) \d{1,2}:\d{1,2} \w\w (\d\d/\d\d/\d{4})`)
	)
	err := json.Unmarshal(jsondata, &result1)
	if err != nil {
		return nil, err
	}
	r := strings.NewReplacer(`\"`, `"`, `<b>`, ``, `</b>`, ``, `<br/>`, ` `)
	result1.EscapedData = r.Replace(result1.EscapedData)
	err = json.Unmarshal([]byte(result1.EscapedData), &result2)
	if err == nil {
		for i := range result2.Wrapper.Events {
			if m := reTime1.FindStringSubmatch(result2.Wrapper.Events[i].CreatedString); m != nil {
				result2.Wrapper.Events[i].Created, _ = time.ParseInLocation("15:04:05 01/02/2006", m[1]+" "+m[2], time.Local)
			}
			if m := reTime1.FindStringSubmatch(result2.Wrapper.Events[i].UpdatedString); m != nil {
				result2.Wrapper.Events[i].Updated, _ = time.ParseInLocation("15:04:05 01/02/2006", m[1]+" "+m[2], time.Local)
			}
		}
	}
	return result2.Wrapper.Events, nil
}
