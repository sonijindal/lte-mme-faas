// mme.go
package function

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	. "go-impl/common"

	"github.com/gocql/gocql"
)

var session *gocql.Session

// Handle a serverless request
func Handle(req []byte) []byte {

	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "mme_faas"
	session, _ = cluster.CreateSession()

	res := process_event(req)
	//Enabling this Close, disconnects the connection randomly, so just comment it for now.
	//session.Close()
	return res
}
func generate_mme_id() uint64 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return (uint64(r1.Uint32())<<32 + uint64(r1.Uint32()))
}
func process_event(req []byte) []byte {
	var msg Message_union_t
	json.Unmarshal(req, &msg)
	switch msg.Msg_type {
	case ATTACH_REQ:
		//ret = hss-sgw-stub (AUTH_INFO_REQ) //ignore ret

		id := generate_mme_id()
		ue_info := Ue_info{Ue_id: msg.Attach_req.Imsi, Tai: msg.Attach_req.Tai, Enb_ue_s1ap_id: msg.Attach_req.Enb_ue_s1ap_id,
			Ue_state: IDLE, Plmn_id: msg.Attach_req.Plmn_id, Mme_ue_s1ap_id: id}
		auth_req := build_auth_request(id, ue_info)
		for insert(id, ue_info, session) != true {
			id = generate_mme_id()
			ue_info.Mme_ue_s1ap_id = id
			ue_info.Message.Auth_req.Mme_ue_s1ap_id = id
		}

		fmt.Printf("ATTACH_REQ received, %d, mme id:%d\n", msg.Attach_req.Enb_ue_s1ap_id, id)
		return auth_req

	case AUTH_RES:
		//generate nas keys
		fmt.Printf("AUTH_RES received, %d\n", msg.Auth_res.Enb_ue_s1ap_id)
		id := msg.Auth_res.Mme_ue_s1ap_id
		ue_info := get(id, session)
		sec_mode_command := build_sec_mode_command(id, ue_info)
		return sec_mode_command

	case SEC_MODE_COMPLETE:
		//ret = hss_sgw_stub(update_location_req) //ignore ret
		//ret = hss_sgw_stub(create_session_req)  //ignore ret
		fmt.Printf("SEC_MODE_COMPLETE received, %d\n", msg.Sec_mode_complete.Enb_ue_s1ap_id)
		id := msg.Sec_mode_complete.Mme_ue_s1ap_id
		ue_info := get(id, session)
		attach_accept := build_attach_accept(id, ue_info)
		return attach_accept

	default:
		//ret = hss_sgw_stub(modify_bearer_req) //ignore ret
		fmt.Printf("Unhandled message(%d) received\n", msg.Msg_type)
		msg.Msg_type = INVALID
		msg_str, _ := json.Marshal(&msg)
		return []byte(msg_str)
	}
}

func build_auth_request(id uint64, ue_info Ue_info) []byte {
	ue_info.Message.Msg_type = AUTH_REQ
	ue_info.Message.Auth_req.Mme_ue_s1ap_id = ue_info.Mme_ue_s1ap_id
	ue_info.Message.Auth_req.Enb_ue_s1ap_id = ue_info.Enb_ue_s1ap_id
	ue_info.Message.Auth_req.Auth_challenge = 0xaa
	msg_str, _ := json.Marshal(&ue_info.Message)
	return msg_str
}

func build_sec_mode_command(id uint64, ue_info Ue_info) []byte {
	ue_info.Message.Msg_type = SEC_MODE_COMMAND
	ue_info.Message.Sec_mode_command.Mme_ue_s1ap_id = ue_info.Mme_ue_s1ap_id
	ue_info.Message.Sec_mode_command.Enb_ue_s1ap_id = ue_info.Enb_ue_s1ap_id
	ue_info.Message.Sec_mode_command.Sec_algo = 0xaa
	msg_str, _ := json.Marshal(&ue_info.Message)
	update(id, ue_info, session)
	return msg_str
}

func build_attach_accept(id uint64, ue_info Ue_info) []byte {
	ue_info.Message.Msg_type = ATTACH_ACCEPT
	ue_info.Message.Attach_accept.Mme_ue_s1ap_id = ue_info.Mme_ue_s1ap_id
	ue_info.Message.Attach_accept.Enb_ue_s1ap_id = ue_info.Enb_ue_s1ap_id
	ue_info.Message.Attach_accept.Ambr = 100
	ue_info.Message.Attach_accept.Sec_cap = 100
	msg_str, _ := json.Marshal(&ue_info.Message)
	update(id, ue_info, session)
	return msg_str
}
