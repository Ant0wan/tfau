GO   := go
NAME := tfau

all: build

build:
	CGO_ENABLED=0 GOOS=linux $(GO) build -v -o $(NAME)
run:
	CGO_ENABLED=0 GOOS=linux $(GO) run -v $(NAME)

test:
	$(GO) test -v

install: build
	$(GO) install

clean:
	rm -rf $(NAME)

fclean: clean

.PHONY: all build test clean fclean install
