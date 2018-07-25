# classic_blasts_to_pdfs
Go program to write PDFs for email blasts using `wkhtmlpdf`.

If the blasts have URLs that contain `democracyinaction.org`, then those are changed to the correct hostname on `salsalabs.com` before being converted to PDFs.

# Background
From time-to-time, Salsa's clients want to retrieve their email blasts for reference.  Unfortunately, that's something that Salas Classic can't do.

But, Salsa Classic has an "email blast archive" that shows email blasts as HTML.  This program leverages that fact by retrieving the HTML, then passing it to [`wkhtmltopdf`](https://wkhtmltopdf.org/).  `Wkhtmlmpdf` converts the HTML to a PDF and stores it on disk.

# Domain fixes
Salsa used to have a domain named `democracyinaction.org`.  That domain was turned off in favor of using `salsalabs.com`.

Clients that uploaded and used images and files when `democracyinaction.org` was alive still have email blasts that reference that domain.

Attempting to retrieve images and files for the old domain doesn't work.  `Wkthmltopdf` does his best to recover.  However, the PDFs are generally blank when the use images and files from `deocracyinaction.org`.

This app solves that problem by automatically modifying URls for the old domain.

## Login credentials

The `classic_blasts_to_pdfs` application looks for your login credentials in a YAML file.  You provide the filename as part of the execution.

  The easiest way to get started is to  copy the `sample_login.yaml` file and edit it.  Here's an example.
```yaml
host: wfc2.wiredforchange.com
email: chuck@chew.cheese
password: extra-super-secret-password!
```
The `email` and `password` are the ones that you normally use to log in. The `host` can be found by using [this page](https://help.salsalabs.com/hc/en-us/articles/115000341773-Salsa-Application-Program-Interface-API-#api_host) in Salsa's documentation.

Save the new login YAML file to disk.  We'll need it when we  run the `classic_blasts_to_pdfs` app.

# Installation
```bash
go get "github.com/salsalabs/godig"

go get "github.com/salsalabs/classic_blasts_to_pdfs"

go install
```

# Usage
```bash
go run main.go --credentials YAML_Credentials_File [--all]] [[--count number]]
```
Use
```go run main.go --help
```
to see the help.  You sould see something like this:
```
A command-line app to read email blasts, correct DIA URLs and write PDFs.

Flags:
  --help         Show context-sensitive help (also try --help-long and --help-man).
  --login=LOGIN  YAML file with login credentials
  --count=10     Start this number of processors.
```
# Output.

The application creates two directories.

* `html`: the modified HTML for each of the blasts.
* `pdfs`: the PDFs for each of the blasts.

# Questions?  Comments?
Use the [Issues link](https://github.com/salsalabs/classic_blasts_to_pdfs/issues) in the repository.  Don't waste your time by contacting Salsa support.
