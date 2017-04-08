#!/bin/bash

TMPDIR=/var/tmp/recipe-linebot-deploy
REPOSITORY_URL=https://github.com/tech0522/recipe-linebot
BRANCH=master
BIN_FILENAME=recipe-linebot
MAPPING_FILENAME=mapping-template.json
SERVICE_PATH=/opt/recipe-linebot
SERVICE_NAME=recipe-linebot

rm -rf $TMPDIR || exit 1
mkdir -p $TMPDIR || exit 1
cd $TMPDIR || exit 1
git clone $REPOSITORY_URL $BIN_FILENAME || exit 1
cd $BIN_FILENAME
git checkout $BRANCH || exit 1
go build || exit 1
mv -f $BIN_FILENAME $SERVICE_PATH/bin/ || exit 1
mv -f $MAPPING_FILENAME $SERVICE_PATH/share/ || exit 1
echo "Succeeded."
