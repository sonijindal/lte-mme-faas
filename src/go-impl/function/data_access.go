// data_access.go
package function

import (
	"bytes"
	"encoding/gob"
	"fmt"
	. "go-impl/common"
	"log"
)

func insert(id int, ue_info Ue_info) bool {
	var idOld int
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

func get(id int) Ue_info {
	var valueOut []byte
	if err := session.Query(`SELECT * FROM mme_faas.ue_info WHERE key=?;`,
		id).Consistency(gocql.One).Scan(&id, &valueOut); err != nil {
		log.Fatal(err)
	}

	decBuf := bytes.NewBuffer(valueOut)
	infoOut := Ue_Info{}
	err = gob.NewDecoder(decBuf).Decode(&infoOut)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id, infoOut.UeRespIp)
	return infoOut
}

func exists(id int) bool {

}
