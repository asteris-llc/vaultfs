TEST?=$$(GO15VENDOREXPERIMENT=1 go list ./... | grep -v /vendor/)
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods -nilfunc -printf -rangeloops -shift -structtags -unsafeptr

default: test
	tf_acc= go15vendorexperiment=1 go install $(test) $(testargs) -timeout=30s -parallel=4

build: fmtcheck generate
	tf_acc= go15vendorexperiment=1 go build $(test) $(testargs)


install: fmtcheck generate
	TF_ACC= GO15VENDOREXPERIMENT=1 go install $(TEST) $(TESTARGS)

# test runs the unit tests and vets the code
test: fmtcheck generate
	TF_ACC= GO15VENDOREXPERIMENT=1 go test $(TEST) $(TESTARGS) -timeout=30s -parallel=4

# testrace runs the race checker
testrace: fmtcheck generate
	TF_ACC= GO15VENDOREXPERIMENT=1 go test -race $(TEST) $(TESTARGS)

cover:
	@go tool cover 2>/dev/null; if [ $$? -eq 3 ]; then \
		go get -u golang.org/x/tools/cmd/cover; \
	fi
	GO15VENDOREXPERIMENT=1 go test $(TEST) -coverprofile=coverage.out
	GO15VENDOREXPERIMENT=1 go tool cover -html=coverage.out
	rm coverage.out

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@echo "go tool vet $(VETARGS) $(TEST) "
	@GO15VENDOREXPERIMENT=1 go tool vet $(VETARGS) $(TEST) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi


# generate runs `go generate` to build the dynamically generated
# source files.
generate:
	@GO15VENDOREXPERIMENT=1 go generate $$(GO15VENDOREXPERIMENT=1 go list ./... | grep -v /vendor/)

fmt:
	gofmt -w .

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: bin default generate test updatedeps vet fmt fmtcheck
