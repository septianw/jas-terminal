#!/bin/bash

VERSION=$(cat VERSION);
COMMIT=$(git rev-parse --short HEAD);

git diff-index --quiet HEAD --

if [[ $? != 0 ]]
then
  echo "There is uncommitted code, commit first, and build again."
  exit 1
fi

sed "s/commitplaceholder/"$COMMIT"/g" version.template > ./package/verion.go
sed -i "s/versionplaceholder/"$VERSION"/g" ./package/version.go
mkdir bungkus
go build -buildmode=plugin -ldflags="-s -w" -o bungkus/terminal.so
cp -Rvf LICENSE CHANGELOG  module.toml schema bungkus
mv bungkus terminal
tar zcvvf terminal-$VERSION-$COMMIT.tar.gz terminal
rm -Rvf terminal
