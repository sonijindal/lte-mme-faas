// data_access.go
package function

import (
	"bytes"
	"encoding/gob"
	//	. "go-impl/common"
	"log"

	"github.com/gocql/gocql"
)

func insert(id uint64, ue_info Ue_info, session *gocql.Session) bool {
	var idOld uint64
	var valueOld []byte
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(ue_info)
	if err != nil {
		log.Fatal(err)
	}
	value := encBuf.Bytes()
	applied, err := session.Query(`INSERT INTO mme_faas.ue_info (key, info) VALUES (?, ?) IF NOT EXISTS;`,
		id, value).ScanCAS(&idOld, &valueOld)
	if err != nil {
		log.Fatal(err)
	}
	return applied
}

func get(id uint64, session *gocql.Session) Ue_info {
	var valueOut []byte
	if err := session.Query(`SELECT * FROM mme_faas.ue_info WHERE key=?;`,
		id).Consistency(gocql.One).Scan(&id, &valueOut); err != nil {
		log.Fatal(err)
	}

	decBuf := bytes.NewBuffer(valueOut)
	infoOut := Ue_info{}
	err := gob.NewDecoder(decBuf).Decode(&infoOut)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(id, infoOut)
	return infoOut
}

func update(id uint64, ue_info Ue_info, session *gocql.Session) bool {
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(ue_info)
	if err != nil {
		log.Fatal(err)
	}
	value := encBuf.Bytes()
	err = session.Query(`INSERT INTO mme_faas.ue_info (key, info) VALUES (?, ?);`,
		id, value).Exec()
	if err != nil {
		log.Fatal(err)
	}
	return true
}
