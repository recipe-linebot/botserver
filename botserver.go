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

const ButtonLabel = "レシピを開く"
const MaxRecipesToReply = 5
const MaxRecipeDescriptionLength = 60
const RecipeDescriptionTailIfTooLong = "..."
const RecipeCarouselAltTextTailing = "..."

func replyRecipe(bot *linebot.Client, replyToken string, phrase string) {
	apiUrl := url.URL{Scheme: "http", Host: ElasticDBHost, Path: path.Join(ElasticDBIndex, ElasticDBRecipeDocType, "_search")}
	resp, err := http.Post(apiUrl.String(), "application/json", bytes.NewBuffer([]byte(`{"size": `+MaxRecipesToReply+`, "query": {"multi_match": {"query": "`+phrase+`", "type": "phrase", "fields": ["materials", "title", "description"]}}}`)))
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
		if len(desc) > MaxRecipeDescriptionLength {
			offset_to_shorten = MaxRecipeDescriptionLength - len(RecipeDescriptionTailIfTooLong)
			desc = desc[0:offset_to_shorten] + RecipeDescriptionTailIfTooLong
		}
		cols = append(cols, linebot.NewCarouselColumn(hit.Source.ImageUrl, hit.Source.Title, desc,
			linebot.NewURITemplateAction(ButtonLabel, hit.Source.Url)))
	}
	altText = result.Hits.Hits[0].Source.Title + RecipeCarouselAltTextTailing
	tmpl := linebot.NewCarouselTemplate(cols...)
	replyMsg := linebot.NewTemplateMessage(altText, tmpl)
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
		replyRecipe(bot, (*event).ReplyToken, recvMsg.Text)
	default:
		log.Printf("%s: dispName=%s\n", event.Timestamp.String(), resp.DisplayName)
	}
}

func serveAsBot() {
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
	http.Handle(ListenPath, handler)
	if err := http.ListenAndServeTLS(ListenAddr, CertFile, KeyFile, nil); err != nil {
		log.Fatal(err)
	}
}
