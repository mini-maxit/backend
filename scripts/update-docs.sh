#!/usr/bin/env bash
go install github.com/swaggo/swag/cmd/swag@latest

cd cmd/app
swag init --dir ./,../../internal/api/http/httputils,../../package/domain/schemas,../.. -o ../../docs --ot yaml --st
