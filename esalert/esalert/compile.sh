#!/bin/sh

mkdir build
rm build/*
cd src

go get ./...
CC=$(which musl-gcc) go build -o ../build/esalert --ldflags '-w -s -linkmode external -extldflags "-static"' . || true 

cd ../build
zip esalert-linux-x86_64.zip esalert
cd -
