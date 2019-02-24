#!/bin/sh

mkdir build
rm build/*
cd src

go get ./...
CC=$(which musl-gcc) go build -o ../build/esalert --ldflags '-w -s -linkmode external -extldflags "-static"' . || true 

cd ../build
zip esalert-linux-x86_64.zip esalert
cd -

cd
for file in build/*.zip
do
	echo -n "${file} - "
	./cryptocli -- \
		file --path "${file}" --read -- \
		dgst --algo sha256 -- \
		hex --encode -- \
		stdout 2> /dev/null
	echo
done
