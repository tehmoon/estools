#!/bin/sh

mkdir build
rm build/*
cd src

go get ./...
go build -o ../build/esalertd . || exit 2
