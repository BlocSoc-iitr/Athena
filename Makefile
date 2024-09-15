unit:
	@echo "Running unit tests"
	go test ./athena_abi

unit-verbose:
	@echo "Running unit tests in verbose mode"
	go test -v ./athena_abi

clean:
	@echo "Cleaning up"
	go clean
