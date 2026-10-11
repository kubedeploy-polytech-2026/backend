SHELL := /bin/bash
.DEFAULT_GOAL := help

IMAGE ?= kubedeploy-backend:v0.1.0

.PHONY: help test build image

help:
	@printf '%s\n' \
		'Targets:' \
		'  test   Run unit tests' \
		'  build  Build the backend binary' \
		'  image  Build the local container image'

test:
	@go test ./...

build:
	@go build -o backend ./cmd/server

image:
	@docker build --tag $(IMAGE) .
