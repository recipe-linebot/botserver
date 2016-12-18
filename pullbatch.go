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
)

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

func pullRecipesOnCategory(categoryId string, categoryName string, config *RecipeLinebotConfig) error {
	time.Sleep(time.Duration(config.RakutenAPI.CallInterval) * time.Second)
	ranking, err := FetchRecipeRanking(categoryId, config.RakutenAPI.AppId)
	if err != nil {
		return err
	}
	if len(ranking.Recipes) == 0 {
		log.Printf("recipe not found: category=%v(%v)", categoryId, categoryName)
	} else {
		var recipes []int
		for _, recipe := range ranking.Recipes {
			log.Printf("post recipe: id=%v, title=%v", recipe.Id, recipe.Title)
			apiUrl := url.URL{Scheme: "http", Host: config.RecipeDB.Host,
				Path: path.Join(config.RecipeDB.Index, config.RecipeDB.RecipeDoctype, strconv.Itoa(recipe.Id))}
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
		apiUrl := url.URL{Scheme: "http", Host: config.RecipeDB.Host,
			Path: path.Join(config.RecipeDB.Index, config.RecipeDB.RankingDoctype, categoryId)}
		document := RankingDocument{Concept: categoryName, Recipes: recipes}
		reqBody, nil := json.Marshal(document)
		if err != nil {
			log.Fatal(err)
		}
		http.Post(apiUrl.String(), "application/json", bytes.NewBuffer(reqBody))
	}
	return nil
}

func pullRecipes(config *RecipeLinebotConfig) {
	categories, err := FetchRecipeCategories(RecipeCategoryAll, config.RakutenAPI.AppId)
	if err != nil {
		log.Fatal(err)
	}
	for _, category := range categories.By.Large {
		if err := pullRecipesOnCategory(category.Id, category.Name, config); err != nil {
			log.Print(err)
		}
	}
	for _, category := range categories.By.Medium {
		categoryUrl, err := url.Parse(category.Url)
		if err != nil {
			log.Print(err)
		}
		categoryId := path.Base(categoryUrl.Path)
		if err := pullRecipesOnCategory(categoryId, category.Name, config); err != nil {
			log.Print(err)
		}
	}
	for _, category := range categories.By.Small {
		categoryUrl, err := url.Parse(category.Url)
		if err != nil {
			log.Print(err)
		}
		categoryId := path.Base(categoryUrl.Path)
		if err := pullRecipesOnCategory(categoryId, category.Name, config); err != nil {
			log.Print(err)
		}
	}
}
