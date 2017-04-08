/**
 * Copyright 2016 tech0522
 */
package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

type PullingProgress struct {
	AllCategories     RecipeAllCategory
	LargeCategoryIdx  int
	MediumCategoryIdx int
	SmallCategoryIdx  int
}

func restorePullingProgress(restorePath string, progress *PullingProgress) error {
	restoreFile, err := os.Open(restorePath)
	if err != nil {
		return err
	}
	defer restoreFile.Close()
	decoder := gob.NewDecoder(bufio.NewReader(restoreFile))
	return decoder.Decode(&progress)
}

func storePullingProgress(progress *PullingProgress, storePath string) error {
	storeFile, err := os.OpenFile(storePath, os.O_WRONLY+os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer storeFile.Close()
	encoder := gob.NewEncoder(bufio.NewWriter(storeFile))
	return encoder.Encode(&progress)
}

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
			resp, err := http.Post(apiUrl.String(), "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode/100 != 2 {
				bodyAsString := "(failed to read)"
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					bodyAsString = string(body)
				}
				log.Fatalf("Bad status code: url=%v, code=%v, body=%v", apiUrl.String(), resp.Status, bodyAsString)
			}
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
		resp, err := http.Post(apiUrl.String(), "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			bodyAsString := "(failed to read)"
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				bodyAsString = string(body)
			}
			log.Fatalf("Bad status code: url=%v, code=%v, body=%v", apiUrl.String(), resp.Status, bodyAsString)
		}
	}
	return nil
}

func pullRecipes(config *RecipeLinebotConfig) {
	log.Print("start pull batch")

	// Restore the progress up to the previous working
	var progress PullingProgress
	restored := true
	err := restorePullingProgress(config.PullBatch.ProgressFilePath, &progress)
	if err != nil {
		if os.IsNotExist(err) {
			restored = false
		} else {
			log.Fatal(err)
		}
	}

	if !restored {
		allCategories, err := FetchRecipeCategories(RecipeCategoryAll, config.RakutenAPI.AppId)
		if err != nil {
			log.Fatal(err)
		}
		progress.AllCategories = *allCategories
	}
	for idx, category := range progress.AllCategories.By.Large {
		if idx <= progress.LargeCategoryIdx {
			continue
		}
		err = pullRecipesOnCategory(category.Id, category.Name, config)
		if err != nil {
			log.Fatal(err)
		}
		progress.LargeCategoryIdx = idx
		err = storePullingProgress(&progress, config.PullBatch.ProgressFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}
	for idx, category := range progress.AllCategories.By.Medium {
		if idx <= progress.MediumCategoryIdx {
			continue
		}
		categoryUrl, err := url.Parse(category.Url)
		if err != nil {
			log.Fatal(err)
		}
		categoryId := path.Base(categoryUrl.Path)
		err = pullRecipesOnCategory(categoryId, category.Name, config)
		if err != nil {
			log.Fatal(err)
		}
		progress.MediumCategoryIdx = idx
		err = storePullingProgress(&progress, config.PullBatch.ProgressFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}
	for idx, category := range progress.AllCategories.By.Small {
		if idx <= progress.SmallCategoryIdx {
			continue
		}
		categoryUrl, err := url.Parse(category.Url)
		if err != nil {
			log.Fatal(err)
		}
		categoryId := path.Base(categoryUrl.Path)
		err = pullRecipesOnCategory(categoryId, category.Name, config)
		if err != nil {
			log.Fatal(err)
		}
		progress.SmallCategoryIdx = idx
		err = storePullingProgress(&progress, config.PullBatch.ProgressFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = os.Remove(config.PullBatch.ProgressFilePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("pull batch finished")
}
