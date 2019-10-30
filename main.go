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
	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const html = "html"
const pdfs = "pdfs"

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

type guide struct {
	Extension string
	Directory string
}

type env struct {
	Summary  bool
	HTMLOnly bool
	Guide    guide
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
func (e env) proc(in chan blast) {
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

//filename parses a blast and creates directory and filenames.  The Directory
//is assured to exist.  Returns the full filename to write.
func (e env) filename(b blast) (fn string, err error) {
	const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
	x := b.Date
	if len(x) == 0 {
		x = b.LastModified
	}
	if len(x) == 0 {
		x = b.DateRequested
	}
	t, _ := time.Parse(form, x)
	when := t.Format("2006-01-02")
	s := strings.Replace(b.Subject, "/", " ", -1)
	if len(s) == 0 {
		s = strings.Replace(b.ReferenceName, "/", " ", -1)
	}
	if len(s) == 0 {
		s = "Unknown"
	}
	s = strings.TrimSpace(s)
	y := t.Format("2006")
	dir := path.Join(e.Guide.Directory, y)
	fn = fmt.Sprintf("%v - %v - %v.%v", when, b.Key, s, e.Guide.Extension)
	if !e.Summary && !exists(dir) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fn, err
		}
	}
	fn = path.Join(dir, fn)
	return fn, nil
}

//handle accepts a blast and writes both HTML and PDF files.
//Errors writing PDFs (e.g. an image was deleted long ago) are
//noted but not fatal.
func (e env) handle(b blast) error {
	fn, err := e.filename(b)
	if err != nil {
		return err
	}
	bn := path.Base(fn)
	if exists(fn) {
		log.Printf("%s already exists\n", bn)
		return nil
	}
	if e.Summary {
		fmt.Printf("%s\n", bn)
		return nil
	}

	s := scrub(b.HTML)
	buf := []byte(s)

	// Write HTML.
	if e.HTMLOnly {
		err = ioutil.WriteFile(fn, buf, os.ModePerm)
		if err != nil {
			return err
		}
		log.Printf("%s\n", bn)
		return nil
	}

	// Write PDF.
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}

	// Set global options
	pdfg.Dpi.Set(600)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeLetter)
	pdfg.Grayscale.Set(false)

	// Blast goes into a PDF page.
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(buf))

	//Set page options.
	page.DisableSmartShrinking.Set(true)
	page.LoadErrorHandling.Set("ignore")
	page.LoadMediaErrorHandling.Set("ignore")
	page.Zoom.Set(0.9)
	pdfg.AddPage(page)
	err = pdfg.Create()
	if err != nil {
		return fmt.Errorf("%s on %v", err, fn)
	}

	err = pdfg.WriteFile(fn)
	if err != nil {
		return err
	}
	log.Printf("%s\n", bn)
	return nil
}

//push reads the email_blast table and pushes email blast onto a queue.
func (e env) push(api *godig.API, in chan blast) error {
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
			in <- b
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
	s = strings.Replace(s, `"/salsa/`, `"https://salsalabs.com/salsa/`, -1)
	return s
}

//main accepts inputs form the user and processes blasts into HTML and PDF
//files.
func main() {
	var (
		app        = kingpin.New("classic_blasts_to_pdfs", "A command-line app to read email blasts, correct DIA URLs and write PDFs.")
		login      = app.Flag("login", "YAML file with login credentials").Required().String()
		count      = app.Flag("count", "Start this number of processors.").Default("10").Int()
		summary    = app.Flag("summary", "Show blast dates, keys and subjects. Do not write PDFs.").Default("false").Bool()
		htmlOnly   = app.Flag("htmlOnly", "Write HTML. Do not write PDFs.").Default("false").Bool()
		apiVerbose = app.Flag("apiVerbose", "Display API calls and responses. Very noisy...").Default("false").Bool()
	)
	app.Parse(os.Args[1:])
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	api.Verbose = *apiVerbose

	g := guide{
		Directory: "blast_pdfs",
		Extension: "pdf",
	}
	if *htmlOnly {
		g = guide{
			Directory: "html",
			Extension: "html",
		}
	}
	e := env{
		Summary:  *summary,
		HTMLOnly: *htmlOnly,
		Guide:    g,
	}

	var wg sync.WaitGroup
	in := make(chan blast)
	for i := 0; i < *count; i++ {
		go func(in chan blast, wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()
			e.proc(in)
		}(in, &wg)
	}
	err = e.push(api, in)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	wg.Wait()
}
