#!/bin/sh

pushd sample

echo "generate for desktop"

go build

echo "generate for web"

rm -rf assets
mkdir assets
cp -r images/* assets/
cp ./Roboto-Bold.ttf assets
cp ./Roboto-Regular.ttf assets
cp ./entypo.ttf assets
go-bindata -tags "js" -pkg "demo" -o demo/bindata.go assets
gopherjs build -o ./gh-pages/demo/sample.js

popd
