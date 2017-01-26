# Abacaxi - The Library Metadata Hub
A tool to extract, transform and load metadata for eresources, esp. ebooks, in a DB, and match them with library records

Requires :

- [MongoDB](https://www.mongodb.com)
- the [Go programming language](https://golang.org/)
- a few extra Go libraries:
  - [bluemonday](https://github.com/microcosm-cc/bluemonday): `$ go get github.com/microcosm-cc/bluemonday`
  - [gorilla mux](http://www.gorillatoolkit.org/pkg/mux): `$ go get github.com/gorilla/mux`
  - [gorilla schema](http://www.gorillatoolkit.org/pkg/Schema): `$ go get github.com/gorilla/Schema`
  - [mgo.v2](https://godoc.org/gopkg.in/mgo.v2): `go get gopkg.in/mgo.v2`
  - [mgo.v2/bson](https://godoc.org/gopkg.in/mgo.v2/bson): `$ go get gopkg.in/mgo.v2/bson`

