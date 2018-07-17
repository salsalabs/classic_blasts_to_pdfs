# fix_awhp_blasts
Go program to read blasts from AWHP and change any urls containing "`org2.democracyinaction.org`" to "`org2.salsalabs.com`".
# Background
Salsa used to have a domain named `democracyinaction.org`.  That was turned down in favor of using  `salsalabs.com`.

Clients that uploaded and used images and files when `democracyinaction.org` was alive still have email blasts that reference that domain.

Salsa offsers a service that retrieves PDFs for completed email blasts. This service breaks down if image and file URLs are located on
`democracyinaction.org`.

This app solves that problem by reading all email blasts.  The ones that contain "democracyinaction.org" are modified to replace that with
"salsalabs.com".  The blasts are written back to the database with the change in place.

## Login credentials

The `fix_awhp_blasts` application looks for your login credentials in a YAML file.  You provide the filename as part of the execution.

You can read up on YAML and its formatting rules [here](https://en.wikipedia.org/wiki/YAML) if you'd like.

  The easiest way to get started is to  copy the `sample_login.yaml` file and edit it.  Here's an example.
```yaml
host: wfc2.wiredforchange.com
email: chuck@echeese.bizi
password: extra-super-secret-password!
```
The `email` and `password` are the ones that you normally use to log in. The `host` can be found by using [this page](https://help.salsalabs.com/hc/en-us/articles/115000341773-Salsa-Application-Program-Interface-API-#api_host) in Salsa's documentation.

Save the new login YAML file to disk.  We'll need it when we  run the `fix_awhp_blasts` app.

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
