/* Before you execute the program, Launch `cqlsh` and execute:
create keyspace example with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table example.tweet(timeline text, id UUID, text text, PRIMARY KEY(id));
create index on example.tweet(timeline);
*/
package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

type UeInfo struct {
	UeRespIp    string
	SgwReqIp    string
	UeId        int
	UeIdType    string
	EnbUeS1apId string
	Ecgi        string
	UeCap       string
}

/*
func encode_ue_info() {
	ue_struct := UeInfo{"One", "Two", 3, "Four", "Five", "Six", "Seven"}
	bookIn := UeInfo{
		Title:  "Void Moon",
		Author: "Michael Connelly",
		ISBN:   "316154067",
	}

	// gob encoding
	//key := []byte(bookIn.ISBN)
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(ue_struct)
	if err != nil {
		log.Fatal(err)
	}

	value := encBuf.Bytes()

	// store key, value, time passes, lookup value using key ...

	// gob decoding
	decBuf := bytes.NewBuffer(value)
	bookOut := Book{}
	err = gob.NewDecoder(decBuf).Decode(&bookOut)

	fmt.Println(string(key), bookOut)
}*/
func main() {
	// connect to the cluster
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "mme_faas"
	//cluster.Consistency = gocql.LocalSerial
	session, _ := cluster.CreateSession()
	defer session.Close()

	ue_struct := UeInfo{"OneOne", "Two2", 33, "Four4", "Five5", "Six6", "Seven7"}

	// gob encoding
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(ue_struct)
	if err != nil {
		log.Fatal(err)
	}
	value := encBuf.Bytes()
	var titleCAS int
	//var revidCAS int
	//var modifiedCAS time.Time
	id := 16
	var valueOutOld []byte
	applied, err := session.Query(`INSERT INTO mme_faas.ue_info (key, info) VALUES (?, ?) IF NOT EXISTS;`,
		id, value).ScanCAS(&titleCAS, &valueOutOld)
	if err != nil {
		log.Fatal(err)
	}
	//if applied {
	fmt.Println("Applied value:", applied, " title:", titleCAS)
	//}

	/*ue_struct := UeInfo{"One", "Two", 3, "Four", "Five", "Six", "Seven"}
	ue_intf := &ue_struct
	//ue_blob := []byte("test blob")

	// insert a tweet
	if err := session.Query(`INSERT INTO mme_faas.ue_info (key, info) VALUES (?, ?);`,
		11, ue_struct).Exec(); err != nil {
		log.Fatal(err)
	}*/

	/*var id int

	var valueOut []byte
	if err := session.Query(`SELECT * FROM mme_faas.ue_info WHERE key=?;`,
		13).Consistency(gocql.One).Scan(&id, &valueOut); err != nil {
		log.Fatal(err)
	}

	decBuf := bytes.NewBuffer(valueOut)
	infoOut := UeInfo{}
	err = gob.NewDecoder(decBuf).Decode(&infoOut)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("13", infoOut.UeRespIp)*/
}
