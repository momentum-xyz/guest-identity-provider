DOCKER_IMAGE="guest-identity-provider"
DOCKER_TAG="develop"

build : output/guest-identity-provider

clean :
	rm -rf output/

run :
	go run cmd/main.go


test :
	go test -coverprofile=output/coverage.out ./...

container-image : output/container.img.tar

run-container: docker-run

output/guest-identity-provider :
	go build -o output/guest-identity-provider ./cmd/main.go

output/container.img.tar : docker-save

docker-build : DOCKER_BUILDKIT=1
docker-build :
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .

docker-save : docker-build
	mkdir -p output
	docker save -o output/container.img.tar ${DOCKER_IMAGE}:${DOCKER_TAG}


docker-run : output/container.img.tar
	docker run --rm ${DOCKER_IMAGE}:${DOCKER_TAG}

.PHONY: clean build run test container-image run-container docker-build docker-save

