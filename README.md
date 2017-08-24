# &#127821; Abacaxi - The Library Metadata Hub
A tool to extract, transform and load metadata for eresources, esp. ebooks, in a DB, and match them with library records

What it does :

- slurps files : publisher csv, kbart files, XML exports from the SFX OpenURL resolver
- dedupe records using the identifiers it has, e.g. isbn / issn / sfx id...
- try to get Unimarc records matching these identifiers
- export those records as unimarc records or kbart files

Requires :

- [MongoDB](https://www.mongodb.com)
- the [Go programming language](https://golang.org/)
- a few extra Go libraries:
  - [bcrypt](https://golang.org/x/crypto/bcrypt): `$ go get golang.org/x/crypto/bcrypt`
  - [bluemonday](https://github.com/microcosm-cc/bluemonday): `$ go get github.com/microcosm-cc/bluemonday`
  - [goisbn](https://github.com/terryh/goisbn): `$ go get -u github.com/terryh/goisbn`
  - [gorilla mux](http://www.gorillatoolkit.org/pkg/mux): `$ go get github.com/gorilla/mux`
  - [gorilla schema](http://www.gorillatoolkit.org/pkg/Schema): `$ go get github.com/gorilla/Schema`
  - [gorilla sessions](http://www.gorillatoolkit.org/pkg/Sessions): `$ go get github.com/gorilla/sessions`
  - [go sudoc](https://github.com/nicomo/gosudoc): `$ go get  github.com/nicomo/gosudoc`
  - [mgo.v2](https://godoc.org/gopkg.in/mgo.v2): `$ go get gopkg.in/mgo.v2`
  - [mgo.v2/bson](https://godoc.org/gopkg.in/mgo.v2/bson): `$ go get gopkg.in/mgo.v2/bson`

Before you start, fill in the config/config.json file : 

- hostname: "http://localhost:8080/" - hostname (and path) to the root, e.g. http://metadata.mylibrary.com/ - don't forget the trailing /
- mongodbhosts: "localhost:27017" - where is mongoDB, e.g. localhost:27017
- authdatabase: "abacaxidb" - name of the mongodb, e.g.  abacaxidb
- sessionstorekey: "long string of letters, numbers and signs", e.g. g9H4FJa+;y3G7$wyye
