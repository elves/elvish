#!/bin/sh -e
go vet ./...
staticcheck ./...
