#!/bin/sh
set -x
set -e

image_name="esquery_builder"
container_name="esquery_builder"

CWD=$( cd "$(dirname "$0")"; pwd -P)
cd "${CWD}"

docker stop -t 0 "${container_name}" || true
docker build --pull -t "${image_name}" .
docker run -d --rm --name "${container_name}" "${image_name}"

docker cp ./compile.sh "${container_name}:/home/builder/compile.sh"
docker cp ./src "${container_name}:/home/builder/src"


rm -rf build || true
mkdir build || true

for a in linux windows openbsd darwin
do
	docker exec "${container_name}" sh compile.sh ${a}
	docker cp "${container_name}:/home/builder/build/esquery-${a}-x86_64.zip" build/
done

if [ "$(id -u)" = 0 ]
then
	ug=$(stat --printf=%U:%G .)
	chown -R "${ug}" build/*
fi

docker stop -t 0 "${container_name}"
