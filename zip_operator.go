package main

import (
	"archive/zip"
	"bytes"
	log "github.com/sirupsen/logrus"
)

type zipOp struct {
	w    *zip.Writer
	c    chan task
	done chan bool
}

type zipW struct {
	fn  string
	z   *zipOp
	buf *bytes.Buffer
}

func (zw *zipW) Write(p []byte) (n int, err error) {
	return zw.buf.Write(p)
}

func (zw *zipW) Flush() {
	zw.z.WriteFile(zw.fn, zw.buf.Bytes())
}

func newZipOp(w *zip.Writer) *zipOp {
	op := &zipOp{w: w, c: make(chan task, 64), done: make(chan bool)}
	go op.zipWriter()
	return op
}

func (z *zipOp) WriteFile(name string, data []byte) {
	log.WithField("file", name).Info("Writing file to zip")
	d := make([]byte, len(data))
	copy(d, data)
	z.c <- task{Name: name, Data: d}
}

func (z *zipOp) Writer(name string) *zipW {
	return &zipW{fn: name, z: z, buf: new(bytes.Buffer)}
}

func (z *zipOp) zipWriter() {
	for t := range z.c {
		a, err := z.w.Create(t.Name)
		if err != nil {
			panic(err)
		}
		n, _ := a.Write(t.Data)
		log.WithField("from", "zipOp").Info("Wrote ", n, " bytes to ", t.Name)
	}
	z.done <- true
}

func (z *zipOp) Done() {
	close(z.c)
}

func (z *zipOp) Wait() {
	<-z.done
}

type task struct {
	Name string
	Data []byte
}
