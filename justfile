lz: lint
	lazygit

fmt:
	gofumpt -l -w .

lint: fmt
	golangci-lint run

w:
	watchexec -e go just lint 

