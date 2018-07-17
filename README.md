# fix_dia
Go program to write PDFs for email blasts.  If the blasts have URLs that contain `democracyinaction.org`, then those are changed to the correct
hostname on `salsalabs.com` before being converted to PDFs.
# Background
Salsa used to have a domain named `democracyinaction.org`.  That was turned down in favor of using  `salsalabs.com`.

Clients that uploaded and used images and files when `democracyinaction.org` was alive still have email blasts that reference that domain.

Salsa offsers a service that retrieves PDFs for completed email blasts. This service breaks down if image and file URLs are located on
`democracyinaction.org`.  This app solves that problem by fixing those URLs before the HTML is converted to a PDF.

## Login credentials

The `fix_dia` application looks for your login credentials in a YAML file.  You provide the filename as part of the execution.

You can read up on YAML and its formatting rules [here](https://en.wikipedia.org/wiki/YAML) if you'd like.

  The easiest way to get started is to  copy the `sample_login.yaml` file and edit it.  Here's an example.
```yaml
host: wfc2.wiredforchange.com
email: chuck@echeese.bizi
password: extra-super-secret-password!
```
The `email` and `password` are the ones that you normally use to log in. The `host` can be found by using [this page](https://help.salsalabs.com/hc/en-us/articles/115000341773-Salsa-Application-Program-Interface-API-#api_host) in Salsa's documentation.

Save the new login YAML file to disk.  We'll need it when we  run the `fix_dia` app.

#Installation
```bash
go get "github.com/salsalabs/godig"

go get "github.com/salsalabs/fix_dia"

go install

#Usage
```bash
go run main.go --credentials YAML_Credentials_File [--all]] [[--count number]]
```
Use
```go run main.go --help
```
to see the help.  YOu sould see something like this:
```
A command-line app to read email blasts, correct DIA URLs and write PDFs.

Flags:
  --help         Show context-sensitive help (also try --help-long and --help-man).
  --login=LOGIN  YAML file with login credentials
  --all          save all blasts, not just the ones with DIA links
  --count=5      Start this number of processors.
```
#Output.

The application creates two directories.

* `html`: the modified HTML for each of the blasts.
* `pdfs1: the PDFs for each of the blasts.

#Questions?  Comments?
Use the [Issues link](https://github.com/salsalabs/fix_dia/issues) in the repository.  Don't waste your time by contacting Salsa support.
