all: bin/mostliked bin/feedweb

bin/mostliked: cmd/mostliked/*.go pkg/mostliked/*.go db/mostliked/*.go
	go build -o $@ ./cmd/mostliked

bin/feedweb: cmd/feedweb/*.go pkg/*/*.go db/*/*.go
	go build -o $@ ./cmd/feedweb

.PHONY: clean
clean:
	rm bin/*
