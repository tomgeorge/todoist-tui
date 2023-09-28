TODOIST_API_TOKEN := $(shell cat ~/todoist-api-token)

.PHONY : help
help : # Display help
	@awk -F ':|##' \
		'/^[^\t].+?:.*?##/ {\
			printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF \
		}' $(MAKEFILE_LIST)

.PHONY : clean
clean: ## clean app
	rm ./bin/todoist-tui || true

.PHONY : build
build: ## build app
	go build -o ./bin/todoist-tui

.PHONY : run
run: clean build ## clean and build
	TODOIST_API_TOKEN=${TODOIST_API_TOKEN} ./bin/todoist-tui
