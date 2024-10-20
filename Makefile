cover:
	go test -json ./... -coverprofile coverprofile_.tmp -coverpkg=./... ; \
    grep -v -e 'mock.go' -e 'docs.go' -e '_easyjson.go' coverprofile_.tmp > coverprofile.tmp ; \
    rm coverprofile_.tmp ; \
    go tool cover -func coverprofile.tmp; \
    rm coverprofile.tmp

test:
	go test ./...

lint:
	golangci-lint run --config=.golangci.yaml

generate:
	go generate ./...
