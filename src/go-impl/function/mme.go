// mme.go
package function

import (
	"encoding/json"
	"fmt"
	. "go-impl/common"
)

// Handle a serverless request
func Handle(req []byte) []byte {
	//return fmt.Sprintf("Hello, Go. You said: %s", string(req))
	res := process_event(req)
	return res
}

func process_event(req []byte) []byte {
	var msg Message_union_t
	json.Unmarshal(req, &msg)
	switch msg.Msg_type {
	case ATTACH_REQ:
		//ret = hss-sgw-stub (AUTH_INFO_REQ) //ignore ret

		ue_info := Ue_info{Ue_id: msg.Attach_req.Imsi, Tai: msg.Attach_req.Tai, Enb_ue_s1ap_id: msg.Attach_req.Enb_ue_s1ap_id,
			Ue_state: IDLE, Plmn_id: msg.Attach_req.Plmn_id}

		Ue_info_arr = append(Ue_info_arr, ue_info)
		id := len(Ue_info_arr) - 1
		Ue_info_arr[id].Enb_ue_s1ap_id = msg.Attach_req.Enb_ue_s1ap_id
		Ue_info_arr[id].Mme_ue_s1ap_id = uint64(id)
		fmt.Printf("ATTACH_REQ received, %d, mme id:%d\n", msg.Attach_req.Enb_ue_s1ap_id, id)
		auth_req := build_auth_request(uint64(id))
		return auth_req

	case AUTH_RES:
		//generate nas keys
		fmt.Printf("AUTH_RES received, %d\n", msg.Auth_res.Enb_ue_s1ap_id)
		id := msg.Auth_res.Mme_ue_s1ap_id
		sec_mode_command := build_sec_mode_command(id)
		return sec_mode_command

	case SEC_MODE_COMPLETE:
		//ret = hss_sgw_stub(update_location_req) //ignore ret
		//ret = hss_sgw_stub(create_session_req)  //ignore ret
		fmt.Printf("SEC_MODE_COMPLETE received, %d\n", msg.Sec_mode_complete.Enb_ue_s1ap_id)
		id := msg.Sec_mode_complete.Mme_ue_s1ap_id
		attach_accept := build_attach_accept(id)
		return attach_accept

	default:
		//ret = hss_sgw_stub(modify_bearer_req) //ignore ret
		fmt.Printf("Unhandled message(%d) received\n", msg.Msg_type)
		msg.Msg_type = INVALID
		msg_str, _ := json.Marshal(&msg)
		return []byte(msg_str)
	}
}

func build_auth_request(id uint64) []byte {
	Ue_info_arr[id].Message.Msg_type = AUTH_REQ
	Ue_info_arr[id].Message.Auth_req.Mme_ue_s1ap_id = Ue_info_arr[id].Mme_ue_s1ap_id
	Ue_info_arr[id].Message.Auth_req.Enb_ue_s1ap_id = Ue_info_arr[id].Enb_ue_s1ap_id
	Ue_info_arr[id].Message.Auth_req.Auth_challenge = 0xaa
	msg_str, _ := json.Marshal(&Ue_info_arr[id].Message)
	return msg_str
}

func build_sec_mode_command(id uint64) []byte {
	Ue_info_arr[id].Message.Msg_type = SEC_MODE_COMMAND
	Ue_info_arr[id].Message.Sec_mode_command.Mme_ue_s1ap_id = Ue_info_arr[id].Mme_ue_s1ap_id
	Ue_info_arr[id].Message.Sec_mode_command.Enb_ue_s1ap_id = Ue_info_arr[id].Enb_ue_s1ap_id
	Ue_info_arr[id].Message.Sec_mode_command.Sec_algo = 0xaa
	msg_str, _ := json.Marshal(&Ue_info_arr[id].Message)
	return msg_str
}

func build_attach_accept(id uint64) []byte {
	Ue_info_arr[id].Message.Msg_type = ATTACH_ACCEPT
	Ue_info_arr[id].Message.Attach_accept.Mme_ue_s1ap_id = Ue_info_arr[id].Mme_ue_s1ap_id
	Ue_info_arr[id].Message.Attach_accept.Enb_ue_s1ap_id = Ue_info_arr[id].Enb_ue_s1ap_id
	Ue_info_arr[id].Message.Attach_accept.Ambr = 100
	Ue_info_arr[id].Message.Attach_accept.Sec_cap = 100
	msg_str, _ := json.Marshal(&Ue_info_arr[id].Message)
	return msg_str
}
