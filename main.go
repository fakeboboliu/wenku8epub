package main

import (
	"archive/zip"
	"github.com/mkideal/cli"
	"github.com/popu125/wenku8epub/epub"
	"os"

	. "github.com/popu125/wenku8epub/helpers"
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
		f := Check(os.Create(argv.Out)).(*os.File)
		z := zip.NewWriter(f)
		zop := epub.NewZipOp(z)

		genor := &epub.EpubGenor{}
		genor.Retry = argv.Retry
		genor.GetPic = !argv.NoPic
		src := NewWenku8(argv.URL, genor)
		src.GetMenu()
		for _, vol := range genor.Vols {
			src.GetVolume(vol)
		}
		genor.MakeEpub(zop)

		zop.Wait()
		z.Close()
		f.Close()
		return nil
	})
}
