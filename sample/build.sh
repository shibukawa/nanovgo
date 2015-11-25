#!/bin/sh
rm -rf assets
mkdir assets
cp -r images/* assets/
cp ./Roboto-Bold.ttf assets
cp ./Roboto-Regular.ttf assets
cp ./entypo.ttf assets
go-bindata -tags "js" -pkg "demo" -o demo/bindata.go assets
