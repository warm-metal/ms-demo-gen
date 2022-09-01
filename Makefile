VERSION = "v0.1.1"

default:
	go vet ./...
	go build -o _output/gen ./cmd/gen
	go build -ldflags="-s -w" -o _output/service ./cmd/service
	go build -ldflags="-s -w" -o _output/traffic_gen ./cmd/traffic_gen

bin: default
	CGO_ENABLED=0 GOOS=linux go build -o _output/gen_linux ./cmd/gen
	CGO_ENABLED=0 GOOS=darwin go build -o _output/gen_macos ./cmd/gen
	CGO_ENABLED=0 GOOS=windows go build -o _output/gen_windows ./cmd/gen

test:
	go test -v -count=1 ./...

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