package epub

import (
	"embed"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	. "github.com/popu125/wenku8epub/helpers"

	log "github.com/sirupsen/logrus"
)

type Volume struct {
	Name       string
	Chapters   []*Chapter
	Downloaded bool
}

type Chapter struct {
	ChapNo int
	Name   string
	URL    string
	Text   string
}

type EpubGenor struct {
	Vols   []*Volume
	Cover  []byte
	Title  string
	Author string

	Pics   []*Picture
	picC   int
	GetPic bool

	Retry int
}

//go:embed tpl/*
var tpl embed.FS

var (
	_chapter_counter = 0

	tpls         = map[string]*template.Template{}
	builtinFiles = map[string][]byte{}
)

func init() {
	tplList, err := tpl.ReadDir("tpl")
	if err != nil {
		panic(err)
	}
	for _, f := range tplList {
		name := f.Name()
		data := Check(tpl.ReadFile("tpl/" + name)).([]byte)
		tpls[name] = template.Must(template.New("").Parse(string(data)))
		builtinFiles[name] = data
		log.WithField("name", name).Info("Loaded template.")
	}
}

func NewVol(name string) *Volume {
	return &Volume{Name: name, Chapters: make([]*Chapter, 0)}
}

func NewChapter(name string, url string) *Chapter {
	_chapter_counter++
	return &Chapter{ChapNo: _chapter_counter, Name: name, URL: url}
}

func (g *EpubGenor) MakeEpub(z *zipOp) {
	z.WriteFile("OPS/images/cover.jpg", g.Cover)
	z.WriteFile("OPS/css/main.css", builtinFiles["main.css"])
	z.WriteFile("mimetype", builtinFiles["mimetype"])
	z.WriteFile("META-INF/container.xml", builtinFiles["container.xml"])

	for _, vol := range g.Vols {
		for _, chap := range vol.Chapters {
			w := z.Writer(fmt.Sprintf("OPS/c_%d.html", chap.ChapNo))
			NoErr(tpls["chapter.html"].Execute(w, chap))
			w.Flush()
		}
	}

	for _, pic := range g.Pics {
		z.WriteFile(fmt.Sprint("OPS/images/", pic.Filename()), pic.Data)
	}

	type bookInfo struct {
		Title    string
		Author   string
		Volumes  []*Volume
		Pictures []*Picture
		Time     string
		UUID     string
	}
	w := z.Writer("OPS/fb.opf")
	bi := bookInfo{Title: g.Title, Volumes: g.Vols, Pictures: g.Pics, Author: g.Author, Time: time.Now().String(), UUID: RandUUID()}
	NoErr(tpls["fb.opf"].Execute(w, bi))
	w.Flush()

	w = z.Writer("OPS/nav.html")
	NoErr(tpls["nav.html"].Execute(w, g.Vols))
	w.Flush()

	z.Done()
}

func (g *EpubGenor) picID() int {
	g.picC++
	return g.picC
}

func (g *EpubGenor) DetectTitle() {
	downs := make([]string, 0)
	for no, vol := range g.Vols {
		if vol.Downloaded {
			downs = append(downs, strconv.Itoa(no))
		}
	}

	if len(downs) != len(g.Vols) {
		g.Title = fmt.Sprint(g.Title, "-", strings.Join(downs, ","))
	}
}
