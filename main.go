/**
 * Copyright (C) 2016 tech0522.tk
 */
package main

import (
	"flag"
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

func main() {
	mode := flag.String("m", "", "the running mode ('botserver', 'pullbatch')")
	flag.Parse()
	if *mode == "botserver" {
		serveAsBot(conf)
	} else if *mode == "pullbatch" {
		pullRecipes(conf)
	} else {
		log.Printf("the specified mode is not implemented: %s", *mode)
	}
}
