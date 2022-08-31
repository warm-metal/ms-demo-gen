default:
	go vet ./...
	go build -o _output/gen ./cmd/gen
	go build -o _output/service ./cmd/service

test:
	go test -v -count=1 ./...

image:
	docker build -f service.dockerfile -t docker.io/warmmetal/ms-demo-service:v0.1.0 .