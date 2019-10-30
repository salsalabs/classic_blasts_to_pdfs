package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const html = "html"
const pdfs = "blast_pdfs"

type blast struct {
	Date          string `json:"Scheduled_Time"`
	LastModified  string `json:"Last_Modified"`
	DateRequested string `json:"Date_Requested"`
	Key           string `json:"email_blast_KEY"`
	ReferenceName string `json:"Reference_Name"`
	Subject       string
	HTML          string `json:"HTML_Content"`
	Text          string `json:"Text_Content"`
}

type env struct {
	API      *godig.API
	Zips     map[string]*zip.Writer
	Summary  bool
	HTMLOnly bool
}

//newEnv builds a new enviornment using the options from the command line.
func newEnv(api *godig.API, summary bool, htmlOnly bool) *env {
	e := env{
		API:      api,
		Zips:     make(map[string]*zip.Writer),
		Summary:  summary,
		HTMLOnly: htmlOnly,
	}
	return &e
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
func (e *env) proc(in chan blast) {
	for {
		select {
		case b, ok := <-in:
			if !ok {
				return
			}
			err := e.handle(b)
			if err != nil {
				log.Printf("proc: key %v, %v\n", b.Key, err)
			}
		}
	}
}

//filename parses a blast and creates a filename with the specified
//extension.
func (e *env) filename(b blast, ext string) (fn string, year string) {
	const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
	x := b.Date
	if len(x) == 0 {
		x = b.LastModified
	}
	if len(x) == 0 {
		x = b.DateRequested
	}
	t, _ := time.Parse(form, x)
	d := t.Format("2006-01-02")
	year = t.Format("2006")
	s := strings.Replace(b.Subject, "/", " ", -1)
	if len(s) == 0 {
		s = strings.Replace(b.ReferenceName, "/", " ", -1)
	}
	if len(s) == 0 {
		s = "Unknown"
	}

	reg, _ := regexp.Compile("[^a-zA-Z0-9 ]+")
	s = reg.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	fn = fmt.Sprintf("%v - %v - %v.%v", d, b.Key, s, ext)

	fn = path.Join(ext, year, fn)
	return fn, year
}

//parentDir creates a the enclosing directory for a filename.
func parentDir(fn string) error {
	dir := path.Dir(fn)
	err := os.MkdirAll(dir, os.ModePerm)
	return err
}

//handle accepts a blast and writes both HTML and PDF files.
//Errors writing PDFs (e.g. an image was deleted long ago) are
//noted but not fatal.
func (e *env) handle(b blast) error {
	fn, _ := e.filename(b, "html")
	if exists(fn) {
		log.Printf("HTML already exists, %s\n", fn)
		return nil
	}

	s := scrub(b.HTML)
	buf := []byte(s)
	err := parentDir(fn)
	if err != nil {
		return err
	}
	ioutil.WriteFile(fn, buf, os.ModePerm)
	bn := path.Base(fn)
	log.Printf("wrote %s\n", bn)

	if e.HTMLOnly {
		return nil
	}

	// Create new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}

	// Set global options
	pdfg.Dpi.Set(600)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeLegal)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.Grayscale.Set(false)
	fn, year := e.filename(b, "pdf")

	// Add a single page for the blast contents in HTML.
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(buf))
	page.DisableSmartShrinking.Set(true)
	page.LoadErrorHandling.Set("ignore")
	page.LoadMediaErrorHandling.Set("ignore")
	page.Zoom.Set(0.9)
	pdfg.AddPage(page)

	err = pdfg.Create()
	if err != nil {
		return fmt.Errorf("create error on %v: %s", b.Key, err)
	}

	//Write to the current year's ZIP writer
	zipWriter, ok := e.Zips[year]
	if !ok {
		//Create a new zip writer
		zipPath := path.Join(pdfs, year)
		zipPath = fmt.Sprintf("%s.zip", zipPath)
		err = parentDir(zipPath)
		if err != nil {
			return err
		}
		w, err := os.Create(zipPath)
		if err != nil {
			m := fmt.Sprintf("Error: %v created zip archive for %v\n", err, year)
			err = errors.New(m)
			return err
		}
		zipWriter = zip.NewWriter(w)
		e.Zips[year] = zipWriter
	}
	//Add a PDF file to the ZIP.
	w, err := zipWriter.Create(fn)
	if err != nil {
		return err

	}
	_, err = w.Write(pdfg.Bytes())
	if err != nil {
		return err
	}
	err = zipWriter.Flush()
	if err != nil {
		return err
	}
	return nil
}

//push reads the email_blast table and pushes email blast onto a queue.
func (e *env) push(in chan blast) error {
	t := e.API.EmailBlast()
	offset := int32(0)
	c := 500
	for c >= 500 {
		var a []blast
		//Add this for email blasts in 2018
		//&condition=Scheduled_Time LIKE 2018%git st
		err := t.Many(offset, c, "Stage=Complete", &a)
		if err != nil {
			return err
		}
		log.Printf("Read %v records from offset %v\n", len(a), offset)
		c = len(a)
		offset += int32(c)
		for _, b := range a {
			if e.Summary {
				fn, _ := e.filename(b, "pdf")
				fmt.Println(fn)
			} else {
				in <- b
			}
		}
	}
	close(in)
	return nil
}

//scrub handles the cases where resource URLs are on domains that Salsa no
//longer supports.  Also handles the danged hash marks.
func scrub(x string) string {
	s := strings.Replace(x, "org2.democracyinaction.org", "org2.salsalabs.com", -1)
	s = strings.Replace(s, "salsa.democracyinaction.org", "org.salsalabs.com", -1)
	s = strings.Replace(s, "hq.demaction.org", "org.salsalabs.com", -1)
	s = strings.Replace(s, "cid:", "https:", -1)
	s = strings.Replace(s, "/salsa/include", "https://salsalabs.com/salsa/include", -1)
	//s = strings.Replace(s, "#", "%23", -1)
	return s
}

//main accepts inputs form the user and processes blasts into HTML and PDF
//files.
func main() {
	var (
		app      = kingpin.New("classic_blasts_to_pdfs", "A command-line app to read email blasts, correct DIA URLs and write PDFs.")
		login    = app.Flag("login", "YAML file with login credentials").Required().String()
		count    = app.Flag("count", "Start this number of processors.").Default("10").Int()
		summary  = app.Flag("summary", "Show blast dates, keys and subjects. Does not write PDFs.").Default("false").Bool()
		htmlOnly = app.Flag("htmlOnly", "Write HTML. Does not write PDFs.").Default("false").Bool()
	)
	app.Parse(os.Args[1:])
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	e := newEnv(api, *summary, *htmlOnly)
	var wg sync.WaitGroup
	in := make(chan blast)
	for i := 0; i < *count; i++ {
		go func(in chan blast, wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()
			e.proc(in)
		}(in, &wg)
	}
	err = e.push(in)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	wg.Wait()
	// close all of the ZIP streams.  Any failure is fatal.
	for k, v := range e.Zips {
		err = v.Close()
		if err != nil {
			m := fmt.Sprintf("Error %v closing ZIP archive for %v\n", e, k)
			err := errors.New(m)
			panic(err)
		}
	}
}
