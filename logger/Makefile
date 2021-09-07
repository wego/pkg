
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

skip_gen=0

all: help

help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-16s${GREEN}%s${RESET}\n", $$1, $$2}' $(MAKEFILE_LIST)

clean: ## Clean mocks, test & log.
	find . -type d -name mocks -a -prune -exec rm -rf {} \;
	rm -rf ./log
	rm -f ./cover.out

gen: ## Clean mocks, tests, logs, then generate mocks for test.
	make clean
	go generate ./... -p $(num_of_pros)

test: ## Clean mocks, tests, logs, generate mocks for test, then run unit test. Add `skip_gen=1` to skip `make gen`
ifeq (0, $(skip_gen))
	make gen
endif
	go test ./... -cover -p $(num_of_pros)

cover: ## Run 'make gen' then run unit test with data racing check & show coverage result. Add `skip_gen=1` to skip `make gen`
ifeq (0, $(skip_gen))
	make gen
endif
	go test ./... -p $(num_of_pros) -race -coverprofile ./cover.out && go tool cover -html=./cover.out
