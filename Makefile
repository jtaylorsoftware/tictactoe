APP=tictactoe

.PHONY: image clean cleanimage cleaninstall

image:
	sudo docker build -t $(APP) -f build/Dockerfile .

$(APP):
	go install -v ./...

clean: cleaninstall cleanimage

cleaninstall:
	go clean -i ./...

cleanimage:
	sudo docker rmi $(APP)