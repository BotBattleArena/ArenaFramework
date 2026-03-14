.PHONY: build test clean run

# install dependencies
install:
	go mod download

# build example
build:
	go build -o bin/tictactoe.exe ./cmd/tictactoe
	go build -o bin/randombot.exe ./cmd/randombot

# run example
run: clean build
	@if not exist bots\inputs mkdir bots\inputs
	@copy /y bin\randombot.exe bots\inputs\randombot1.exe >nul
	@copy /y bin\randombot.exe bots\inputs\randombot2.exe >nul
	bin\tictactoe.exe

# Clean build artifacts
clean:
	@if exist bin rmdir /s /q bin
	@if exist bots\inputs rmdir /s /q bots\inputs
