package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/popu125/wenku8epub/epub"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/popu125/wenku8epub/helpers"
)

var (
	sels = map[string]string{
		"vol":     ".vcss",
		"rows":    ".css > tbody > tr",
		"chap":    ".ccss > a",
		"content": "#content",
		"title":   "#title",
		"author":  "#info",
	}
)

func init() {
	log.WithField("count", len(sels)).Info("Loaded selectors.")
}

func getWenku8(url string, genor *EpubGenor) {

	prefix := strings.TrimSuffix(url, "index.htm")
	args := strings.Split(url[strings.Index(url, "://")+3:], "/")
	page, novelId := args[2], args[3]

	genor.Cover = Check(ioutil.ReadAll(Check(HttpGetWithRetry(fmt.Sprintf("http://img.wkcdn.com/image/%s/%s/%ss.jpg", page, novelId, novelId), genor.Retry)).(*http.Response).Body)).([]byte)

	body := mustGet(url)
	menuPage := Check(goquery.NewDocumentFromReader(body)).(*goquery.Document)

	genor.Title = menuPage.Find(sels["title"]).Text()
	genor.Author = strings.Split(menuPage.Find(sels["author"]).Text(), "：")[1]

	var workingVol *epub.Volume
	menuPage.Find(sels["rows"]).Each(func(i int, row *goquery.Selection) {
		volSel := row.Find(sels["vol"])
		if volSel.Length() == 1 {
			if workingVol != nil {
				genor.Vols = append(genor.Vols, *workingVol)
			}
			workingVol = epub.NewVol(volSel.Text())
		} else {
			row.Find(sels["chap"]).Each(func(i int, chapSel *goquery.Selection) {
				if href, ok := chapSel.Attr("href"); ok {
					doc := getChapter(prefix + href).Find(sels["content"])
					genor.DetectAndReplacePic(doc, prefix)
					chap := Check(doc.Html()).(string)
					workingVol.Chapters = append(workingVol.Chapters, *epub.NewChapter(chapSel.Text(), chap))
				}
			})
		}
	})
	genor.Vols = append(genor.Vols, *workingVol)
}

func getChapter(url string) *goquery.Document {
	log.WithField("url", url).Info("Getting chapter...")
	body := mustGet(url)
	page := Check(goquery.NewDocumentFromReader(body)).(*goquery.Document)
	return page
}

func mustGet(url string) *mahonia.Reader {
	body := mahonia.NewDecoder("gbk").NewReader(Check(HttpGetWithRetry(url, 5)).(*http.Response).Body)
	return body
}
