package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	SubTitle       string `xml:"title"`
	SubLink        string `xml:"link"`
	SubDescription string `xml:"description"`
}

func retrieveRSS(url string) RSS {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var rss RSS
	err = xml.Unmarshal(data, &rss)
	if err != nil {
		panic(err)
	}

	return rss
}

func getDataPayload(url string) (map[string]string, int) {

	rss := retrieveRSS(url)

	rssPayload := map[string]string{
		"title":       rss.Channel.Title,
		"description": rss.Channel.Description,
		"link":        rss.Channel.Link,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	items := []string{}

	for _, item := range rss.Channel.Items {

		//TODO: Support breaking large RSS file into multiple FCM records
		if count(items) > 2500 {
			break
		}
		items = append(items, "{\"subtitle\":\""+item.SubTitle+"\", \"description\":\""+item.SubDescription+"\", \"link\":\""+item.SubLink+"\"}")
	}

	rssPayload["content"] = fmt.Sprintf("[%s]", strings.Join(items, ","))

	// Ensure datapayload fits inside a single FCM notification message.
	var count int
	for _, value := range rssPayload {
		count += len(value)
	}

	return rssPayload, count
}

func count(items []string) int {
	var count int
	for _, item := range items {
		count += len(item)
	}
	return count
}
