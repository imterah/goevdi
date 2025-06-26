APP_DIR := ./
OUTPUT := example
TAGS := noaudio x11

.PHONY: all build clean

all: build

build:
	go build -v -tags '$(TAGS)' -o $(OUTPUT) $(APP_DIR)

clean:
	rm -f $(OUTPUT)
