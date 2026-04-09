.PHONY: dev build run clean

dev:
	air

run:
	go run ./cmd/api/main.go

build:
	go build -o server.exe ./cmd/api/main.go

clean:
	@if exist tmp rmdir /s /q tmp
	@if exist server.exe del server.exe
