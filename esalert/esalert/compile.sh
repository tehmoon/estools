#!/bin/sh

mkdir build
rm build/*
cd src

go get ./...
go build -o ../build/esalert . || exit 2
