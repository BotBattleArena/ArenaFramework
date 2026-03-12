.PHONY: proto build test clean run-example

# Tools
PROTOC = protoc
# Check if local protoc exists
ifneq ("$(wildcard bin/protoc/bin/protoc.exe)","")
	PROTOC = bin\protoc\bin\protoc.exe
endif

# install dependencies
install: install-go install-protoc

install-go:
	go mod download
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

install-protoc:
	@if not exist bin mkdir bin
	@echo Downloading protoc...
	@powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-win64.zip -OutFile protoc.zip"
	@echo Extracting protoc...
	@powershell -Command "Expand-Archive -Path protoc.zip -DestinationPath bin/protoc -Force"
	@del protoc.zip
	@echo protoc installed in bin/protoc

# Generate code from proto files
proto: proto-go proto-csharp

proto-go:
	@if not exist gen\go mkdir gen\go
	$(PROTOC) --go_out=./gen/go --go_opt=paths=source_relative \
		--proto_path=proto \
		proto/arena/v1/arena.proto

proto-csharp:
	@if not exist gen\csharp mkdir gen\csharp
	$(PROTOC) --csharp_out=./gen/csharp \
		--proto_path=proto \
		proto/arena/v1/arena.proto

# Build all executables
build:
	go build ./...

# Build example binaries
build-examples:
	go build -o bin/tictactoe.exe ./cmd/tictactoe
	go build -o bin/randombot.exe ./cmd/randombot

# Run example: build randombot, set up inputs dir, run tictactoe
run-example: build-examples
	@if not exist inputs mkdir inputs
	@copy /y bin\randombot.exe inputs\player1.exe >nul
	@copy /y bin\randombot.exe inputs\player2.exe >nul
	bin\tictactoe.exe

# Run tests
test:
	go test ./...

# Clean build artifacts (keeps tools)
clean:
	@if exist bin\tictactoe.exe del /q bin\tictactoe.exe
	@if exist bin\randombot.exe del /q bin\randombot.exe
	@if exist inputs rmdir /s /q inputs
	@if exist gen rmdir /s /q gen
