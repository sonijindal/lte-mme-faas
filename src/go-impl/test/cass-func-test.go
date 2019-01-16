//enb.go

// This is the UE/enb simulator
//export GOPATH=$GOPATH:/users/sonika05/lte-enb-mme
// go run enb.go <mme_ip> <num_ue> <async optional>
package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io"
	"net/http"
	"os"
	"time"
)

var req_url string

var time_arr []time.Time
var fd *os.File

func main() {
	num_ue := 10000
	req_url = "http://127.0.0.1:31112/async-function/cass-func-go"
	go send_request(num_ue)
	fmt.Scanln()
}
func send_request(num_ue int) {
	//req_url := "http://" + mme_ip + ":8002"
	for i := 0; i < num_ue; i++ {
		send_response(i, "INSERT")
	}
}

func send_response(i int, msg string) {
	form := new(bytes.Buffer)
	json.NewEncoder(form).Encode(msg)

	go func(i int, req_url string, form io.Reader) {
		client := &http.Client{}
		req, _ := http.NewRequest("POST", req_url, form)
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			//fmt.Println(resp.Body)
		} else {
			fmt.Println("Message Failed:", i, err)
		}
	}(i, req_url, form)
}
