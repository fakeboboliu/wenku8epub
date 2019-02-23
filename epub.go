package main

import (
	"fmt"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"math/rand"
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

func makeEpub(z *zipOp, vols []Volume, title string, cover []byte) {
	z.WriteFile("OPS/images/cover.jpg", cover)
	z.WriteFile("OPS/css/main.css", check(tplBox.Find("main.css")).([]byte))
	z.WriteFile("mimetype", check(tplBox.Find("mimetype")).([]byte))
	z.WriteFile("META-INF/container.xml", check(tplBox.Find("container.xml")).([]byte))

	for _, vol := range vols {
		fmt.Println("h")
		for _, chap := range vol.Chapters {
			w := z.Writer(fmt.Sprintf("OPS/c_%d.html", chap.ID))
			noErr(tpls["chapter.html"].Execute(w, chap))
			w.Flush()
		}
	}

	type bookInfo struct {
		Title   string
		Volumes []Volume
		Time    string
		UUID    string
	}
	w := z.Writer("OPS/fb.opf")
	bi := bookInfo{Title: title, Volumes: vols, Time: time.Now().String(), UUID: randUUID()}
	noErr(tpls["fb.opf"].Execute(w, bi))
	w.Flush()

	w = z.Writer("OPS/nav.html")
	noErr(tpls["nav.html"].Execute(w, vols))
	w.Flush()

	z.Done()
}

func randUUID() string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	uuid := []byte("")
	for i := 36; i > 0; i-- {
		switch i {
		case 27, 22, 17, 12:
			uuid = append(uuid, 45) // 45 is "-"
		default:
			uuid = append(uuid, chars[rand.Intn(36)])
		}
	}
	return string(uuid)
}

func noErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}
