all: bin/mostliked bin/feedweb

bin/mostliked: cmd/mostliked/main.go pkg/mostliked/handler.go db/mostliked/*.go pkg/feeds/*.go
	go build -o $@ ./cmd/mostliked

bin/feedweb: cmd/feedweb/*.go pkg/*/generator.go db/*/*.go pkg/feeds/*.go
	go build -o $@ ./cmd/feedweb

.PHONY: clean
clean:
	rm bin/*
