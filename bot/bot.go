/**
 * Copyright (C) 2016 tech0522.tk
 */
package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

// Hide the private setting values
const ListenAddr = "XXX.XXX.XXX.XXX:XXX"
const CertFile = "/path/to/fullchain.pem"
const KeyFile = "/path/to/privkey.pem"
const ChSecret = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
const ChToken = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
const ButtonLabel = "レシピを開く"
const ElasticDBHost = "XXX.XXX.XXX.XXX:XXX"
const ElasticDBIndex = "recipe-linebot"
const ElasticDBRecipeDocType = "recipe"

func replyRecipeCarousel(bot *linebot.Client, replyToken string, phrase string) {
	apiUrl := url.URL{Scheme: "http", Host: ElasticDBHost, Path: path.Join(ElasticDBIndex, ElasticDBRecipeDocType, "_search")}
	resp, err := http.Post(apiUrl.String(), "application/json", bytes.NewBuffer([]byte(`{"size": 5, "query": {"multi_match": {"query": "`+phrase+`", "type": "phrase", "fields": ["materials", "title", "description"]}}}`)))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatal("Bad status code: code=" + resp.Status + " body=" + string(body))
	}
	var result struct {
		Hits struct {
			Total int `json:"total"`
			Hits  []struct {
				Source struct {
					Title       string `json:"title"`
					Description string `json:"description"`
					ImageUrl    string `json:"image_url"`
					Url         string `json:"detail_url"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	if result.Hits.Total == 0 {
		return
	}
	var cols []*linebot.CarouselColumn
	for _, hit := range result.Hits.Hits {
		desc := hit.Source.Description
		if len(desc) > 60 {
			desc = desc[0:57] + "..."
		}
		cols = append(cols, linebot.NewCarouselColumn(hit.Source.ImageUrl, hit.Source.Title, desc,
			linebot.NewURITemplateAction(ButtonLabel, hit.Source.Url)))
	}
	tmpl := linebot.NewCarouselTemplate(cols...)
	replyMsg := linebot.NewTemplateMessage(result.Hits.Hits[0].Source.Title+",...", tmpl)
	if _, err = bot.ReplyMessage(replyToken, replyMsg).Do(); err != nil {
		log.Fatal(err)
	}
}

func onMessageEvent(bot *linebot.Client, event *linebot.Event) {
	resp, err := bot.GetProfile((*event.Source).UserID).Do()
	if err != nil {
		log.Fatal(err)
	}
	switch recvMsg := event.Message.(type) {
	case *linebot.TextMessage:
		log.Printf("%s: dispName=%s, text=%s\n", event.Timestamp.String(), resp.DisplayName, recvMsg.Text)
		replyRecipeCarousel(bot, (*event).ReplyToken, recvMsg.Text)
	default:
		log.Printf("%s: dispName=%s\n", event.Timestamp.String(), resp.DisplayName)
	}
}

func main() {
	handler, err := httphandler.New(ChSecret, ChToken)
	if err != nil {
		log.Fatal(err)
	}
	handler.HandleEvents(func(events []*linebot.Event, req *http.Request) {
		bot, err := handler.NewClient()
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				onMessageEvent(bot, event)
			} else {
				log.Printf("%s\n", event.Timestamp.String())
			}
		}
	})
	http.Handle("/recipe-linebot/callback", handler)
	if err := http.ListenAndServeTLS(ListenAddr, CertFile, KeyFile, nil); err != nil {
		log.Fatal(err)
	}
}
