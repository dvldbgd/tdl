build:
	go build -o tdl .

install: build
	sudo mv tdl /usr/local/bin/

