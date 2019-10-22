#!/bin/bash

mkdir bungkus
go build -buildmode=plugin -ldflags="-s -w" -o bungkus/terminal.so
cp -Rvf LICENSE CHANGELOG  module.toml schema bungkus
mv bungkus terminal
tar zcvvf terminal.tar.gz terminal
rm -Rvf terminal
