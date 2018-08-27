package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const html = "html"
const pdfs = "pdfs"

type blast struct {
	Date    string `json:"Date_Created"`
	Key     string `json:"email_blast_KEY"`
	Subject string
	HTML    string `json:"HTML_Content"`
	Text    string `json:"Text_Content"`
}

func exists(x string) bool {
	_, err := os.Stat(x)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatalf("%v %v\n", err, x)
	}
	return true
}

func proc(in chan blast) {
	for {
		select {
		case b, ok := <-in:
			if !ok {
				return
			}
			err := handle(b)
			if err != nil {
				log.Printf("proc: key %v, %v\n", b.Key, err)
			}
		}
	}
}

func filename(b blast, ext string) string {
	const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
	t, _ := time.Parse(form, b.Date)
	d := t.Format("2006-01-02")

	s := strings.Replace(b.Subject, "/", " ", -1)
	if len(s) == 0 {
		s = "No Title"
	}
	s = strings.TrimSpace(s)

	return fmt.Sprintf("%v - %v - %v.%v", d, b.Key, s, ext)
}

func handle(b blast) error {
	// Create new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}

	// Set global options
	pdfg.Dpi.Set(600)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeLetter)
	pdfg.Grayscale.Set(false)

	s := scrub(b.HTML)
	fn := filename(b, "html")
	fn = path.Join(html, fn)
	if exists(fn) {
		log.Printf("%s: HTML already exists\n", b.Key)
		return nil
	}
	buf := []byte(s)
	ioutil.WriteFile(fn, buf, os.ModePerm)
	//log.Printf("wrote %s\n", fn)

	fn = filename(b, "pdf")
	fn = path.Join(pdfs, fn)
	if exists(fn) {
		log.Printf("%s already exists\n", fn)
		return nil
	}

	page := wkhtmltopdf.NewPageReader(bytes.NewReader(buf))
	page.DisableSmartShrinking.Set(true)
	page.LoadErrorHandling.Set("ignore")
	page.LoadMediaErrorHandling.Set("ignore")
	page.Zoom.Set(1.0)
	pdfg.AddPage(page)
	err = pdfg.Create()
	if err != nil {
		return fmt.Errorf("create error on %v: %s", b.Key, err)
	}

	err = pdfg.WriteFile(fn)
	if err != nil {
		return err
	}
	log.Printf("wrote %s\n", fn)
	return nil
}

func scrub(x string) string {
	s := strings.Replace(x, "org2.democracyinaction.org", "org2.salsalabs.com", -1)
	s = strings.Replace(s, "salsa.democracyinaction.org", "org.salsalabs.com", -1)
	s = strings.Replace(s, "cid:", "https:", -1)
	return s
}

func main() {
	var (
		app     = kingpin.New("classic_blasts_to_pdfs", "A command-line app to read email blasts, correct DIA URLs and write PDFs.")
		login   = app.Flag("login", "YAML file with login credentials").Required().String()
		count   = app.Flag("count", "Start this number of processors.").Default("10").Int()
		summary = app.Flag("summary", "Show blast dates, keys and subjects.  Does not write PDFs").Default("false").Bool()
	)
	app.Parse(os.Args[1:])
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	if !exists(pdfs) {
		err := os.Mkdir(pdfs, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("%v, %v\n", err, pdfs)
		}
	}
	if !exists(html) {
		err := os.Mkdir(html, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("%v, %v\n", err, html)
		}
	}

	var wg sync.WaitGroup
	in := make(chan blast)
	for i := 0; i < *count; i++ {
		go func(in chan blast, wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()
			proc(in)
			if err != nil {
				log.Fatal(err)
			}
		}(in, &wg)
	}
	t := godig.Table{API: api, Name: "email_blast"}
	offset := int32(0)
	c := 500
	for c >= 500 {
		var a []blast
		err := t.Many(offset, c, "Stage=Complete", &a)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Read %v records from offset %v\n", len(a), offset)
		c = len(a)
		offset += int32(c)
		for _, b := range a {
			if *summary {
				fmt.Println(filename(b, "pdf"))
			} else {
				in <- b
			}
		}
	}
	close(in)
	wg.Wait()
}
