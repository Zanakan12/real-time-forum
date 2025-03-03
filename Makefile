.DEFAULT_GOAL := run

run:
	go run -tags sqlite_userauth cmd/golang-server-layout/main.go

.PHONY: run
