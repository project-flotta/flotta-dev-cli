##@ Build

build: generate-tools
	go build -mod=vendor -o bin/flotta main.go

##@ Development

generate-tools:
ifeq (, $(shell which go-homedir))
	(cd /tmp && go get github.com/mitchellh/go-homedir)
endif
ifeq (, $(shell which cobra))
	(cd /tmp/ && go get github.com/spf13/cobra)
endif
ifeq (, $(shell which viper))
	(cd /tmp/ && go get github.com/spf13/viper)
endif
	@exit

vendor:
	go mod tidy -go=1.16 && go mod tidy -go=1.17
	go mod vendor
