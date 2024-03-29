VERSION = "v0.1.7"
COMMIT = $(shell git rev-parse --short HEAD)

default:
	go vet ./...
	go build -o _output/gen ./cmd/gen
	go build -ldflags="-s -w" -o _output/service ./cmd/service
	go build -ldflags="-s -w" -o _output/traffic_gen ./cmd/traffic_gen

bin: default
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.Version=$(VERSION)-$(COMMIT)'" -o _output/gen_linux ./cmd/gen
	CGO_ENABLED=0 GOOS=darwin go build -ldflags="-X 'main.Version=$(VERSION)-$(COMMIT)'" -o _output/gen_macos ./cmd/gen
	CGO_ENABLED=0 GOOS=windows go build -ldflags="-X 'main.Version=$(VERSION)-$(COMMIT)'" -o _output/gen_windows ./cmd/gen

test:
	go test -v -count=1 ./...

benchmark:
	go test -benchmem -run=^$$ -bench ^BenchmarkHttpService$$ -benchmem -cpuprofile perf-http.out ./pkg/service
	#go test -benchmem -run=^$$ -bench ^BenchmarkRemoteQuery$$ -benchmem -cpuprofile perf-client.out ./pkg/service
	GOMAXPROCS=2 go test -benchmem -run=^$$ -bench ^BenchmarkNonDataHttpService$$ -benchmem -cpuprofile perf-nondata.out ./pkg/service

image:
	docker build -f service.dockerfile --target service -t docker.io/warmmetal/ms-demo-service:$(VERSION) .
	docker build -f service.dockerfile --target traffic-gen -t docker.io/warmmetal/ms-demo-traffic:$(VERSION) .

release: bin image
	docker tag docker.io/warmmetal/ms-demo-service:$(VERSION) docker.io/warmmetal/ms-demo-service:latest
	docker push docker.io/warmmetal/ms-demo-service:$(VERSION)
	docker push docker.io/warmmetal/ms-demo-service:latest
	docker tag docker.io/warmmetal/ms-demo-traffic:$(VERSION) docker.io/warmmetal/ms-demo-traffic:latest
	docker push docker.io/warmmetal/ms-demo-traffic:$(VERSION)
	docker push docker.io/warmmetal/ms-demo-traffic:latest