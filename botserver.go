/**
 * Copyright 2017 recipe-linebot
 */
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

const RecipeOpenButtonLabel = "レシピを開く"
const PagingButtonLabel = "さらに読み込む"
const PagingCarouselTitle = "　"
const PagingCarouselText = "「%s」のレシピが他に %d 件あります。"
const PagingCarouselImageURL = "https://recipe-linebot.github.io/botserver/image/paging_%02d.jpg"
const NumPagingCarouselImages = 4
const NotFoundReplyStickerPackageID = "2"
const NotFoundReplyStickerID = "38"
const MaxRecipesToReply = 5
const MaxRecipeDescLength = 60
const RecipeDescTailIfTooLong = "..."
const RecipeCarouselAltTextTailing = "..."

type PagingPostbackData struct {
	Offset   int    `json:"offset"`
	RawQuery string `json:"rawQuery"`
}

func roundRecipeDescription(desc string) string {
	if len(desc) <= MaxRecipeDescLength {
		return desc
	}
	descEnd := MaxRecipeDescLength - len(RecipeDescTailIfTooLong)
	for !utf8.RuneStart(desc[descEnd]) {
		descEnd--
	}
	return desc[0:descEnd] + RecipeDescTailIfTooLong
}

func newRecipesMessage(result *RecipeDBSearchResult, rawQuery string, offset int) *linebot.TemplateMessage {
	// Build the carousel columns
	var cols []*linebot.CarouselColumn
	needsPaging := false
	for i, hit := range result.Hits.Hits {
		if i == MaxRecipesToReply-1 {
			needsPaging = true
			break
		}
		col := linebot.NewCarouselColumn(hit.Source.ImageUrl, hit.Source.Title,
			roundRecipeDescription(hit.Source.Desc),
			linebot.NewURITemplateAction(RecipeOpenButtonLabel, hit.Source.Url))
		cols = append(cols, col)
	}

	// If next page exists, append button to fetch next result
	if needsPaging {
		var pbData PagingPostbackData
		pbData.RawQuery = rawQuery
		pbData.Offset = offset + MaxRecipesToReply - 1
		pbDataAsJson, err := json.Marshal(pbData)
		if err != nil {
			log.Fatal(err)
		}
		col := linebot.NewCarouselColumn(
			fmt.Sprintf(PagingCarouselImageURL, (offset/(MaxRecipesToReply-1)%NumPagingCarouselImages)),
			PagingCarouselTitle,
			fmt.Sprintf(PagingCarouselText, rawQuery, result.Hits.Total-pbData.Offset),
			linebot.NewPostbackTemplateAction(PagingButtonLabel, string(pbDataAsJson), ""))
		cols = append(cols, col)
	}

	// Build as carousel message
	altText := result.Hits.Hits[0].Source.Title + RecipeCarouselAltTextTailing
	tmpl := linebot.NewCarouselTemplate(cols...)
	return linebot.NewTemplateMessage(altText, tmpl)
}

func replyMessage(bot *linebot.Client, event *linebot.Event, rawQuery string, offset int, config *RecipeLinebotConfig) {
	log.Printf("search: query=%s\n", rawQuery)
	result := searchRecipes(rawQuery, offset, MaxRecipesToReply, config)
	var replyMsg linebot.Message
	if result.Hits.Total == 0 {
		replyMsg = linebot.NewStickerMessage(NotFoundReplyStickerPackageID, NotFoundReplyStickerID)
	} else {
		replyMsg = newRecipesMessage(&result, rawQuery, offset)
	}
	if _, err := bot.ReplyMessage(event.ReplyToken, replyMsg).Do(); err != nil {
		log.Fatal(err)
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
			// Log an event receivement
			dispName := "(unknown)"
			profile, err := bot.GetProfile(event.Source.UserID).Do()
			if err == nil {
				dispName = profile.DisplayName
			} else {
				log.Print(err)
			}
			log.Printf("eventType: %s, userDispName\n", event.Type, dispName)

			// Suggest recipes according to user request
			switch event.Type {
			case linebot.EventTypeMessage:
				switch recvMsg := event.Message.(type) {
				case *linebot.TextMessage:
					replyMessage(bot, event, recvMsg.Text, 0, config)
				}
			case linebot.EventTypePostback:
				var pbData PagingPostbackData
				json.Unmarshal([]byte(event.Postback.Data), &pbData)
				replyMessage(bot, event, pbData.RawQuery, pbData.Offset, config)
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
