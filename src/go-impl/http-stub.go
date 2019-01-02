// http-stub.go

// This is the http serverless framework stub, to be removed when mme runs on openfaas
// go run http-stub.go <enb_ip>
package main

import (
	"bytes"
	"fmt"
	. "go-impl/function"
	"io/ioutil"
	"net/http"
	"os"
)

var enb_ip string

func main() {
	args := os.Args[1:]
	fmt.Println(args[0])
	enb_ip = args[0]
	http.HandleFunc("/", msg_from_enb)
	http.ListenAndServe(":8002", nil)
}

func msg_from_enb(w http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	//fmt.Printf("Message from enb: %s\n", body)
	resp := Handle(body)
	//fmt.Printf("Message for enb: %s\n", resp)
	send_response(resp)
}

/*func Handle(req []byte) []byte {
	return (req)
	//process_event()
	//return res
}*/

func send_response(resp []byte) {
	req_url := "http://" + enb_ip + ":8001/"
	//form := new(bytes.Buffer)
	//json.NewEncoder(form).Encode(resp)
	//fmt.Printf("Sending response to: %s\n", req_url)
	http.Post(req_url, "application/json", bytes.NewBuffer(resp))
}
