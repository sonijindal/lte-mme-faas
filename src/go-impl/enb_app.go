//enb.go

// This is the UE/enb simulator
//export GOPATH=$GOPATH:/users/sonika05/lte-enb-mme
// go run enb.go <mme_ip> <num_ue> <async optional>
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
	"strings"
	"time"
)

var mme_ip string
var req_url string
var async int

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var time_arr []time.Time
var fd *os.File

func main() {
	args := os.Args[1:]
	fmt.Println(args[0])
	fmt.Println(args[1])
	mme_ip = args[0]
	num_ue, _ := strconv.Atoi(args[1])
	async = 0
	if len(args) == 3 {
		async, _ = strconv.Atoi(args[2])
	}
	Ue_info_arr = make([]Ue_info, num_ue)
	time_arr = make([]time.Time, num_ue)
	file := "./test/results/app" + strconv.Itoa(async) + "-" + time.Now().Format("2006.01.02-15:04:05")
	fmt.Println(file)
	var err error
	fd, err = os.Create(file)
	check(err)
	defer fd.Close()
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!Start Time: ", time.Now())
	if async == 1 {
		req_url = "http://128.110.154.116:31112/async-function/mme-faas-go"
		go send_request(num_ue)
		http.HandleFunc("/", process_async_msg)
		http.ListenAndServe("128.110.154.116:8001", nil)
	} else {
		req_url = "http://128.110.154.116:31112/function/mme-faas-go"
		go send_request(num_ue)
		fmt.Scanln()
	}
}
func process_async_msg(w http.ResponseWriter, req *http.Request) {

	b, _ := ioutil.ReadAll(req.Body)
	body := string(b[:])
	//fmt.Printf("Message from mme %s, len:%d\n", body, len(body))
	body = body[1 : len(body)-2]
	numbers := strings.Fields(body)
	input_len := len(numbers)
	//fmt.Println(numbers, input_len)
	body_bytes := make([]byte, input_len)
	for i := 0; i < input_len; i++ {
		in, _ := strconv.Atoi(numbers[i])
		body_bytes[i] = byte(in)
		//fmt.Println(numbers[i], body_bytes[i])
	}

	/*for i := 0; i < len(body); i++ {
		fmt.Printf("%c ", body[i])
	}
	fmt.Printf("\n")*/
	var msg Message_union_t
	err := json.Unmarshal(body_bytes, &msg)
	if err != nil {
		panic(err)
	}

	go func(msg Message_union_t) {
		switch msg.Msg_type {
		case AUTH_REQ:

			auth_req := msg.Auth_req
			id := auth_req.Enb_ue_s1ap_id
			//fmt.Printf("AUTH_REQ received, %d\n", id)
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
			//fmt.Printf("SEC_MODE_COMMAND received, %d\n", id)
			Ue_info_arr[id].Message.Msg_type = SEC_MODE_COMPLETE
			Ue_info_arr[id].Message.Sec_mode_complete.Mme_ue_s1ap_id =
				sec_mode_command.Mme_ue_s1ap_id
			Ue_info_arr[id].Message.Sec_mode_complete.Enb_ue_s1ap_id = id
			Ue_info_arr[id].Message.Sec_mode_complete.Tai = Ue_info_arr[id].Tai
			Ue_info_arr[id].Message.Sec_mode_complete.Plmn_id = Ue_info_arr[id].Plmn_id

			send_response(Ue_info_arr[id].Message)

		case ATTACH_ACCEPT:
			id := msg.Attach_accept.Enb_ue_s1ap_id
			elapsed := time.Since(time_arr[id])
			fd.WriteString(strconv.Itoa(int(id)) + ":" + elapsed.String() + "\n")
			//fmt.Printf("ATTACH_ACCEPT received, %d\n", msg.Attach_accept.Enb_ue_s1ap_id)

		default:
			fmt.Printf("Unhandled message(%s) received\n", msg.Msg_type)
			break

		}
	}(msg)
}
func process_msg(req *http.Response) {

	b, _ := ioutil.ReadAll(req.Body)
	body := string(b[:])
	//fmt.Printf("Message from mme %s, len:%d\n", body, len(body))
	body = body[1 : len(body)-2]
	numbers := strings.Fields(body)
	input_len := len(numbers)
	//fmt.Println(numbers, input_len)
	body_bytes := make([]byte, input_len)
	for i := 0; i < input_len; i++ {
		in, _ := strconv.Atoi(numbers[i])
		body_bytes[i] = byte(in)
		//fmt.Println(numbers[i], body_bytes[i])
	}

	/*for i := 0; i < len(body); i++ {
		fmt.Printf("%c ", body[i])
	}
	fmt.Printf("\n")*/
	var msg Message_union_t
	err := json.Unmarshal(body_bytes, &msg)
	if err != nil {
		fmt.Printf("Message from mme %s, len:%d\n", body, len(body))
		panic(err)
	}

	go func(msg Message_union_t) {
		switch msg.Msg_type {
		case AUTH_REQ:

			auth_req := msg.Auth_req
			id := auth_req.Enb_ue_s1ap_id
			//fmt.Printf("AUTH_REQ received, %d\n", id)
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
			//fmt.Printf("SEC_MODE_COMMAND received, %d\n", id)
			Ue_info_arr[id].Message.Msg_type = SEC_MODE_COMPLETE
			Ue_info_arr[id].Message.Sec_mode_complete.Mme_ue_s1ap_id =
				sec_mode_command.Mme_ue_s1ap_id
			Ue_info_arr[id].Message.Sec_mode_complete.Enb_ue_s1ap_id = id
			Ue_info_arr[id].Message.Sec_mode_complete.Tai = Ue_info_arr[id].Tai
			Ue_info_arr[id].Message.Sec_mode_complete.Plmn_id = Ue_info_arr[id].Plmn_id

			send_response(Ue_info_arr[id].Message)

		case ATTACH_ACCEPT:
			id := msg.Attach_accept.Enb_ue_s1ap_id
			elapsed := time.Since(time_arr[id])
			fd.WriteString(strconv.Itoa(int(id)) + ":" + elapsed.String() + "\n")
			//fmt.Printf("ATTACH_ACCEPT received, %d\n", msg.Attach_accept.Enb_ue_s1ap_id)
			fmt.Println("End: ", time.Now())

		default:
			fmt.Printf("Unhandled message(%s) received\n", msg.Msg_type)
			break

		}
	}(msg)
}

func send_request(num_ue int) {
	//req_url := "http://" + mme_ip + ":8002"
	for i := 0; i < num_ue; i++ {
		//fmt.Printf("Sending ATTACH_REQ:%d\n", i)
		Ue_info_arr[i].Message.Msg_type = ATTACH_REQ
		Ue_info_arr[i].Message.Attach_req.Imsi = strconv.Itoa(i)
		Ue_info_arr[i].Message.Attach_req.Enb_ue_s1ap_id = uint64(i)
		Ue_info_arr[i].Message.Attach_req.Plmn_id = 1
		Ue_info_arr[i].Message.Attach_req.Tai = 1
		Ue_info_arr[i].Message.Attach_req.Net_cap = 1
		time_arr[i] = time.Now()
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
	//req_url := "http://" + mme_ip + ":8002"

	form := new(bytes.Buffer)
	json.NewEncoder(form).Encode(msg)

	go func(req_url string, form io.Reader) {
		client := &http.Client{}
		req, _ := http.NewRequest("POST", req_url, form)
		if async == 1 {
			req.Header.Add("X-Callback-Url", "http://128.110.154.116:8001")
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Couldn't send req!")
			} else {
				fmt.Println(resp)
			}
		} else {
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				process_msg(resp)
			} else {
				fmt.Println("Couldn't send req!", err)
			}
			//resp, _ := http.Post(req_url, "application/json", form)
		}
	}(req_url, form)

}
