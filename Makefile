install-tools:
	@echo installing tools
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/vektra/mockery/v2@latest
	@go install github.com/go-swagger/go-swagger/cmd/swagger@latest
	@go install github.com/cucumber/godog/cmd/godog@latest
	@echo done

generate:
	@echo running code generation
	@go generate ./...
	@echo done

clean-services:
	docker image rm mallbots-baskets mallbots-customers mallbots-stores

build-services:
	docker build -t mallbots-baskets --file docker/Dockerfile.microservices --build-arg=service=baskets .
	docker build -t mallbots-customers --file docker/Dockerfile.microservices --build-arg=service=customers .
	docker build -t mallbots-stores --file docker/Dockerfile.microservices --build-arg=service=stores .
