.phony: certs

all: certs tic-tac-toe

certs:
		make -C ssl

tic-tac-toe: *.go */*.go
		go build
		mkdir -p storage

clean:
		make -C ssl clean
		rm -f tic-tac-toe

