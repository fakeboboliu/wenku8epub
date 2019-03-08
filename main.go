package main

import (
	"archive/zip"
	"github.com/mkideal/cli"
	"os"
)

type argT struct {
	cli.Helper
	URL   string `cli:"*u,url" usage:"URL of menu page"`
	Out   string `cli:"*o,out" usage:"Output file"`
	Retry int    `cli:"r,retry" usage:"Retry times while downloading images" dft:"2"`
	NoPic bool   `cli:"nopic" usage:"Do not download images" dft:"false"`
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		f := check(os.Create(argv.Out)).(*os.File)
		z := zip.NewWriter(f)
		zop := newZipOp(z)

		genor := &EpubGenor{}
		genor.retry = argv.Retry
		genor.GetPic = !argv.NoPic
		getWenku8(argv.URL, genor)
		genor.MakeEpub(zop)

		zop.Wait()
		z.Close()
		f.Close()
		return nil
	})
}
