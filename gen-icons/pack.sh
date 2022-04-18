#!/usr/bin/bash
echo "Compiling binary"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go

echo "Creating deployment.zip file"
7z a deployment.zip main

echo "Cleaning up"
rm main