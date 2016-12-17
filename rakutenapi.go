/**
 * Copyright (C) 2016 tech0522.tk
 */
package main

import (
	"encoding/json"
	"errors"
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
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New("Bad status code: code=" + resp.Status + " body=" + string(body))
	}
	var allCategory RecipeAllCategory
	if err := json.NewDecoder(resp.Body).Decode(&allCategory); err != nil {
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
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New("Bad status code: code=" + resp.Status + " body=" + string(body))
	}
	var ranking RecipeRanking
	if err := json.NewDecoder(resp.Body).Decode(&ranking); err != nil {
		return nil, err
	}
	return &ranking, nil
}
