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
The tough part about all of this is installing the package in MacOS. If that's what you need, then you'll need to read [OSX: About Gatekeeper](https://support.apple.com/en-us/HT202491).

See the section named "How to open an app from a unidentified developer and exempt it from Gatekeeper". Use the instructions on the wkhtmltopdf package file. Right click on the package file and follow the instructions.

## Settings for wkhtmltopdf

The `classic_blasts_to_pdf` app configures `wkhtmltopdf` with these settings.

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

## Installing `classic_blasts_to_pdfs`.

use these steps to install `classic_blasts_to_pdfs` as an executable in `~/go/bin`.

```bash
go get "github.com/salsalabs/classic_blasts_to_pdfs"
go install
```


# Usage

```
usage: classic_blasts_to_pdfs --login=LOGIN [<flags>]

A command-line app to read email blasts, correct DIA URLs and write PDFs.

Flags:
  --help         Show context-sensitive help (also try --help-long and --help-man).
  --login=LOGIN  YAML file with login credentials
  --count=10     Start this number of processors.
  --summary      Show blast dates, keys and subjects. Do not write PDFs.
  --htmlOnly     Write HTML. Do not write PDFs.
  --apiVerbose   Display API calls and responses. Very noisy...
```

# Output

The application can create these outputs based on the options provided in the command line arguments.

-   `html/[[year]]`: (`--htmlOnly`) HTML for each of the blasts sent during "year".  No PDFs are created.
-   `blast_pdfs/[[year]]` (default) PDFs for each of the blast sent during "year" to HTML is generated.

Choosing `--summary` reads email blasts and writes their filenames to the console.

```
2019/11/04 12:51:47 Read 500 records from offset 0
2014-01-28 - 1278888 - FW: This is outrageous:.pdf
2014-01-28 - 1248540 - Sen. Johnson's vote cost HOW much?.pdf
2014-01-28 - 1256661 - URGENT: Need your name before tomorrow.pdf
2014-01-28 - 1247732 - Sen. Johnson's vote cost HOW much?.pdf
2014-01-28 - 1256243 - Sen. Petrowski needs to hear from you, [[First_Name]].pdf
2014-01-28 - 1267003 - Today, don't let them forget.pdf
2014-01-28 - 1174855 - Five times more likely to be murdered.pdf
2014-01-28 - 1253417 - They need to know:.pdf
2014-01-28 - 1266975 - Remember: Monday.pdf
```

# Delivery

We generally deliver blast PDFs to clients in ZIP archives, where each archive
contains the blast PDFs for a particular year. Here's a shell script that you
can use to create the zip files.

```bash
cd blast_pdfs
for i in *; do zip -r $i{.zip} $i; rm -rf $i; done
```
When the script is done, you should be able to list the directory contents

```bash
ls -al
```

and see a directory somewhat like this.

```
drwxr-xr-x  13 someuser  staff       416 Nov  4 12:39 .
drwxr-xr-x   5 someuser  staff       160 Nov  4 12:26 ..
-rw-r--r--   1 someuser  staff   3332693 Nov  4 12:39 2009.zip
-rw-r--r--   1 someuser  staff   3179658 Nov  4 12:39 2010.zip
-rw-r--r--   1 someuser  staff    343584 Nov  4 12:39 2011.zip
-rw-r--r--   1 someuser  staff   2127746 Nov  4 12:39 2012.zip
-rw-r--r--   1 someuser  staff    636265 Nov  4 12:39 2013.zip
-rw-r--r--   1 someuser  staff  37528142 Nov  4 12:39 2014.zip
-rw-r--r--   1 someuser  staff   9631910 Nov  4 12:39 2015.zip
-rw-r--r--   1 someuser  staff  16619817 Nov  4 12:39 2016.zip
-rw-r--r--   1 someuser  staff  12956176 Nov  4 12:39 2017.zip
-rw-r--r--   1 someuser  staff  17752595 Nov  4 12:39 2018.zip
-rw-r--r--   1 someuser  staff    245461 Nov  4 12:39 2019.zip
```

# Questions?  Comments?

Use the [Issues link](https://github.com/salsalabs/classic_blasts_to_pdfs/issues) in the repository.  Don't waste your time contacting Salsa support.
