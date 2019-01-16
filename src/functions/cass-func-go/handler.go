package function

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"time"

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

type Attach_req_t struct {
	Imsi           string
	Enb_ue_s1ap_id uint64
	Plmn_id        uint8
	Tai            uint8
	Net_cap        uint8
}
type Auth_req_t struct {
	Enb_ue_s1ap_id uint64
	Mme_ue_s1ap_id uint64
	Auth_challenge uint8
}
type Auth_res_t struct {
	Enb_ue_s1ap_id        uint64
	Mme_ue_s1ap_id        uint64
	Auth_challenge_answer uint8
}
type Sec_mode_command_t struct {
	Enb_ue_s1ap_id uint64
	Mme_ue_s1ap_id uint64
	Sec_algo       uint8
}
type Sec_mode_complete_t struct {
	Enb_ue_s1ap_id uint64
	Mme_ue_s1ap_id uint64
	Tai            uint8
	Plmn_id        uint8
}
type Attach_accept_t struct {
	Enb_ue_s1ap_id uint64
	Mme_ue_s1ap_id uint64
	Ambr           uint8
	Sec_cap        uint8
}

type Message_union_t struct {
	Msg_type          S1ap_message_t
	Attach_req        Attach_req_t
	Auth_req          Auth_req_t
	Auth_res          Auth_res_t
	Sec_mode_command  Sec_mode_command_t
	Sec_mode_complete Sec_mode_complete_t
	Attach_accept     Attach_accept_t
}

type Ue_info struct {
	Ue_id          string /*IMSI or GUTI*/
	Enb_ue_s1ap_id uint64
	Mme_ue_s1ap_id uint64
	Plmn_id        uint8
	Tai            uint8
	Message        Message_union_t
	Datalen        int
	Ue_state       Ue_state_t
}

var Ue_info_arr []Ue_info

type Ue_state_t int

const (
	IDLE      Ue_state_t = 0
	CONNECTED Ue_state_t = 1
)

type S1ap_message_t int

const (
	ATTACH_REQ S1ap_message_t = iota + 1
	AUTH_REQ
	AUTH_RES
	AUTH_INFO_REQ
	AUTH_INFO_ANS
	UPDATE_LOC_REQ
	UPDATE_LOC_ANS
	CREATE_SESSION_REQ
	CREATE_SESSION_RES
	MODIFY_BEARER_REQ
	MODIFY_BEARER_RES
	SEC_MODE_COMMAND
	SEC_MODE_COMPLETE
	ATTACH_ACCEPT
	ATTACH_COMPLETE
	ATTACH_ACCEPT_SENT_TIMER
	INVALID
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
func generate_mme_id() uint64 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return (uint64(r1.Uint32())<<32 + uint64(r1.Uint32()))
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}

var session *gocql.Session

// Handle a serverless request
func Handle(req []byte) string {
	if session == nil {
		//fmt.Println("Creating new connection")
		cluster := gocql.NewCluster("128.110.154.116")
		cluster.Keyspace = "mme_faas"
		var err error
		session, err = cluster.CreateSession()
		if err != nil {
			return fmt.Sprintf("Error in creating session")
		}
	} else {
		fmt.Println("Using old connection!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
	id := generate_mme_id()

	var ue_info Ue_info

	for insert(id, ue_info, session) != true {

		id = generate_mme_id()
	}
	return "Inserted"

}
