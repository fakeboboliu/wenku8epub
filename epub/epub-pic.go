package epub

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	. "github.com/popu125/wenku8epub/helpers"
	log "github.com/sirupsen/logrus"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"mime"
	"net/url"
	"path"
	"strings"
	"sync"
)

type Picture struct {
	ID        int
	Data      []byte
	Ext       string
	MediaType string
	url       string
}

func (pic Picture) Filename() string {
	return fmt.Sprint("p_", pic.ID, pic.Ext)
}

func (g *EpubGenor) DetectAndReplacePic(sel *goquery.Selection, prefix string) {
	if !g.GetPic {
		return
	}

	if sel.Find("img").Length() > 0 {
		var wg sync.WaitGroup
		picChan := make(chan *Picture, 64)
		done := make(chan bool)

		go func() {
			for pic := range picChan {
				go func(pic *Picture) {
					wg.Add(1)
					data, err := getPic(pic.url, g.Retry)
					log.WithField("src", pic.url).Info("Got <img>")
					if err != nil {
						log.WithField("src", pic.url).WithField("err", err).Info("Get <img> failed")
						wg.Done()
						return
					}
					pic.Data = data
					g.Pics = append(g.Pics, pic)
					wg.Done()
				}(pic)
			}
			wg.Wait()
			done <- true
		}()

		sel.Find("img").Each(func(i int, imgSel *goquery.Selection) {
			if src, ok := imgSel.Attr("src"); ok {
				log.WithField("src", src).Info("Found <img>")
				var ext string
				if u, err := url.Parse(src); err != nil {
					return
				} else {
					ext = path.Ext(u.Path)
					if !u.IsAbs() {
						src = fmt.Sprint(prefix, src)
					}
				}
				pic := &Picture{ID: g.picID(), Ext: ext, MediaType: mime.TypeByExtension(ext), url: src}
				imgSel.SetAttr("src", fmt.Sprint("images/p_", pic.ID, pic.Ext))
				picChan <- pic
			}
		})

		close(picChan)
		<-done
	}

}

func getPic(src string, retry int) ([]byte, error) {
	resp, err := HttpGetWithRetry(src, retry)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (g *EpubGenor) CompressPic(quality int) {
	for _, pic := range g.Pics {
		oldSize := len(pic.Data)

		f := bytes.NewReader(pic.Data)
		var img image.Image
		var err error
		switch strings.ToLower(pic.MediaType) {
		case "image/jpeg":
			img, err = jpeg.Decode(f)
		case "image/png":
			img, err = png.Decode(f)
		case "image/gif":
			img, err = gif.Decode(f)
		}
		if err != nil {
			log.WithField("pic", pic.Filename()).WithError(err).Error()
			return
		}

		buf := new(bytes.Buffer)
		jpeg.Encode(buf, img, &jpeg.Options{Quality: 50})
		pic.Data = buf.Bytes()
		pic.Ext = ".jpg"

		log.WithField("before", oldSize).WithField("after", len(pic.Data)).WithField("pic", pic.ID).Info("Compressed image")
	}
}
