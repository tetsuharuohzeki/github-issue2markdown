DIST_NAME := jikiru

clean:
	rm -rf ./$(DIST_NAME)
	rm -rf ./articles
	mkdir ./articles

bootstrap:
	rm -rf vendor/
	go get -u github.com/golang/dep/...
	dep ensure

build: $(DIST_NAME)

run: $(DIST_NAME)
	./$(DIST_NAME)

$(DIST_NAME): clean
	go build -o $(DIST_NAME)

travis:
	$(MAKE) bootstrap -C $(CURDIR)
	$(MAKE) build -C $(CURDIR)
