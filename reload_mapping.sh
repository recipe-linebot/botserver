#!/bin/bash

RECIPEDB_HOST=127.0.0.1
RECIPEDB_PORT=9200
MAPPING_PATH=/opt/recipe-linebot/share/mapping-template.json

curl -XPUT "$RECIPEDB_HOST:$RECIPEDB_PORT/_template/recipe-linebot?pretty" -d @$MAPPING_PATH
