//enb.go

// This is the UE/enb simulator
//export GOPATH=$GOPATH:/users/sonika05/lte-enb-mme
// go run enb_app_aws_concur.go <num_ue> <concurrency>
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "go-impl/common"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var mme_ip string
var req_url string
var async int
var concurrency int
var client *http.Client
var completed int

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var time_arr []time.Time
var fd *os.File

var Free_workers chan int

func main() {
	args := os.Args[1:]
	completed = 0
	num_ue, _ := strconv.Atoi(args[0])
	concurrency, _ := strconv.Atoi(args[1])

	fmt.Println(num_ue)
	fmt.Println(concurrency)
	Ue_info_arr = make([]Ue_info, num_ue)
	time_arr = make([]time.Time, num_ue)
	file := "./test/results/enb_app_aws_cass-" + time.Now().Format("2006.01.02-15:04:05")
	fmt.Println(file)
	var err error
	fd, err = os.Create(file)
	check(err)
	defer fd.Close()
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!Start Time: ", time.Now())

	//req_url = "https://z560gr2qw3.execute-api.us-east-2.amazonaws.com/default/mme_faas_aws"
	req_url = "https://19obbklu86.execute-api.us-east-2.amazonaws.com/default/mme_faas_aws_dynamo"
	//num_ue_per_worker := num_ue / concurrency
	//start := 0
	client = &http.Client{}
	Free_workers = make(chan int, concurrency)

	log.Println("main start")
	for i := 0; i < num_ue; i++ {
		Free_workers <- 1
		go send_request(i)
	}
	fmt.Scanln()
}

func process_msg(req *http.Response, client *http.Client) {

	b, _ := ioutil.ReadAll(req.Body)
	body := string(b[:])

	var msg Message_union_t
	err := json.Unmarshal(b, &msg)
	if err != nil {
		fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!ERROR! Message from mme %s, len:%d\n", body, len(body))
		<-Free_workers
		return
	}
	//fmt.Println("Received Response")
	//go func(msg Message_union_t) {
	switch msg.Msg_type {
	case AUTH_REQ:

		auth_req := msg.Auth_req
		id := auth_req.Enb_ue_s1ap_id
		//fmt.Printf("AUTH_REQ received, %d\n", id)
		Ue_info_arr[id].Message.Msg_type = AUTH_RES
		Ue_info_arr[id].Message.Auth_res.Mme_ue_s1ap_id =
			auth_req.Mme_ue_s1ap_id
		//fmt.Printf("AUTH_REQ mme ue id, %d\n", auth_req.Mme_ue_s1ap_id)
		Ue_info_arr[id].Message.Auth_res.Enb_ue_s1ap_id = id
		Ue_info_arr[id].Message.Auth_res.Auth_challenge_answer =
			auth_req.Auth_challenge
		send_response(Ue_info_arr[id].Message, client)

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
		send_response(Ue_info_arr[id].Message, client)

	case ATTACH_ACCEPT:
		<-Free_workers
		id := msg.Attach_accept.Enb_ue_s1ap_id
		elapsed := time.Since(time_arr[id])
		fd.WriteString(strconv.Itoa(int(id)) + ":" + elapsed.String() + "\n")
		//fmt.Printf("ATTACH_ACCEPT received, %d\n", msg.Attach_accept.Enb_ue_s1ap_id)
		fmt.Println("End: ", time.Now(), " Completed:", completed)
		completed++

	default:
		fmt.Printf("ENB Unhandled message(%s) received\n", msg.Msg_type)
		<-Free_workers
		break

	}
	//}(msg)
}

func send_request(i int) {
	client := &http.Client{}
	//fmt.Printf("Sending ATTACH_REQ:%d\n", i)
	Ue_info_arr[i].Message.Msg_type = ATTACH_REQ
	Ue_info_arr[i].Message.Attach_req.Imsi = strconv.Itoa(i)
	Ue_info_arr[i].Message.Attach_req.Enb_ue_s1ap_id = uint64(i)
	Ue_info_arr[i].Message.Attach_req.Plmn_id = 1
	Ue_info_arr[i].Message.Attach_req.Tai = 1
	Ue_info_arr[i].Message.Attach_req.Net_cap = 1
	time_arr[i] = time.Now()
	send_response(Ue_info_arr[i].Message, client)
}

func send_response(msg Message_union_t, client *http.Client) {
	form := new(bytes.Buffer)
	json.NewEncoder(form).Encode(msg)

	req, _ := http.NewRequest("POST", req_url, form)

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		process_msg(resp, client)
	} else {
		fmt.Println("Couldn't send req!", err)
	}
}