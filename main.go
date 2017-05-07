// Copyright © 2017 Jan Oliver Oelerich <janoliver@oelerich.org>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"github.com/andybalholm/milter"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"net/textproto"
	"time"
	"flag"
	"fmt"
)

var schema = `
CREATE TABLE IF NOT EXISTS smtp_record (
    mid TEXT NOT NULL PRIMARY KEY,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    contentType DATETIME NULL,
    is_dkim BOOLEAN NULL
);
`

type idMilter struct {
	db          *sqlx.DB
	mid         string
	from        string
	to          string
	time        time.Time
	contentType string
	isDKIM      bool
}

func (s idMilter) WriteToDBIfComplete() {
	if s.mid == "" || s.from == "" || s.to == "" {
		return
	}
	s.db.MustExec("INSERT OR REPLACE INTO smtp_record (mid, sender, recipient, time) VALUES ($1, $2, $3, $4)", s.mid, s.from, s.to, s.time)
	log.Printf("Wrote record (Message-Id: %s)", s.mid)
}

func (s idMilter) Connect(hostname string, network string, address string, macros map[string]string) milter.Response {
	return milter.Continue
}

func (s idMilter) Helo(name string, macros map[string]string) milter.Response {
	return milter.Continue
}

func (s idMilter) From(sender string, macros map[string]string) milter.Response {
	return milter.Continue
}

func (s idMilter) To(recipient string, macros map[string]string) milter.Response {
	return milter.Continue
}

func (s idMilter) Headers(h textproto.MIMEHeader) milter.Response {
	t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", h.Get("Date"))
	if err == nil {
		s.time = t
	}
	s.from = h.Get("From")
	s.to = h.Get("To")
	s.mid = h.Get("Message-Id")
	s.contentType = h.Get("Content-Type")
	s.isDKIM = h.Get("DKIM-Signature") != ""
	s.WriteToDBIfComplete()
	return milter.Continue
}

func (s idMilter) Body(body []byte, m milter.Modifier) milter.Response {
	return milter.Continue
}

func main() {
	port := flag.Int("port", 9929, "The port to listen on.")
	dbPath := flag.String("db-path", "gopstats.sqlite3", "The SQLite3 Database path")
	flag.Parse()

	db, err := sqlx.Connect("sqlite3", *dbPath)
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(schema)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalln("Couldn't start milter server! Something's wrong.")
	}
	milter.Serve(ln, func() milter.Milter {
		m := idMilter{}
		m.db = db
		return m
	})
}
