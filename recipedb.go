/**
 * Copyright 2017 recipe-linebot
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
)

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
