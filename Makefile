GO := go
NAME := tfau
TARGET_OS := $(shell go env GOOS)
TARGET_ARCH := $(shell go env GOARCH)

all: build

build:
	CGO_ENABLED=0 GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) $(GO) build -v -o $(NAME)

run:
	CGO_ENABLED=0 GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) $(GO) run -v $(NAME)

test:
	$(GO) test -v

install: build
	$(GO) install

clean:
	rm -rf $(NAME)

fclean: clean

.PHONY: all build test clean fclean install

