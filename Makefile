NAME = moogo
ARCH = amd64
OS = darwin
#linux darwin windows
ifeq ($(OS),windows)
	OUTNAME = $(NAME).exe
else
	OUTNAME = $(NAME)_$(OS)
endif
all:
	CGO_ENABLED=0 GOARCH=$(ARCH) GOOS=$(OS) go build -x -v -ldflags "-w" -o ./build/$(OUTNAME) main.go
	upx -9 ./build/$(OUTNAME)
.PHONY : clean
clean:
	rm -rf ./build/*