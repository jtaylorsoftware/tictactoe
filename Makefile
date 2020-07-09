APP=tictactoe
HOST_PORT=5432
CONTAINER_PORT=5432

.PHONY: build
build:
	docker build -t $(APP) -f build/Dockerfile .
build-dev:
	mkdir -p bin && go build -o bin/ . 
clean:
	go clean .
	rm bin/$(APP)
	docker rm $(APP)
run:
	docker run -a stdout --rm -p=$(HOST_PORT):$(CONTAINER_PORT) --name="$(APP)" $(APP)
run-dev:
	bin/$(APP)
stop:
	docker stop $(APP)