GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=receivehugoemail


build:
	GO111MODULE=on $(GOBUILD) -v ./cmd/aws/$(BINARY_NAME).go

build-linux:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v ./cmd/aws/$(BINARY_NAME).go
	GO111MODULE=on  $(GOGET) -u github.com/aws/aws-lambda-go/cmd/build-lambda-zip
	${GOBIN}/build-lambda-zip.exe -o $(BINARY_NAME).zip $(BINARY_NAME)

run: build
	./$(BINARY_NAME).exe
	