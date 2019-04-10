#!/bin/sh

mkdir build
rm build/*
cd src

export GOOS=${1}
go get ./...
linkmode=internal

[ "$GOOS" = "linux" ] && linkmode=external

CC=$(which musl-gcc) go build -o ../build/esquery --ldflags "-w -s -linkmode ${linkmode} -extldflags \"-static\"" . || true 

cd ../build
zip "esquery-${GOOS}-x86_64.zip" esquery
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
