.PHONY: test fmt vet

test:
	go test -v -count=1 ./...

fmt:
	gofmt -w .

vet:
	go vet ./...
