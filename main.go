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
	Date          string `json:"Scheduled_Time"`
	Key           string `json:"email_blast_KEY"`
	ReferenceName string `json:"Reference_Name"`
	Subject       string
	HTML          string `json:"HTML_Content"`
	Text          string `json:"Text_Content"`
}

//exists returns true if the specified file exists.
func exists(f string) bool {
	_, err := os.Stat(f)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatalf("%v %v\n", err, f)
	}
	return true
}

//proc accepts blasts from the input queue and handles them.
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

//filename parses a blast and returns a filename with the specified
//extension.
func filename(b blast, ext string) string {
	const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
	t, _ := time.Parse(form, b.Date)
	d := t.Format("2006-01-02")
	s := strings.Replace(b.Subject, "/", " ", -1)
	if len(s) == 0 {
		s = strings.Replace(b.ReferenceName, "/", " ", -1)
	}
	if len(s) == 0 {
		s = "Unknown"
	}
	s = strings.TrimSpace(s)
	return fmt.Sprintf("%v - %v - %v.%v", d, b.Key, s, ext)
}

//handle accepts a blast and writes both HTML and PDF files.
//Errors writing PDFs (e.g. an image was deleted long ago) are
//noted but not fatal.
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

//push reads the email_blast table and pushes email blast onto a queue.
func push(api *godig.API, summary bool, in chan blast) error {
	t := api.EmailBlast()
	offset := int32(0)
	c := 500
	for c >= 500 {
		var a []blast
		err := t.Many(offset, c, "Stage=Complete", &a)
		if err != nil {
			return err
		}
		log.Printf("Read %v records from offset %v\n", len(a), offset)
		c = len(a)
		offset += int32(c)
		for _, b := range a {
			if summary {
				fmt.Println(filename(b, "pdf"))
			} else {
				in <- b
			}
		}
	}
	close(in)
	return nil
}

//scrub handles the cases where resource URLs are on domains that Salsa no
//longer supports.
func scrub(x string) string {
	s := strings.Replace(x, "org2.democracyinaction.org", "org2.salsalabs.com", -1)
	s = strings.Replace(s, "salsa.democracyinaction.org", "org.salsalabs.com", -1)
	s = strings.Replace(s, "hq.demaction.org", "org.salsalabs.com", -1)
	s = strings.Replace(s, "cid:", "https:", -1)
	return s
}

//main accepts inputs form the user and processes blasts into HTML and PDF
//files.
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
		}(in, &wg)
	}
	err = push(api, *summary, in)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	wg.Wait()
}
