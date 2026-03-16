APP := ccmonitor

.PHONY: build run test clean

build:
	go build -o $(APP) .

run: build
	./$(APP)

test:
	go test ./...

clean:
	rm -f $(APP)
