.PHONY: generate

generate:
	@go generate ./...
	@echo "[OK] Files added to embed box!"

build: generate
	@go build -o ./build/checksum-CRC32 ./main.go
	@echo "[OK] App binary was created!"
