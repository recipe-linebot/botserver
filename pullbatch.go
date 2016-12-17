/**
 * Copyright (C) 2016 tech0522.tk
 */
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/minagawa-sho/recipe-linebot/rakutenapi"
)

// Hide the private setting values
const RakutenAppId = "XXXXXXXXXXXXXXXXX"
const ElasticDBHost = "XXX.XXX.XXX.XXX:XXX"
const ElasticDBIndex = "recipe-linebot"
const ElasticDBRecipeDocType = "recipe"
const ElasticDBRankingDocType = "ranking"
const APICallInterval = 1 * time.Second

type RecipeDocument struct {
	Materials   []string `json:"materials"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ImageUrl    string   `json:"image_url"`
	DetailUrl   string   `json:"detail_url"`
}

type RankingDocument struct {
	Concept string `json:"concept"`
	Recipes []int  `json:"recipes"`
}

func pullRecipesOnCategory(categoryId string, categoryName string) error {
	time.Sleep(APICallInterval)
	ranking, err := rakutenapi.FetchRecipeRanking(categoryId, RakutenAppId)
	if err != nil {
		return err
	}
	if len(ranking.Recipes) == 0 {
		log.Printf("recipe not found: category=%v(%v)", categoryId, categoryName)
	} else {
		var recipes []int
		for _, recipe := range ranking.Recipes {
			log.Printf("post recipe: id=%v, title=%v", recipe.Id, recipe.Title)
			apiUrl := url.URL{Scheme: "http", Host: ElasticDBHost, Path: path.Join(ElasticDBIndex, ElasticDBRecipeDocType, strconv.Itoa(recipe.Id))}
			imageUrl, err := url.Parse(recipe.LargeImageUrl)
			if err != nil {
				log.Fatal(err)
			}
			imageUrl.Scheme = "https"
			document := RecipeDocument{Materials: recipe.Materials, Title: recipe.Title, Description: recipe.Description,
				ImageUrl: imageUrl.String(), DetailUrl: recipe.Url}
			reqBody, nil := json.Marshal(document)
			if err != nil {
				log.Fatal(err)
			}
			http.Post(apiUrl.String(), "application/json", bytes.NewBuffer(reqBody))
			recipes = append(recipes, recipe.Id)
		}
		log.Printf("post ranking: category=%v(%v), recipes=%v", categoryId, categoryName, recipes)
		apiUrl := url.URL{Scheme: "http", Host: ElasticDBHost, Path: path.Join(ElasticDBIndex, ElasticDBRankingDocType, categoryId)}
		document := RankingDocument{Concept: categoryName, Recipes: recipes}
		reqBody, nil := json.Marshal(document)
		if err != nil {
			log.Fatal(err)
		}
		http.Post(apiUrl.String(), "application/json", bytes.NewBuffer(reqBody))
	}
	return nil
}

func pullRecipes() {
	categories, err := rakutenapi.FetchRecipeCategories(rakutenapi.RecipeCategoryAll, RakutenAppId)
	if err != nil {
		log.Print(err)
	}
	for _, category := range categories.By.Large {
		if err := pullRecipesOnCategory(category.Id, category.Name); err != nil {
			log.Print(err)
		}
	}
	for _, category := range categories.By.Medium {
		categoryUrl, err := url.Parse(category.Url)
		if err != nil {
			log.Print(err)
		}
		categoryId := path.Base(categoryUrl.Path)
		if err := pullRecipesOnCategory(categoryId, category.Name); err != nil {
			log.Print(err)
		}
	}
	for _, category := range categories.By.Small {
		categoryUrl, err := url.Parse(category.Url)
		if err != nil {
			log.Print(err)
		}
		categoryId := path.Base(categoryUrl.Path)
		if err := pullRecipesOnCategory(categoryId, category.Name); err != nil {
			log.Print(err)
		}
	}
}
