
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

num_of_pros:=1
OS:=$(shell uname -s)

ifeq ($(OS),Linux)
	num_of_pros := $(shell grep -c ^processor /proc/cpuinfo)
else ifeq ($(OS),Darwin)
	num_of_pros := $(shell sysctl hw.ncpu | awk '{print $$2}')
endif

ifeq ($(shell test $(num_of_pros) -gt 4; echo $$?),0)
	num_of_pros := $(shell echo $(num_of_pros)/2 | bc)
endif

all: help

help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-16s${GREEN}%s${RESET}\n", $$1, $$2}' $(MAKEFILE_LIST)

clean: ## Clean mocks, cover test.
	rm -f ./cover.out

test: ## Clean, tests, logs for test, then run unit test.
	go test ./... -race -cover -p $(num_of_pros)

cover: ## Run unit test with data racing check & show coverage result.
	go test ./... -p $(num_of_pros) -race -coverprofile ./cover.out && go tool cover -html=./cover.out
