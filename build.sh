#!/bin/bash
date
echo "Truncating sum ..." && truncate -s0 go.sum
#echo "Downloading mods ..." && go mod download
#echo "Downloading new version of modules ..." && go get -v -u
echo "Removing unnecessary libraries ..." && go mod tidy
podman build -t go-pyrecognizer-fe .
#echo "Building module ..." && go build -o main
#echo "Stripping executable ..." && strip -s main