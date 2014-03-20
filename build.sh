#!/bin/sh

templar -dir=templates -dst=common/templates.go -type=template && \
	templar -dir=static -dst=common/blobs.go -type=blob && \
	go build -o diplicity/diplicity diplicity/diplicity.go 

