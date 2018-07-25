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

const template = "http://org2.salsalabs.com/o/6931/t/0/blastContent.jsp?email_blast_KEY=%s"
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
		panic(err)
	}
	return true
}

func proc(in chan blast, done chan bool) {
	for {
		select {
		case b := <-in:
			err := handle(b)
			if err != nil {
				log.Printf("%v: %v\n", b.Key, err)
			}
		case <-done:
			return
		}
	}
}

func filename(b blast) string {
	const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
	t, _ := time.Parse(form, b.Date)
	d := t.Format("2006-01-02")

	s := strings.Replace(b.Subject, "/", " ", -1)
	if len(s) == 0 {
		s = "No Title"
	}
	s = strings.TrimSpace(s)

	return fmt.Sprintf("%v - %v - %v.html", d, b.Key, s)
}

func handle(b blast) error {
	// Create new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal(err)
	}

	// Set global options
	pdfg.Dpi.Set(600)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeLetter)
	pdfg.Grayscale.Set(false)

	s := scrub(b.HTML)
	fn := filename(b)
	fn = path.Join(html, fn)
	if exists(fn) {
		log.Printf("%s: HTML already exists\n", b.Key)
		return nil
	}
	buf := []byte(s)
	ioutil.WriteFile(fn, buf, os.ModePerm)
	//log.Printf("wrote %s\n", fn)

	fn = filename(b)
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
		log.Fatalf("Create error on %v:\n%s\n\n", b.Key, err)
	}

	err = pdfg.WriteFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s\n", fn)
	return nil
}

func scrub(x string) string {
	s := strings.Replace(x, "org2.democracyinaction.org", "org2.salsalabs.com", -1)
	s = strings.Replace(s, "salsa.democracyinaction.org", "org.salsalabs.com", -1)
	return s
}

func main() {
	var (
		app     = kingpin.New("fix_dia", "A command-line app to read email blasts, correct DIA URLs and write PDFs.")
		login   = app.Flag("login", "YAML file with login credentials").Required().String()
		count   = app.Flag("count", "Start this number of processors.").Default("5").Int()
		summary = app.Flag("summary", "Show blast keys and subjects.  Do not write PDFs").Default("false").Bool()
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
	done := make(chan bool)
	for i := 0; i < *count; i++ {
		go func(in chan blast, done chan bool) {
			wg.Add(1)
			defer wg.Done()
			proc(in, done)
			if err != nil {
				log.Fatal(err)
			}
		}(in, done)
		//log.Printf("Started processor %v\n", i+1)
	}
	t := godig.Table{API: api, Name: "email_blast"}
	offset := 0
	c := 500
	for c >= 500 {
		var a []blast
		err := t.Many(offset, c, "Stage=Complete", &a)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Read %v records from offset %v\n", len(a), offset)
		offset += c
		c = len(a)
		for _, b := range a {
			if *summary {
				fmt.Println(filename(b))
			} else {
				in <- b
			}
		}
	}
	close(done)
	wg.Wait()
}
