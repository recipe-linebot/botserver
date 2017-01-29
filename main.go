/**
 * Copyright (C) 2016 tech0522.tk
 */
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

type RecipeLinebotConfig struct {
	BotServer struct {
		ListenAddr   string `json:"listen_addr"`
		APIPath      string `json:"api_path"`
		CertFilePath string `json:"certfile_filepath"`
		KeyFilePath  string `json:"keyfile_filepath"`
		ChSecret     string `json:"channel_secret"`
		ChToken      string `json:"channel_token"`
	} `json:"bot_server"`
	PullBatch struct {
		ProgressFilePath string `json:"progress_filepath"`
	} `json:"pull_batch"`
	RecipeDB struct {
		Host           string `json:"host"`
		Index          string `json:"index"`
		RecipeDoctype  string `json:"recipe_doctype"`
		RankingDoctype string `json:"ranking_doctype"`
	} `json:"recipe_db"`
	RakutenAPI struct {
		AppId        string `json:"app_id"`
		CallInterval int    `json:"call_interval_sec"`
	} `json:"rakuten_api"`
}

func main() {
	mode := flag.String("m", "", "running mode ('botserver', 'pullbatch')")
	confpath := flag.String("c", "", "config file path")
	flag.Parse()
	confdata, err := ioutil.ReadFile(*confpath)
	if err != nil {
		log.Fatal(err)
	}
	var conf RecipeLinebotConfig
	err = json.Unmarshal(confdata, &conf)
	if err != nil {
		log.Fatal(err)
	}
	if *mode == "botserver" {
		serveAsBot(&conf)
	} else if *mode == "pullbatch" {
		pullRecipes(&conf)
	} else {
		log.Printf("the specified mode is not implemented: %s", *mode)
	}
}
