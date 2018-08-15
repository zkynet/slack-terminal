#!/bin/bash
GOOS=linux GOARCH=amd64 go build -o linux/bot .
go build -o mac/bot .
GOOS=windows GOARCH=386 go build -o windows/bot.exe .