clean:
	rm ./bin/todoist-tui
build:
	go build -o ./bin/todoist-tui
run: clean build
	TODOIST_API_TOKEN=${TODOIST_API_TOKEN} ./bin/todoist-tui
