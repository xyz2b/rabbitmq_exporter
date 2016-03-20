VERSION  := 0.10.0
TARGET   := rabbitmq_exporter
ROOTPKG := github.com/kbudde/$(TARGET)

include Makefile.COMMON


docker:
	cd $(SELFLINK) && CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -o $(TARGET)_static && docker build -f Dockerfile.scratch -t $(TARGET):scratch .
	

.PHONY: docker