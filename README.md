# classic_blasts_to_pdfs

Go program to write PDFs for email blasts using `wkhtmlpdf`.

# Background

From time to time, Salsa's clients want to retrieve their email blasts for reference.  Unfortunately, that's something that Salas Classic can't do.

But, Salsa Classic has an "email blast archive" that shows email blasts as HTML.  This program leverages that fact by retrieving the HTML from the email blast archive, then passing it to [`wkhtmltopdf`](https://wkhtmltopdf.org/).  `Wkhtmlmpdf` converts the HTML to a PDF.  Internal logic adds the PDF to a ZIP archive for the year that the blast was sent.

# Domain fixes

Salsa used to have a domain named `democracyinaction.org`.  That domain was turned off in favor of `salsalabs.com`.

Clients that uploaded and used images and files when `democracyinaction.org` was alive still have email blasts that reference that domain.

Attempting to retrieve images and files for the old domain doesn't work.  `Wkthmltopdf` does his best to recover.  However, the PDFs are generally blank when the use images and files from `deocracyinaction.org`.

This app solves that problem by automatically modifying URLs that contain the old domain to point to `salsalabs.com`.

## Login credentials

The `classic_blasts_to_pdfs` application looks for your login credentials in a YAML file.  You provide the filename as part of the execution.

  The easiest way to get started is to  copy the `sample_login.yaml` file and edit it.  Here's an example.

```yaml
host: salsa4.salalabs.com
email: chuck@chew.cheese
password: extra-super-secret-password!
```

The `email` and `password` are the ones that you normally use to log in. The `host` can be found by using [this page](https://help.salsalabs.com/hc/en-us/articles/115000341773-Salsa-Application-Program-Interface-API-#api_host) in Salsa's documentation.

Save the new login YAML file to disk.  We'll need it when we  run the `classic_blasts_to_pdfs` app.

# Installation

## Install wkhtmltopdf

[Click here](https://wkhtmltopdf.org/) to download the wkhtmltopdf application.

It's a snap if you're on windows or on Linux.
The tough part about all of this is installing the package in MacOSX. It's a snap to install in Windows or Linux, not so easy in OSX. You'll need to read [OSX: About Gatekeeper](https://support.apple.com/en-us/HT202491).

See the section named "How to open an app from a unidentified developer and exempt it from Gatekeeper". Use the instructions on the wkhtmltopdf package file. Right click on the package file and follow the instructions.

## Settings for wkhtmltopdf

The `classic_blasts_to_pdf` app invokes `wkhtmltopdf` with these settings.

### PDF settings

| Key         | Values     |
| ----------- | ---------- |
| PageSize    | U.S. Legal |
| Orientation | Portrait   |
| Grayscale   | false      |

### Page Settings

| Key                       | Values |
| ------------------------- | ------ |
| Disable Smart Shrinking   | true   |
| Load Error Handling       | ignore |
| Load Media Error Handling | ignore |
| Zoom                      | 0.9    |

    wkhtmltopdf --zoom 3 --page-size Letter --disable-smart-shrinking

-   `--zoom 3` shows of email blasts nicely
-   `--page-size letter` sets the page size to U.S. Letter format.  The default is A4.
-   `--disable-smart-shrinking` seems to help. Taking it away results in squashed content. Let me know if you find something that works better.

## Installing `classic_blasts_to_pdfs`.
These steps will install `classic_blasts_to_pdfs` as an executable in `~/go/bin`.

```bash
go get "github.com/salsalabs/classic_blasts_to_pdfs"
go install
```
# Usage

    usage: classic_blasts_to_pdfs --login=LOGIN [<flags>]

    A command-line app to read email blasts, correct DIA URLs and write PDFs.

    Flags:
      --help         Show context-sensitive help (also try --help-long and --help-man).
      --login=LOGIN  YAML file with login credentials
      --count=10     Start this number of processors.
      --summary      Show blast dates, keys and subjects. Does not write PDFs.
      --htmlOnly     Write HTML. Does not write PDFs.```

    Use --help to get a list of options.

# Output

The application creates two directories.

-   `html`: the modified HTML for each of the blasts.
-   `blast_pdfs`: the PDFs for each of the blasts.
-   `blast_pdfs/[[year]]` PDFs for a particular year.  We generally export these as zip archives.

# Questions?  Comments?

Use the [Issues link](https://github.com/salsalabs/classic_blasts_to_pdfs/issues) in the repository.  Don't waste your time contacting Salsa support.
