// mme.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	//	"handler/function/common"
	"math/rand"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var db *dynamodb.DynamoDB

func main() {
	lambda.Start(Handle)
}

func Handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	res := HandleReq([]byte(request.Body))
	return events.APIGatewayProxyResponse{
		Body:       string(res[:]),
		StatusCode: 200,
	}, nil
}

// Handle a serverless request
func HandleReq(req []byte) []byte {
	res := process_event(req)
	return res
}
func generate_mme_id() uint64 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return (uint64(r1.Uint32())<<32 + uint64(r1.Uint32()))
}
func process_event(req []byte) []byte {
	if db == nil {
		db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-2"))
	}
	var msg Message_union_t
	json.Unmarshal(req, &msg)
	switch msg.Msg_type {
	case ATTACH_REQ:
		//hss_sgw_stub(AUTH_INFO_REQ) //ignore ret

		id := generate_mme_id()
		ue_info := Ue_info{Ue_id: msg.Attach_req.Imsi, Tai: msg.Attach_req.Tai, Enb_ue_s1ap_id: msg.Attach_req.Enb_ue_s1ap_id,
			Ue_state: IDLE, Plmn_id: msg.Attach_req.Plmn_id, Mme_ue_s1ap_id: id}
		auth_req := build_auth_request(id, ue_info)
		for insert(id, ue_info, db) != true {
			id = generate_mme_id()
			ue_info.Mme_ue_s1ap_id = id
			ue_info.Message.Auth_req.Mme_ue_s1ap_id = id
		}
		var msg Message_union_t
		err := json.Unmarshal(auth_req, &msg)
		if err != nil {
			panic(err)
		}
		//fmt.Printf("Sending response of size:%d\n", len(auth_req))
		//fmt.Printf("ATTACH_REQ received, %d, mme id:%d\n", msg.Attach_req.Enb_ue_s1ap_id, id)
		return auth_req

	case AUTH_RES:
		//generate nas keys

		id := msg.Auth_res.Mme_ue_s1ap_id
		//fmt.Printf("AUTH_RES received, %d, mme_ue_id:%d\n", msg.Auth_res.Enb_ue_s1ap_id, id)
		ue_info, err := get(id, db)
		if err != nil {
			fmt.Println("Couldnt retrieve ue id:", id, err.Error())
			break
		}

		sec_mode_command := build_sec_mode_command(id, ue_info, db)
		return sec_mode_command

	case SEC_MODE_COMPLETE:
		//hss_sgw_stub(UPDATE_LOC_REQ)     //ignore ret
		//hss_sgw_stub(CREATE_SESSION_REQ) //ignore ret

		id := msg.Sec_mode_complete.Mme_ue_s1ap_id
		//fmt.Printf("SEC_MODE_COMPLETE received, %d, mme_ue_id:%d\n", msg.Sec_mode_complete.Enb_ue_s1ap_id, id)
		ue_info, err := get(id, db)
		if err != nil {
			fmt.Println("Couldnt retrieve ue id:", id, err.Error())
			break
		}
		attach_accept := build_attach_accept(id, ue_info, db)

		return attach_accept

	default:
		//hss_sgw_stub(MODIFY_BEARER_REQ) //ignore ret
		fmt.Printf("MME Unhandled message(%d) received\n", msg.Msg_type)
		msg.Msg_type = INVALID
		msg_str, err := json.Marshal(&msg)
		if err != nil {
			fmt.Println("Couldnt Marshal INVALID message")
			break
		}
		return []byte(msg_str)
	}

	msg.Msg_type = ERROR
	msg_str, err := json.Marshal(&msg)
	if err != nil {
		fmt.Println("Couldnt Marshal ERROR message")
	}
	return []byte(msg_str)
}
func hss_sgw_stub(msg_type S1ap_message_t) int {
	//If remote hss_sgw_stub, create tcp connection. The response can
	//sync or async
	conn, _ := net.Dial("tcp", "128.110.154.116:8081")
	text := "STB_MSG"
	fmt.Fprintf(conn, text+"\n")
	// Remove this for async response from stub
	_, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error in reading response from hss_sgw_stub")
		return 1
	}
	return 0
	//TODO: Add appropriate sleep for each message type
	// switch {
	// case auth_info_req:
	// 	return auth_info_answer

	// case update_location_req:
	// 	return update_location_answer

	// case create_session_req:
	// 	return create_session_res

	// case modify_bearer_req:
	// 	return modify_bearer_res
	// }
}
func build_auth_request(id uint64, ue_info Ue_info) []byte {
	ue_info.Message.Msg_type = AUTH_REQ
	ue_info.Message.Auth_req.Mme_ue_s1ap_id = ue_info.Mme_ue_s1ap_id
	ue_info.Message.Auth_req.Enb_ue_s1ap_id = ue_info.Enb_ue_s1ap_id
	ue_info.Message.Auth_req.Auth_challenge = 0xaa
	msg_str, err := json.Marshal(&ue_info.Message)
	if err != nil {
		fmt.Println("Couldnt Marshal AUTH_REQ message")
	}
	return msg_str
}

func build_sec_mode_command(id uint64, ue_info Ue_info, db *dynamodb.DynamoDB) []byte {

	ue_info.Message.Msg_type = SEC_MODE_COMMAND
	ue_info.Message.Sec_mode_command.Mme_ue_s1ap_id = ue_info.Mme_ue_s1ap_id
	ue_info.Message.Sec_mode_command.Enb_ue_s1ap_id = ue_info.Enb_ue_s1ap_id
	ue_info.Message.Sec_mode_command.Sec_algo = 0xaa
	msg_str, err := json.Marshal(&ue_info.Message)
	if err != nil {
		fmt.Println("Couldnt Marshal SEC_MODE_COMMAND message")
	} else {
		update(id, ue_info, db)
	}
	return msg_str
}

func build_attach_accept(id uint64, ue_info Ue_info, db *dynamodb.DynamoDB) []byte {
	ue_info.Message.Msg_type = ATTACH_ACCEPT
	ue_info.Message.Attach_accept.Mme_ue_s1ap_id = ue_info.Mme_ue_s1ap_id
	ue_info.Message.Attach_accept.Enb_ue_s1ap_id = ue_info.Enb_ue_s1ap_id
	ue_info.Message.Attach_accept.Ambr = 100
	ue_info.Message.Attach_accept.Sec_cap = 100
	msg_str, err := json.Marshal(&ue_info.Message)
	if err != nil {
		fmt.Println("Couldnt Marshal ATTACH_ACCEPT message")
	} else {
		update(id, ue_info, db)
	}
	return msg_str
}
