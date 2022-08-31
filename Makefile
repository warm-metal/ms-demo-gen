default:
	go vet ./...
	go build -o _output/gen ./cmd/gen
	go build -o _output/service ./cmd/service
	go build -o _output/traffic_gen ./cmd/traffic_gen

test:
	go test -v -count=1 ./...

image:
	docker build -f service.dockerfile --target service -t docker.io/warmmetal/ms-demo-service:v0.1.0 .
	docker build -f service.dockerfile --target traffic-gen -t docker.io/warmmetal/ms-demo-traffic:v0.1.0 .