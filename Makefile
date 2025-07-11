BINARY  := tubcast
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
	./$(BINARY).o -create-show --title="Test-2" --description="Show description"

sync-channel: install
	./$(BINARY).o -sync-channel --title="Test-2" --description="Show description" --channel-id="marianneBL"

sync: install
	./$(BINARY).o -sync

add-video: install
	./$(BINARY).o -add-video --title="Test-2" --description="Show description" --video-url="https://youtu.be/r0McrrrFNtA?si=8f08sySlLVL8PAP-"
