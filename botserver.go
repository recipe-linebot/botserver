/**
 * Copyright (C) 2016 tech0522.tk
 */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"unicode/utf8"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

const RecipeOpenButtonLabel = "レシピを開く"
const PagingButtonLabel = "さらに読み込む"
const PagingCarouselTitle = "　"
const PagingCarouselText = "「%s」のレシピが他に %d 件あります。"
const PagingCarouselImageURL = "https://minagawa-sho.github.io/recipe-linebot/image/paging_%02d.jpg"
const NumPagingCarouselImages = 4
const NotFoundReplyStickerPackageID = "2"
const NotFoundReplyStickerID = "38"
const MaxRecipesToReply = 5
const MaxRecipeDescLength = 60
const RecipeDescTailIfTooLong = "..."
const RecipeCarouselAltTextTailing = "..."

type RecipeDBSearchQuery struct {
	From  int `json:"from"`
	Size  int `json:"size"`
	Query struct {
		MultiMatch struct {
			Query  string   `json:"query"`
			Type   string   `json:"type"`
			Fields []string `json:"fields"`
		} `json:"multi_match"`
	} `json:"query"`
}

type RecipeDBSearchResult struct {
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			Source struct {
				Title    string `json:"title"`
				Desc     string `json:"description"`
				ImageUrl string `json:"image_url"`
				Url      string `json:"detail_url"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type PagingPostbackData struct {
	From     int    `json:"from"`
	RawQuery string `json:"rawQuery"`
}

func searchRecipes(rawQuery string, from, size int, config *RecipeLinebotConfig) (result RecipeDBSearchResult) {
	// Build the search query
	var query RecipeDBSearchQuery
	query.From = from
	query.Size = size
	query.Query.MultiMatch.Query = rawQuery
	query.Query.MultiMatch.Type = "cross_fields"
	query.Query.MultiMatch.Fields = []string{"materials.keyword^100", "materials^5", "title", "description"}
	queryAsJson, err := json.Marshal(query)
	if err != nil {
		log.Fatal(err)
	}

	// Post the search request
	apiUrl := url.URL{Scheme: "http", Host: config.RecipeDB.Host,
		Path: path.Join(config.RecipeDB.Index, config.RecipeDB.RecipeDoctype, "_search")}
	resp, err := http.Post(apiUrl.String(), "application/json", bytes.NewReader(queryAsJson))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Interpret the response as search result
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatal("Bad status code: code=" + resp.Status + " body=" + string(body))
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}
	return result
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

func newRecipesMessage(result *RecipeDBSearchResult, rawQuery string, from int) *linebot.TemplateMessage {
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
		pbData.From = from + MaxRecipesToReply - 1
		pbDataAsJson, err := json.Marshal(pbData)
		if err != nil {
			log.Fatal(err)
		}
		col := linebot.NewCarouselColumn(
			fmt.Sprintf(PagingCarouselImageURL, (from / (MaxRecipesToReply - 1) % NumPagingCarouselImages)),
			PagingCarouselTitle,
			fmt.Sprintf(PagingCarouselText, rawQuery, result.Hits.Total-pbData.From),
			linebot.NewPostbackTemplateAction(PagingButtonLabel, string(pbDataAsJson), ""))
		cols = append(cols, col)
	}

	// Build as carousel message
	altText := result.Hits.Hits[0].Source.Title + RecipeCarouselAltTextTailing
	tmpl := linebot.NewCarouselTemplate(cols...)
	return linebot.NewTemplateMessage(altText, tmpl)
}

func suggestRecipes(bot *linebot.Client, event *linebot.Event, rawQuery string, from int, config *RecipeLinebotConfig) {
	log.Printf("search: query=%s\n", rawQuery)
	result := searchRecipes(rawQuery, from, MaxRecipesToReply, config)
	var replyMsg linebot.Message
	if result.Hits.Total == 0 {
		replyMsg = linebot.NewStickerMessage(NotFoundReplyStickerPackageID, NotFoundReplyStickerID)
	} else {
		replyMsg = newRecipesMessage(&result, rawQuery, from)
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
					suggestRecipes(bot, event, recvMsg.Text, 0, config)
				}
			case linebot.EventTypePostback:
				var pbData PagingPostbackData
				json.Unmarshal([]byte(event.Postback.Data), &pbData)
				suggestRecipes(bot, event, pbData.RawQuery, pbData.From, config)
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
