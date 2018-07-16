# fix_awhp_blasts
Go program to read blasts from AWHP and change any urls containing "`org2.democracyinaction.org`" to "`org2.salsalabs.com`".
# Background
Salsa used to have a domain named `democracyinaction.org`.  That was turned down in favor of using the common `salsalabs.com` domain.

Clients that uploaded and used images and files when `democracyinaction.org` was alive still have email blasts that reference that domain.
Salsa offsers a service that retrieves PDFs for completed email blasts. This service breaks down if image and file URLs are located on
`democracyinaction.org`.

This app solves that problem by reading all email blasts.  The ones that contain "democracyinaction.org" are modified to replace that with
"salsalabs.com".  The blasts are written back to the database with the change in place.

#Installation
```bash
go get "github.com/salsalabs/godig"

go get "github.com/salsalabs/fix_awhp_blasts"

go install

#Usage
```bash
go run main.go --credentials YAML_Credentials_File

#Questions?  Comments?
Use the [Issues link](https://github.com/salsalabs/fix_awhp_blasts/issues) in the repository.  Don't waste your time by contacting Salsa support.
