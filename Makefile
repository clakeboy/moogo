NAME = redpack
ARCH = amd64
OS = linux
#linux
all:
	CGO_ENABLED=0 GOARCH=$(ARCH) GOOS=$(OS) go build -x -v -ldflags "-w" -o $(NAME) main.go
	upx -9 $(name)
.PHONY : clean
clean:
	rm -f $(name)