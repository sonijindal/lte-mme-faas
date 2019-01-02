//enb.go

// This is the UE/enb simulator
//export GOPATH=$GOPATH:/users/sonika05/lte-enb-mme
// go run enb.go <mme_ip> <num_ue>
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "go-impl/common"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

var mme_ip string

func main() {
	args := os.Args[1:]
	fmt.Println(args[0])
	fmt.Println(args[1])
	mme_ip = args[0]
	num_ue, _ := strconv.Atoi(args[1])
	Ue_info_arr = make([]Ue_info, num_ue)
	go send_request(num_ue)
	http.HandleFunc("/", process_msg)
	http.ListenAndServe(":8001", nil)
}

func process_msg(w http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	//fmt.Printf("Message from mme %s\n", body)

	var msg Message_union_t
	err := json.Unmarshal(body, &msg)
	if err != nil {
		panic(err)
	}
	//io.WriteString(w, "Success")
	go func(msg Message_union_t) {
		switch msg.Msg_type {
		case AUTH_REQ:

			auth_req := msg.Auth_req
			id := auth_req.Enb_ue_s1ap_id
			fmt.Printf("AUTH_REQ received, %d\n", id)
			Ue_info_arr[id].Message.Msg_type = AUTH_RES
			Ue_info_arr[id].Message.Auth_res.Mme_ue_s1ap_id =
				auth_req.Mme_ue_s1ap_id
			Ue_info_arr[id].Message.Auth_res.Enb_ue_s1ap_id = id
			Ue_info_arr[id].Message.Auth_res.Auth_challenge_answer =
				auth_req.Auth_challenge

			send_response(Ue_info_arr[id].Message)

		case SEC_MODE_COMMAND:

			sec_mode_command := msg.Sec_mode_command
			id := sec_mode_command.Enb_ue_s1ap_id
			fmt.Printf("SEC_MODE_COMMAND received, %d\n", id)
			Ue_info_arr[id].Message.Msg_type = SEC_MODE_COMPLETE
			Ue_info_arr[id].Message.Sec_mode_complete.Mme_ue_s1ap_id =
				sec_mode_command.Mme_ue_s1ap_id
			Ue_info_arr[id].Message.Sec_mode_complete.Enb_ue_s1ap_id = id
			Ue_info_arr[id].Message.Sec_mode_complete.Tai = Ue_info_arr[id].Tai
			Ue_info_arr[id].Message.Sec_mode_complete.Plmn_id = Ue_info_arr[id].Plmn_id

			send_response(Ue_info_arr[id].Message)

		case ATTACH_ACCEPT:
			fmt.Printf("ATTACH_ACCEPT received, %d\n", msg.Attach_accept.Enb_ue_s1ap_id)

		default:
			fmt.Printf("Unhandled message(%s) received\n", msg.Msg_type)
			break

		}
	}(msg)
}

func send_request(num_ue int) {
	//req_url := "http://" + mme_ip + ":8002"
	for i := 0; i < num_ue; i++ {
		fmt.Printf("Sending ATTACH_REQ:%d\n", i)
		Ue_info_arr[i].Message.Msg_type = ATTACH_REQ
		Ue_info_arr[i].Message.Attach_req.Imsi = strconv.Itoa(i)
		Ue_info_arr[i].Message.Attach_req.Enb_ue_s1ap_id = uint64(i)
		Ue_info_arr[i].Message.Attach_req.Plmn_id = 1
		Ue_info_arr[i].Message.Attach_req.Tai = 1
		Ue_info_arr[i].Message.Attach_req.Net_cap = 1
		send_response(Ue_info_arr[i].Message)
		//form := new(bytes.Buffer)
		// json.NewEncoder(form).Encode(msg)

		// go func(req_url string, form io.Reader) {
		// 	/*resp, err := */ http.Post(req_url, "application/json", form)
		// }(req_url, form)
		//time.Sleep(10 * time.Millisecond)
	}
}

func send_response(msg Message_union_t) {
	req_url := "http://" + mme_ip + ":8002"
	form := new(bytes.Buffer)
	json.NewEncoder(form).Encode(msg)

	go func(req_url string, form io.Reader) {
		/*resp, err := */ http.Post(req_url, "application/json", form)
	}(req_url, form)

}
