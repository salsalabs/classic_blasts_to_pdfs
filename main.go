//Multicopy is a multi-threaded URL retriever.  You provide
//provide login credentials to an instance of Salas Classic.
//Multicopy walks the directory tree in the images and files
//repository and saves files to disk.  Files are stored in the
//same structure on disk as they appear in the repository.
//
// Installation:
//
// go get github.com/salsalabs/multicopy
//
// go install github.com/salsalabs/multicopy
//
// Execution:
//
// multicopy --login [YAML file] --dir [DIR]
//
// Help:
//
// multicopy --help
//
package main

import (
	"log"
	"os"
	"sync"

	"github.com/salsalabs/godig"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Blast is the structure for an email blast.
type Blast struct {
	Key  string `json:"email_blast_KEY"`
	HTML string `json:"HTML_Content"`
	Text string `json"Text"`
}

//Get returns the email blast with the provided key.
func Get(t *godig.Table, key string) (*Blast, error) {
	return nil, nil
}

//Put saves the provided email blast to the database.
func Put(t *godig.Table, b *Blast) error {
	return nil
}

//main is the application.  Gathers arguments, starts listeners, reads
//URLs and processes them.
func main() {
	var (
		app   = kingpin.New("multicopy", "A command-line app to copy images and files from a Salsa HQ to your disk.")
		login = app.Flag("login", "YAML file with login credentials").Required().String()
		dir   = app.Flag("dir", "Store contents starting in this directory.").Default(".").String()
		count = app.Flag("count", "Start this number of processors.").Default("20").Int()
	)
	app.Parse(os.Args[1:])

	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	files := make(chan string)
	done := make(chan bool)
	var wg sync.WaitGroup

	// Start the processors.
	for i := 1; i <= *count; i++ {
		go func(i int) {
			wg.Add(1)
			defer wg.Done()
			Run(api, *dir, files, done)
		}(i)
	}

	// Start processing folders at the root dir.  Load will use
	// itself to process subdirs.
	err = Load(api, "/", files)
	if err != nil {
		panic(err)
	}

	// Tell the processors that we're through.
	close(done)

	// Wait for everything to finish.
	wg.Wait()
}
