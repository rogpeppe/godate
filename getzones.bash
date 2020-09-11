#!/bin/bash

# copied from $GOROOT/lib/time/update.bash

CODE=2020a
DATA=2020a

set -e
WORK="$(pwd)/work"
rm -rf work
mkdir work
cd ./work
trap 'rm -r $WORK' 0
curl -sSLO https://www.iana.org/time-zones/repository/releases/tzcode$CODE.tar.gz
curl -sSLO  https://www.iana.org/time-zones/repository/releases/tzdata$DATA.tar.gz
tar xzf tzdata$DATA.tar.gz
tar xzf tzcode$CODE.tar.gz
make --silent tzdata.zi
{
	cat << "EOF"
	// Code xx generated by getzones.bash. DO NOT EDIT.

	package main

	var zoneNames = map[string]string{
EOF
	awk  '
	/^Z/ {
		printf("\t\"%s\": \"\",\n", $2)
	}
	/^L/ {
		printf("\t\"%s\": \"%s\",\n", $3, $2)
	}' tzdata.zi
	echo '}'
} | gofmt > ../zonenames.go