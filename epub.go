package main

import (
	"fmt"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"text/template"
	"time"
)

type Volume struct {
	Name     string
	Chapters []Chapter
}

type Chapter struct {
	ID   int
	Name string
	Text string
}

type EpubGenor struct {
	Vols   []Volume
	Cover  []byte
	Title  string
	Author string

	Pics   []Picture
	picC   int
	GetPic bool

	retry int
}

var (
	_chapter_counter = 0

	tpls   = map[string]*template.Template{}
	tplBox *packr.Box
)

func init() {
	tplBox = packr.New("tplBox", "./tpl")
	for _, name := range tplBox.List() {
		s := check(tplBox.FindString(name)).(string)
		tpls[name] = template.Must(template.New("").Parse(s))
		log.WithField("name", name).Info("Loaded template.")
	}
}

func newVol(name string) *Volume {
	return &Volume{Name: name, Chapters: make([]Chapter, 0)}
}

func newChapter(name string, text string) *Chapter {
	_chapter_counter++
	return &Chapter{ID: _chapter_counter, Name: name, Text: text}
}

func (g *EpubGenor) MakeEpub(z *zipOp) {
	z.WriteFile("OPS/images/cover.jpg", g.Cover)
	z.WriteFile("OPS/css/main.css", check(tplBox.Find("main.css")).([]byte))
	z.WriteFile("mimetype", check(tplBox.Find("mimetype")).([]byte))
	z.WriteFile("META-INF/container.xml", check(tplBox.Find("container.xml")).([]byte))

	for _, vol := range g.Vols {
		for _, chap := range vol.Chapters {
			w := z.Writer(fmt.Sprintf("OPS/c_%d.html", chap.ID))
			noErr(tpls["chapter.html"].Execute(w, chap))
			w.Flush()
		}
	}

	for _, pic := range g.Pics {
		z.WriteFile(fmt.Sprint("OPS/images/", pic.Filename()), pic.Data)
	}

	type bookInfo struct {
		Title    string
		Author   string
		Volumes  []Volume
		Pictures []Picture
		Time     string
		UUID     string
	}
	w := z.Writer("OPS/fb.opf")
	bi := bookInfo{Title: g.Title, Volumes: g.Vols, Pictures: g.Pics, Author: g.Author, Time: time.Now().String(), UUID: randUUID()}
	noErr(tpls["fb.opf"].Execute(w, bi))
	w.Flush()

	w = z.Writer("OPS/nav.html")
	noErr(tpls["nav.html"].Execute(w, g.Vols))
	w.Flush()

	z.Done()
}

func (g *EpubGenor) picID() int {
	g.picC++
	return g.picC
}
