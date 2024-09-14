LOG_FILE_NAME = log.txt

FLAGS = -race

.PHONY: run build test clean

run:
	cd cmd && go run $(FLAGS) main.go > $(LOG_FILE_NAME) 2>&1

run-no-flags:
	cd cmd && go run main.go > $(LOG_FILE_NAME) 2>&1

build:
	cd cmd && go build main.go

test:
	cd test && go test -v

clean:
	cd cmd && rm -f main $(LOG_FILE_NAME)
