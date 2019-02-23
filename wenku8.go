package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	sels = map[string]string{
		"vol":     ".vcss",
		"rows":    ".css > tbody > tr",
		"chap":    ".ccss > a",
		"content": "#content",
		"title":   "#title",
	}
)

func init() {
	log.WithField("count", len(sels)).Info("Loaded selectors.")
}

func getWenku8(url string) ([]Volume, string, []byte) {
	prefix := strings.TrimSuffix(url, "index.htm")
	args := strings.Split(url[strings.Index(url, "://")+3:], "/")
	page, novelId := args[2], args[3]

	cover := check(ioutil.ReadAll(check(http.Get(fmt.Sprintf("http://img.wkcdn.com/image/%s/%s/%ss.jpg", page, novelId, novelId))).(*http.Response).Body)).([]byte)

	body := mustGet(url)

	menuPage := check(goquery.NewDocumentFromReader(body)).(*goquery.Document)
	vols := make([]Volume, 0)

	title := menuPage.Find(sels["title"]).Text()

	var workingVol *Volume
	menuPage.Find(sels["rows"]).Each(func(i int, row *goquery.Selection) {
		volSel := row.Find(sels["vol"])
		if volSel.Length() == 1 {
			if workingVol != nil {
				vols = append(vols, *workingVol)
			}
			workingVol = newVol(volSel.Text())
		} else {
			row.Find(sels["chap"]).Each(func(i int, chapSel *goquery.Selection) {
				if href, ok := chapSel.Attr("href"); ok {
					chap := getChapter(prefix + href)
					workingVol.Chapters = append(workingVol.Chapters, *newChapter(chapSel.Text(), chap))
				}
			})
		}
	})

	return vols, title, cover
}

func getChapter(url string) string {
	log.WithField("url", url).Info("Getting chapter...")
	body := mustGet(url)
	page := check(goquery.NewDocumentFromReader(body)).(*goquery.Document)
	content := check(page.Find(sels["content"]).Html()).(string)
	return content
}

func mustGet(url string) *mahonia.Reader {
	body := mahonia.NewDecoder("gbk").NewReader(check(http.Get(url)).(*http.Response).Body)
	return body
}
