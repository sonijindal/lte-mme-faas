// ue.go
package main

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
	ERROR
)
