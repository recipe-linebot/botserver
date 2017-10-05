/**
 * Copyright 2017 recipe-linebot
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
	RecipeDB struct {
		Host           string `json:"host"`
		Index          string `json:"index"`
		RecipeDoctype  string `json:"recipe_doctype"`
		RankingDoctype string `json:"ranking_doctype"`
	} `json:"recipe_db"`
}

func main() {
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
	serveAsBot(&conf)
}
