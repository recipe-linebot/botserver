/**
 * Copyright (C) 2016 tech0522.tk
 */
package main

import (
	"flag"
	"io/ioutil"
	"log"
)

// Hide the private setting values
const ListenAddr = "XXX.XXX.XXX.XXX:XXX"
const ListenPath = "/XXX/XXX"
const CertFile = "/path/to/fullchain.pem"
const KeyFile = "/path/to/privkey.pem"
const ChSecret = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
const ChToken = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
const ElasticDBHost = "XXX.XXX.XXX.XXX:XXX"
const ElasticDBIndex = "recipe-linebot"
const ElasticDBRecipeDocType = "recipe"
const ElasticDBRankingDocType = "ranking"
const RakutenAppId = "XXXXXXXXXXXXXXXXX"
const APICallInterval = 1 * time.Second

type RecipeLinebotConfig struct {
	BotServer struct {
		ListenAddr   string `json: "listen_addr"`
		APIPath      string `json: "api_path"`
		CertFilePath string `json: "cert_file_path"`
		KeyFilePath  string `json: "key_file_path"`
		ChSecret     string `json: "channel_secret"`
		ChToken      string `json: "channel_token"`
	} `json: "bot_server"`
	RecipeDB struct {
		Host           string `json: "host"`
		Index          string `json: "index"`
		RecipeDoctype  string `json: "recipe_doctype"`
		RankingDoctype string `json: "ranking_doctype"`
	} `json: "recipe_db"`
	RakutenAPI struct {
		AppId        string `json: "app_id"`
		CallInterval int    `json: "call_interval"`
	} `json: "rakuten_api"`
}

func main() {
	mode := flag.String("m", "", "running mode ('botserver', 'pullbatch')")
	confpath := flag.String("c", "", "config file path")
	flag.Parse()
	conffile, err := os.Open(confpath)
	if err != nil {
		log.Fatal(err)
	}
	defer conffile.Close()
	var conf RecipeLinebotConfig
	err = json.Unmarchal(confdata, &conf)
	if err != nil {
		log.Fatal(err)
	}
	if *mode == "botserver" {
		serveAsBot(conf)
	} else if *mode == "pullbatch" {
		pullRecipes(conf)
	} else {
		log.Printf("the specified mode is not implemented: %s", *mode)
	}
}
