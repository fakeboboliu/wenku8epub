package main

import (
	"archive/zip"
	"github.com/mkideal/cli"
	"os"
)

type argT struct {
	cli.Helper
	URL string `cli:"*u,url" usage:"URL of menu page"`
	Out string `cli:"*o,out" usage:"Output file"`
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		f := check(os.Create(argv.Out)).(*os.File)
		z := zip.NewWriter(f)
		zop := newZipOp(z)

		genor := getWenku8(argv.URL)
		genor.MakeEpub(zop)

		zop.Wait()
		z.Close()
		f.Close()
		return nil
	})
}
