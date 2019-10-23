#!/bin/bash

APIVERSION=0.x.x
VERSION=$(cat VERSION);
COMMIT=$(git rev-parse --short HEAD);

WRITTENVERSION=$APIVERSION'-'$VERSION'-'$COMMIT

git diff-index --quiet HEAD --

if [[ $? != 0 ]]
then
  echo "There is uncommitted code, commit first, and build again."
  exit 1
fi

sed "s/versionplaceholder/"$WRITTENVERSION"/g" version.template > ./package/version.go
sed -i "s/versionplaceholder/"$WRITTENVERSION"/g" ./module.toml

mkdir bungkus
go build -buildmode=plugin -ldflags="-s -w" -o bungkus/terminal.so
cp -Rvf LICENSE CHANGELOG  module.toml schema bungkus
mv bungkus terminal
tar zcvvf terminal-$VERSION-$COMMIT.tar.gz terminal
rm -Rvf terminal
