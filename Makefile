BINARY  := tubecast
DESTDIR := $(HOME)/bin

.PHONY: all build install clean

all: build

build:
	go build -o $(BINARY).o .

install: build
	mkdir -p $(DESTDIR)
	install -m0755 $(BINARY).o $(DESTDIR)/$(BINARY)

clean:
	rm -f $(BINARY).o

create-show: install
	./$(BINARY).o -create-show --title="med" --description="Show description"

sync-channel: install
	./$(BINARY).o -sync-channel --title="med" --channel-id="ThePrimeTimeagen"

sync: install
	./$(BINARY).o -sync

add-video: install
	./$(BINARY).o -add-video --title="medicine" --description="Show description" --video-url="https://youtu.be/ix5nrWedyqg?si=v-Bk4OyGInzVnFZ2"
