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
	"unicode/utf8"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

const ButtonLabel = "レシピを開く"
const NotFoundReplyStickerPackageID = "2"
const NotFoundReplyStickerID = "38"
const MaxRecipesToReply = "5"
const MaxRecipeDescLength = 60
const RecipeDescTailIfTooLong = "..."
const RecipeCarouselAltTextTailing = "..."

type RecipeDBSearchResult struct {
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			Source struct {
				Title    string `json:"title"`
				Desc  	 string `json:"description"`
				ImageUrl string `json:"image_url"`
				Url      string `json:"detail_url"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func replyRecipe(bot *linebot.Client, replyToken string, phrase string, config *RecipeLinebotConfig) {
	apiUrl := url.URL{Scheme: "http", Host: config.RecipeDB.Host,
		Path: path.Join(config.RecipeDB.Index, config.RecipeDB.RecipeDoctype, "_search")}
	resp, err := http.Post(apiUrl.String(), "application/json",
		bytes.NewBuffer([]byte(`{"size": `+MaxRecipesToReply+`, "query": {"multi_match": {"query": "`+phrase+`", "type": "phrase", "fields": ["materials", "title", "description"]}}}`)))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatal("Bad status code: code=" + resp.Status + " body=" + string(body))
	}
	var result RecipeDBSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	var replyMsg linebot.Message
	if result.Hits.Total == 0 {
		replyMsg = linebot.NewStickerMessage(NotFoundReplyStickerPackageID, NotFoundReplyStickerID)
	} else {
		var cols []*linebot.CarouselColumn
		for _, hit := range result.Hits.Hits {
			desc := hit.Source.Desc
			if len(desc) > MaxRecipeDescLength {
				descEnd := MaxRecipeDescLength - len(RecipeDescTailIfTooLong)
				for !utf8.RuneStart(desc[descEnd]) {
					descEnd--
				}
				desc = desc[0:descEnd] + RecipeDescTailIfTooLong
			}
			cols = append(cols, linebot.NewCarouselColumn(hit.Source.ImageUrl, hit.Source.Title, desc,
				linebot.NewURITemplateAction(ButtonLabel, hit.Source.Url)))
		}
		altText := result.Hits.Hits[0].Source.Title + RecipeCarouselAltTextTailing
		tmpl := linebot.NewCarouselTemplate(cols...)
		replyMsg = linebot.NewTemplateMessage(altText, tmpl)
	}
	if _, err = bot.ReplyMessage(replyToken, replyMsg).Do(); err != nil {
		log.Fatal(err)
	}
}

func onMessageEvent(bot *linebot.Client, event *linebot.Event, config *RecipeLinebotConfig) {
	resp, err := bot.GetProfile((*event.Source).UserID).Do()
	if err != nil {
		log.Fatal(err)
	}
	switch recvMsg := event.Message.(type) {
	case *linebot.TextMessage:
		log.Printf("receive text message: from=%s, text=%s\n", resp.DisplayName, recvMsg.Text)
		replyRecipe(bot, (*event).ReplyToken, recvMsg.Text, config)
	default:
		log.Printf("receive a some kind of event: from=%s\n", resp.DisplayName)
	}
}

func serveAsBot(config *RecipeLinebotConfig) {
	log.Printf("start bot server: addr=%s, path=%s", config.BotServer.ListenAddr, config.BotServer.APIPath)
	handler, err := httphandler.New(config.BotServer.ChSecret, config.BotServer.ChToken)
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
				onMessageEvent(bot, event, config)
			} else {
				log.Printf("%s\n", event.Timestamp.String())
			}
		}
	})
	http.Handle(config.BotServer.APIPath, handler)
	if err := http.ListenAndServeTLS(config.BotServer.ListenAddr, config.BotServer.CertFilePath,
		config.BotServer.KeyFilePath, nil); err != nil {
		log.Fatal(err)
	}
	log.Print("bot server finished")
}
