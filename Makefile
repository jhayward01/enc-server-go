# make help                # Print makefile reference
help:
	@grep -e "^\# make" Makefile |  cut -c 3-

# make fmt                 # Run format and static analysis tools
fmt:
	gofmt -s -w .
	go vet ./...

# make proto               # Build GRPC protos
proto:
	 protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/v2-apis/be/service/service.proto

# make build               # Build repo
build:
	go build -v ./...

# make test                # Test repo
test::
	go test -v ./...

# make build-all           # Format, build, and test repo
build-all: fmt proto build test
	
# make install-client-fe   # Install FE client
install-client-fe:
	go install -v cmd/feclient/feclient.go 

# make install-client-be   # Install BE client
install-client-be:
	go install -v cmd/beclient/beclient.go 

# make install-server-fe   # Install FE server
install-server-fe:
	go install -v cmd/feserver/feserver.go 
	
# make install-server-be   # Install BE server
install-server-be:
	go install -v cmd/beserver/beserver.go 

# make install-servers     # Install BE/FE servers
install-servers: install-server-be install-server-fe

# make all                 # Install all binaries
all: install-client install-servers

# make client              # Run FE client
client: install-client-fe
	feclient --v2

# make be-client           # Run BE client
be-client: install-client-be
	beclient --v2

# make servers             # Run BE/FE servers in docker-compose
servers:
	docker compose up -d --build
	
# make itests              # Run integration tests
itests: install-client-fe
	./test/tests.sh
	
# make stop                # Stop BE/FE servers in docker-compose
stop:
	docker compose down

# make server-be-cmd       # Run BE server in terminal
server-be-cmd: install-servers
	ENC_SERVER_GO_CONFIG_PATH='config/config.cmd.yaml' beserver --v2

# make server-fe-cmd       # Run FE server in terminal
server-fe-cmd: install-servers
	ENC_SERVER_GO_CONFIG_PATH='config/config.cmd.yaml' feserver --v2

# make start-cluster       # Start application in local Kubernetes cluster
start-cluster: 
	./deployments/minikube/start_cluster.sh

# make stop-cluster        # Stop application in local Kubernetes cluster
stop-cluster: 
	./deployments/minikube/stop_cluster.sh
