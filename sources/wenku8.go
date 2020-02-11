package sources

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

type wenku8Src struct {
	url   string
	genor *epub.EpubGenor
}

func NewWenku8(url string, genor *epub.EpubGenor) *wenku8Src {
	return &wenku8Src{url: url, genor: genor,}
}

func (s wenku8Src) GetMenu() {
	args := strings.Split(s.url[strings.Index(s.url, "://")+3:], "/")
	page, novelId := args[2], args[3]

	s.genor.Cover = Check(ioutil.ReadAll(Check(HttpGetWithRetry(fmt.Sprintf("http://img.wenku8.com/image/%s/%s/%ss.jpg", page, novelId, novelId), s.genor.Retry)).(*http.Response).Body)).([]byte)

	body := mustGet(s.url)
	menuPage := Check(goquery.NewDocumentFromReader(body)).(*goquery.Document)

	s.genor.Title = menuPage.Find(sels["title"]).Text()
	s.genor.Author = strings.Split(menuPage.Find(sels["author"]).Text(), "：")[1]

	var workingVol *epub.Volume
	menuPage.Find(sels["rows"]).Each(func(i int, row *goquery.Selection) {
		volSel := row.Find(sels["vol"])
		if volSel.Length() == 1 {
			log.WithField("name", volSel.Text()).Info("New Volume")
			if workingVol != nil {
				s.genor.Vols = append(s.genor.Vols, workingVol)
			}
			workingVol = epub.NewVol(volSel.Text())
		} else {
			row.Find(sels["chap"]).Each(func(i int, chapSel *goquery.Selection) {
				if href, ok := chapSel.Attr("href"); ok {
					log.WithField("name", chapSel.Text()).Info("New Chapter")
					workingVol.Chapters = append(workingVol.Chapters, epub.NewChapter(chapSel.Text(), href))
				}
			})
		}
	})
	s.genor.Vols = append(s.genor.Vols, workingVol)
}

func (s wenku8Src) GetVolume(vol *epub.Volume) {
	prefix := strings.TrimSuffix(s.url, "index.htm")
	for _, chap := range vol.Chapters {
		doc := getChapter(prefix + chap.URL).Find(sels["content"])
		s.genor.DetectAndReplacePic(doc, prefix)
		chap.Text = Check(doc.Html()).(string)
	}
	vol.Downloaded = true
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
