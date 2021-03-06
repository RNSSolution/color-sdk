PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
CAT := $(if $(filter $(OS),Windows_NT),type,cat)
LEDGER_ENABLED ?= true
GOBIN ?= $(GOPATH)/bin
GOSUM := $(shell which gosum)

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/ColorPlatform/color-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

# process linker flags

ldflags = -X github.com/ColorPlatform/color-sdk/version.Version=$(VERSION) \
	-X github.com/ColorPlatform/color-sdk/version.Commit=$(COMMIT) \
  -X "github.com/ColorPlatform/color-sdk/version.BuildTags=$(build_tags)"

ifneq ($(GOSUM),)
ldflags += -X github.com/ColorPlatform/color-sdk/version.VendorDirHash=$(shell $(GOSUM) go.sum)
endif

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/ColorPlatform/color-sdk/types.DBBackend=cleveldb
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

# Total number of leagues
LOCALNET_LEAGUES?=3
# Number of nodes in a league
LOCALNET_NODES?=3
# The local IP network
LOCALNET_NETWORK?="192.165.0.0/24"
# The starting IP address assigned to Docker containers
LOCALNET_STARTING_IP?="192.165.0.2"
# The starting port for forwarding ports of Prism executable from Docker
LOCALNET_STARTING_PORT?=26656

all: tools install lint test

# The below include contains the tools target.
include scripts/Makefile

########################################
### CI

ci: tools install test_cover lint test

########################################
### Build/Install

build: go.sum
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/colord.exe ./cmd/gaia/cmd/colord
	go build -mod=readonly $(BUILD_FLAGS) -o build/colorcli.exe ./cmd/gaia/cmd/colorcli
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/colord ./cmd/gaia/cmd/colord
	go build -mod=readonly $(BUILD_FLAGS) -o build/colorcli ./cmd/gaia/cmd/colorcli
	go build -mod=readonly $(BUILD_FLAGS) -o build/colorreplay ./cmd/gaia/cmd/colorreplay
	go build -mod=readonly $(BUILD_FLAGS) -o build/colorkeyutil ./cmd/gaia/cmd/colorkeyutil
endif

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

update_gaia_lite_docs:
	@statik -src=client/lcd/swagger-ui -dest=client/lcd -f

install: go.sum check-ledger update_gaia_lite_docs
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/gaia/cmd/colord
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/gaia/cmd/colorcli
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/gaia/cmd/colorreplay
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/gaia/cmd/colorkeyutil

install_debug: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/gaia/cmd/colordebug

dist:
	@bash publish/dist.sh
	@bash publish/publish.sh

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: tools go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw_deps: tools
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i github.com/ColorPlatform/color-sdk/cmd/gaia/cmd/gaiad -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf snapcraft-local.yaml build/

distclean: clean
	rm -rf vendor/

########################################
### Documentation

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/ColorPlatform/color-sdk/types"
	godoc -http=:6060


########################################
### Testing

test: test_unit

test_cli: build
	@go test -mod=readonly -p 4 `go list ./cmd/gaia/cli_test/...` -tags=cli_test

test_ledger:
    # First test with mock
	@go test -mod=readonly `go list github.com/ColorPlatform/color-sdk/crypto` -tags='cgo ledger test_ledger_mock'
    # Now test with a real device
	@go test -mod=readonly -v `go list github.com/ColorPlatform/color-sdk/crypto` -tags='cgo ledger'

test_unit:
	@VERSION=$(VERSION) go test -mod=readonly $(PACKAGES_NOSIMULATION) -tags='ledger test_ledger_mock'

test_race:
	@VERSION=$(VERSION) go test -mod=readonly -race $(PACKAGES_NOSIMULATION)

test_race:
	@VERSION=$(VERSION) go test -mod=readonly -race $(PACKAGES_NOSIMULATION)
	
test_sim_gaia_nondeterminism:
	@echo "Running nondeterminism test..."
	@go test -mod=readonly ./cmd/gaia/app -run TestAppStateDeterminism -SimulationEnabled=true -v -timeout 10m

test_sim_gaia_custom_genesis_fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	@go test -mod=readonly ./cmd/gaia/app -run TestFullGaiaSimulation -SimulationGenesis=${HOME}/.gaiad/config/genesis.json \
		-SimulationEnabled=true -SimulationNumBlocks=100 -SimulationBlockSize=200 -SimulationCommit=true -SimulationSeed=99 -SimulationPeriod=5 -v -timeout 24h

test_sim_gaia_fast:
	@echo "Running quick Gaia simulation. This may take several minutes..."
	@go test -mod=readonly ./cmd/gaia/app -run TestFullGaiaSimulation -SimulationEnabled=true -SimulationNumBlocks=100 -SimulationBlockSize=200 -SimulationCommit=true -SimulationSeed=99 -SimulationPeriod=5 -v -timeout 24h

test_sim_gaia_import_export:
	@echo "Running Gaia import/export simulation. This may take several minutes..."
	@bash scripts/multisim.sh 50 5 TestGaiaImportExport

test_sim_gaia_simulation_after_import:
	@echo "Running Gaia simulation-after-import. This may take several minutes..."
	@bash scripts/multisim.sh 50 5 TestGaiaSimulationAfterImport

test_sim_gaia_custom_genesis_multi_seed:
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.gaiad/config/genesis.json will be used."
	@bash scripts/multisim.sh 400 5 TestFullGaiaSimulation ${HOME}/.gaiad/config/genesis.json

test_sim_gaia_multi_seed:
	@echo "Running multi-seed Gaia simulation. This may take awhile!"
	@bash scripts/multisim.sh 400 5 TestFullGaiaSimulation

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true
test_sim_gaia_benchmark:
	@echo "Running Gaia benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ github.com/ColorPlatform/color-sdk/cmd/gaia/app -bench ^BenchmarkFullGaiaSimulation$$  \
		-SimulationEnabled=true -SimulationNumBlocks=$(SIM_NUM_BLOCKS) -SimulationBlockSize=$(SIM_BLOCK_SIZE) -SimulationCommit=$(SIM_COMMIT) -timeout 24h

test_sim_gaia_profile:
	@echo "Running Gaia benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@go test -mod=readonly -benchmem -run=^$$ github.com/ColorPlatform/color-sdk/cmd/gaia/app -bench ^BenchmarkFullGaiaSimulation$$ \
		-SimulationEnabled=true -SimulationNumBlocks=$(SIM_NUM_BLOCKS) -SimulationBlockSize=$(SIM_BLOCK_SIZE) -SimulationCommit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

test_cover:
	@export VERSION=$(VERSION); bash -x tests/test_cover.sh

lint: tools ci-lint
ci-lint:
	golangci-lint run
	go vet -composites=false -tests=false ./...
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

format: tools
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs goimports -w -local github.com/ColorPlatform/color-sdk

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)


########################################
### Devdoc

DEVDOC_SAVE = docker commit `docker ps -a -n 1 -q` devdoc:local

devdoc_init:
	docker run -it -v "$(CURDIR):/go/src/github.com/ColorPlatform/color-sdk" -w "/go/src/github.com/ColorPlatform/color-sdk" tendermint/devdoc echo
	# TODO make this safer
	$(call DEVDOC_SAVE)

devdoc:
	docker run -it -v "$(CURDIR):/go/src/github.com/ColorPlatform/color-sdk" -w "/go/src/github.com/ColorPlatform/color-sdk" devdoc:local bash

devdoc_save:
	# TODO make this safer
	$(call DEVDOC_SAVE)

devdoc_clean:
	docker rmi -f $$(docker images -f "dangling=true" -q)

devdoc_update:
	docker pull tendermint/devdoc


########################################
### Local validator nodes using docker and docker-compose

build-docker-colordnode:
	$(MAKE) -C networks/local

build/node0/colord/config/genesis.json: Makefile
	docker run --rm -v  \
			$(CURDIR)/build:/colord:Z tendermint/colordnode testnet --l $(LOCALNET_LEAGUES) --v $(LOCALNET_NODES) \
			--starting-ip-address $(LOCALNET_STARTING_IP) \
			&& \
	echo Init done; \


localnet-init: build/node0/colord/config/genesis.json
	
# Run testnet locally
localnet-start: docker-compose.yml localnet-stop localnet-init
	 docker-compose up


docker-compose.yml: Makefile
	python3 networks/local/generate-docker-compose-yml.py $(LOCALNET_LEAGUES) $(LOCALNET_NODES) $(LOCALNET_NETWORK) $(LOCALNET_STARTING_IP) $(LOCALNET_STARTING_PORT) > docker-compose.yml

# Stop testnet
localnet-stop:
	docker-compose down


########################################
### Packaging

snapcraft-local.yaml: snapcraft-local.yaml.in
	sed "s/@VERSION@/${VERSION}/g" < $< > $@

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: install install_debug dist clean distclean \
draw_deps test test_cli test_unit \
test_cover lint benchmark devdoc_init devdoc devdoc_save devdoc_update \
build-linux build-docker-gaiadnode localnet-start localnet-stop \
format check-ledger test_sim_gaia_nondeterminism test_sim_modules test_sim_gaia_fast \
test_sim_gaia_custom_genesis_fast test_sim_gaia_custom_genesis_multi_seed \
test_sim_gaia_multi_seed test_sim_gaia_import_export \
go-mod-cache
