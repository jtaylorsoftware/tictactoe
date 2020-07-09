APP = tictactoe
BINARY_PATH = $$GOBIN/$(APP)

build:
	@go build -o  bin/ . 2>make.log
	@echo "built $(APP) to bin/" | tee make.log
clean:
	@go clean 
install:
	@go install . 2>make.log
	@echo installed to $(BINARY_PATH) | tee make.log
uninstall:
	$(eval RET := $(shell rm $(BINARY_PATH) 2>make.log; echo $$?))
	@if [ $(RET) -eq 1 ]; then	\
		echo "failed to uninstall $(BINARY_PATH) - see make.log";	\
	else	\
		echo "uninstalled $(APP)";	\
	fi;