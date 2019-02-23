builder = go build -ldflags="-s -w" -x -o dist/wenku8
package = github.com/popu125/wenku8epub

all: linux windows macos

linux:
	GOOS=linux GOARCH=386 $(builder)-linux-386 $(package)
	GOOS=linux GOARCH=amd64 $(builder)-linux-amd64 $(package)

windows:
	GOOS=windows GOARCH=386 $(builder)-win32.exe $(package)
	GOOS=windows GOARCH=amd64 $(builder)-win_x86_64.exe $(package)

macos:
	GOOS=darwin GOARCH=amd64 $(builder)-macos-amd64 $(package)

upx:
	upx dist/*

prepare:
	go get -x github.com/popu125/wenku8epub

clean:
	rm -rf dist/*