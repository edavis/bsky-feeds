all: bin/mostliked bin/feedweb

bin/mostliked: cmd/mostliked/*.go pkg/mostliked/*.go
	go build -o $@ ./cmd/mostliked

bin/feedweb: cmd/feedweb/*.go pkg/*/*.go mostliked/*.go
	go build -o $@ ./cmd/feedweb
