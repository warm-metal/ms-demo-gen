default:
	go vet ./...
	go build -o _output/gen ./cmd/gen
	go build -o _output/service ./cmd/service