#!/usr/bin/env bash
go install github.com/mini-maxit/swag/cmd/swag@latest

swag init --dir ./cmd/app,./internal/api/http/httputils,./package/domain/schemas,. -o ./docs --ot yaml --st
