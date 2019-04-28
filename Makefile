builder = go build -ldflags="-s -w" -x -o dist/wenku8
package = github.com/popu125/wenku8epub/cli

all: linux windows macos upx

packr:
	cd cli
	packr2
	cd ..

linux:
	GOOS=linux GOARCH=386 $(builder)-linux-386 $(package)
	GOOS=linux GOARCH=amd64 $(builder)-linux-amd64 $(package)

windows:
	GOOS=windows GOARCH=386 $(builder)-win32.exe $(package)
	GOOS=windows GOARCH=amd64 $(builder)-win_x86_64.exe $(package)

macos:
	GOOS=darwin GOARCH=amd64 $(builder)-macos-amd64 $(package)

upx:
	upx dist/wenku8-*

prepare:
	go get -d -x github.com/popu125/wenku8epub/cli

clean:
	rm -rf dist/wenku8-*
	packr2 clean