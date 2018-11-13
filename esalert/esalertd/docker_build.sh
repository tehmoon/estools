#!/bin/sh
set -e

image_name="esalertd_builder"
container_name="esalertd_builder"

CWD=$( cd "$(dirname "$0")"; pwd -P)
cd "${CWD}"

docker stop -t 0 "${container_name}" || true
docker build -t "${image_name}" .
docker run -d --rm --name "${container_name}" "${image_name}"

docker cp ./compile.sh "${container_name}:/home/builder/compile.sh"
docker cp ./src "${container_name}:/home/builder/src"

docker exec "${container_name}" sh compile.sh

rm build/* || true
docker cp "${container_name}:/home/builder/build/esalertd" build/esalertd

if [ "$(id -u)" = 0 ]
then
	ug=$(stat --printf=%U:%G .)
	chown -R "${ug}" build/*
fi

docker stop -t 0 "${container_name}"
