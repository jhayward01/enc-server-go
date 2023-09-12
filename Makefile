# make help                # Print makefile reference
help:
	@grep -e "^\# make" Makefile |  cut -c 3-

# make fmt                 # Run format and static analysis tools
fmt:
	gofmt -s -w .
	go vet ./...
	staticcheck ./...

# make build               # Build repo
build:
	go build -v ./...
	
# make tests               # Test repo
tests:
	go test -v ./...

# make build-all           # Format, build, and test repo
build-all: fmt build test
	
# make install-client-fe   # Install FE client
install-client-fe:
	go install -v cmd/feclient.go 
	
# make install-server-be   # Install BE server
install-server-be:
	go install -v cmd/beserver.go 

# make install-servers     # Install BE/FE servers
install-servers: install-server-be install-server-fe

# make all                 # Install all binaries
all: install-client install-servers

# make client              # Run FE client
client: install-client-fe
	feclient

# make servers             # Run BE/FE servers in docker-compose
servers:
	docker-compose up -d --build
	
# make itests              # Run integration tests
itests: install-client
	./itest/itests.sh
	
# make stop                # Stop BE/FE servers in docker-compose
stop:
	docker-compose down

# make server-be-cmd       # Run BE server in terminal
server-be-cmd: install-servers
	ENC_SERVER_GO_CONFIG_PATH='config/config.cmd.yaml' beserver

# make server-fe-cmd       # Run FE server in terminal
server-fe-cmd: install-servers
	ENC_SERVER_GO_CONFIG_PATH='config/config.cmd.yaml' feserver

# make start_cluster       # Start application in local Kubernetes cluster
start-cluster: 
	./k8s/start_cluster.sh

# make stop_cluster        # Stop application in local Kubernetes cluster
stop-cluster: 
	./k8s/stop_cluster.sh
