/**
 * Copyright (C) 2016, 2017 tech0522.tk
 */
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

const APIEndpoint = "https://app.rakuten.co.jp/services/api/"
const RecipeCategoryListAPIPath = "Recipe/CategoryList/"
const RecipeCategoryListAPILatestVersion = "20121121"
const RecipeCategoryRankingAPIPath = "Recipe/CategoryRanking/"
const RecipeCategoryRankingAPILatestVersion = "20121121"

type RecipeParentCategory struct {
	Id   string `json:"categoryId"`
	Name string `json:"categoryName"`
	Url  string `json:"categoryUrl"`
}

type RecipeChildCategory struct {
	Id       int    `json:"categoryId"`
	Name     string `json:"categoryName"`
	Url      string `json:"categoryUrl"`
	ParentId string `json:"parentCategoryId"`
}

type RecipeAllCategoryByLevel struct {
	Large  []RecipeParentCategory `json:"large"`
	Medium []RecipeChildCategory  `json:"medium"`
	Small  []RecipeChildCategory  `json:"small"`
}

type RecipeAllCategory struct {
	By RecipeAllCategoryByLevel `json:"result"`
}

type RecipeCategoryType string

const (
	RecipeCategoryLarge  RecipeCategoryType = "large"
	RecipeCategoryMedium RecipeCategoryType = "medium"
	RecipeCategorySmall  RecipeCategoryType = "small"
	RecipeCategoryAll    RecipeCategoryType = ""
)

func FetchRecipeCategories(categoryType RecipeCategoryType, appId string) (*RecipeAllCategory, error) {
	apiUrl, err := url.Parse(APIEndpoint)
	if err != nil {
		return nil, err
	}
	apiUrl.Path = path.Join(apiUrl.Path, RecipeCategoryListAPIPath, RecipeCategoryListAPILatestVersion)
	apiUrl.RawQuery = "applicationId=" + appId
	if categoryType != RecipeCategoryAll {
		apiUrl.RawQuery += "&categoryType=" + string(categoryType)
	}
	resp, err := http.Get(apiUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad status code: url=%v, code=%v, body=%v", apiUrl.String(), resp.Status, string(body))
	}
	var allCategory RecipeAllCategory
	err = json.Unmarshal(body, &allCategory)
	if err != nil {
		return nil, err
	}
	return &allCategory, nil
}

type RecipeSummary struct {
	Id             int      `json:"recipeId"`
	Title          string   `json:"recipeTitle"`
	Url            string   `json:"recipeUrl"`
	LargeImageUrl  string   `json:"foodImageUrl"`
	MediumImageUrl string   `json:"mediumImageUrl"`
	SmallImageUrl  string   `json:"smallImageUrl"`
	PickUp         int      `json:"pickup"`
	Shop           int      `json:"shop"`
	Nickname       string   `json:"nickname"`
	Description    string   `json:"recipeDescription"`
	Materials      []string `json:"recipeMaterial"`
	Indication     string   `json:"recipeIndication"`
	Cost           string   `json:"recipeCost"`
	PublishDay     string   `json:"recipePublishday"`
	Rank           string   `json:"rank"`
}

type RecipeRanking struct {
	Recipes []RecipeSummary `json:"result"`
}

func FetchRecipeRanking(categoryId string, appId string) (*RecipeRanking, error) {
	apiUrl, err := url.Parse(APIEndpoint)
	if err != nil {
		return nil, err
	}
	apiUrl.Path = path.Join(apiUrl.Path, RecipeCategoryRankingAPIPath, RecipeCategoryRankingAPILatestVersion)
	apiUrl.RawQuery = "applicationId=" + appId
	if categoryId != "" {
		apiUrl.RawQuery += "&categoryId=" + categoryId
	}
	resp, err := http.Get(apiUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bad status code: url=%v, code=%v, body=%v", apiUrl.String(), resp.Status, string(body))
	}
	var ranking RecipeRanking
	err = json.Unmarshal(body, &ranking)
	if err != nil {
		return nil, err
	}
	return &ranking, nil
}
