GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOBIN=$(GOPATH)/bin
ZIPBUILD=$(GOBIN)/build-lambda-zip
HUGO=hugolambda
READ=readsend
RECEIVE=receivehugoemail

build-all: build-readsend build-receivehugoemail build-hugolambda

build-all-linux: 
	$(MAKE) build-linux BINARY_NAME=$(READ)
	$(MAKE) build-linux BINARY_NAME=$(HUGO)
	$(MAKE) build-linux BINARY_NAME=$(RECEIVE)

build: check-binary
	echo $(BINARY_NAME)
	GO111MODULE=on $(GOBUILD) -v ./cmd/aws/$(BINARY_NAME)/$(BINARY_NAME).go

build-linux: check-binary
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v ./cmd/aws/$(BINARY_NAME)/$(BINARY_NAME).go
	
ifeq ("$(wildcard $(ZIPBUILD))","")
	GO111MODULE=on  $(GOGET) -u github.com/aws/aws-lambda-go/cmd/build-lambda-zip
endif
ifeq ($(BINARY_NAME),$(HUGO))
		$(GOBIN)/build-lambda-zip -o $(BINARY_NAME).zip $(BINARY_NAME) hugo
else
		$(GOBIN)/build-lambda-zip -o $(BINARY_NAME).zip $(BINARY_NAME)
endif

run: check-binary build
	./$(BINARY_NAME).exe

test:
	GO111MODULE=on $(GOTEST) -v ./...

build-readsend:
	$(MAKE) build BINARY_NAME=readsend	

build-receivehugoemail:
	$(MAKE) build BINARY_NAME=receivehugoemail

build-hugolambda:
	$(MAKE) build BINARY_NAME=hugolambda

clean: 
	rm *.exe *.zip readsend receivehugoemail hugolambda

check-binary:
ifndef BINARY_NAME
	$(error BINARY_NAME is undefined)
endif


	