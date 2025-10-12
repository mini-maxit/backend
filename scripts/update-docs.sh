#!/usr/bin/env bash

cd cmd/app
swag init --dir ./,../../internal/api/http/httputils,../../package/domain/schemas,../.. -o ../../docs --ot yaml
