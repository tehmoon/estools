#!/bin/sh

mkdir build
rm build/*
cd src

go get ./...
CC=$(which musl-gcc) go build -o ../build/esalertd --ldflags '-w -s -linkmode external -extldflags "-static"' . || true 

cd ../build
zip esalertd-linux-x86_64.zip esalertd
cd -
