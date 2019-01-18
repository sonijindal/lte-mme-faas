// dynamo_access.go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	//"github.com/aws/aws-lambda-go/events"
	//"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

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

type Item struct {
	Key  uint64  `json:"key"`
	Info Ue_info `json:"info"`
}

// Declare a new DynamoDB instance. Note that this is safe for concurrent
// use.
var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-2"))

func getItem(key int) (*Ue_info, error) {
	// Prepare the input for the query.
	input := &dynamodb.GetItemInput{
		TableName: aws.String("ue_info"),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {
				N: aws.String(strconv.Itoa(key)),
			},
		},
	}

	// Retrieve the item from DynamoDB. If no matching item is found
	// return nil.
	result, err := db.GetItem(input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	// The result.Item object returned has the underlying type
	// map[string]*AttributeValue. We can use the UnmarshalMap helper
	// to parse this straight into the fields of a struct. Note:
	// UnmarshalListOfMaps also exists if you are working with multiple
	// items.
	/*bk := new(book)
	err = dynamodbattribute.UnmarshalMap(result.Item, bk)
	if err != nil {
		return nil, err
	}

	return bk, nil*/
	item := new(Ue_info)

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	return item, err
}
func insert(id uint64, ue_info Ue_info) bool {
	item := new(Item)
	item.Key = id
	item.Info = ue_info
	av, err := dynamodbattribute.MarshalMap(item)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("ue_info"),
	}

	_, err = db.PutItem(input)

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//fmt.Println("Successfully added")
	return true
}

func get(id uint64) *Ue_info {
	// Prepare the input for the query.
	input := &dynamodb.GetItemInput{
		TableName: aws.String("ue_info"),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {
				N: aws.String(strconv.FormatUint(id, 10)),
			},
		},
	}

	// Retrieve the item from DynamoDB. If no matching item is found
	// return nil.
	result, err := db.GetItem(input)
	if err != nil {
		return nil
	}
	if result.Item == nil {
		return nil
	}

	item := new(Ue_info)

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	return item
}

func update(id uint64, ue_info Ue_info) bool {
	item := new(Item)
	item.Key = id
	item.Info = ue_info
	av, err := dynamodbattribute.MarshalMap(item)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("ue_info"),
	}

	_, err = db.PutItem(input)

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//fmt.Println("Successfully updated")
	return true
}

/*func get(id uint64, session *gocql.Session) Ue_info {
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
}*/
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

func main() {
	// go run dynamo-test.go <type> <#times>
	args := os.Args[1:]
	fmt.Println(args[0])
	fmt.Println(args[1])
	test_type, _ := strconv.Atoi(args[0])
	times, _ := strconv.Atoi(args[1])
	//lambda.Start(Handle)
	file := "./results/dynamo" + args[0] + "-" + time.Now().Format("2006.01.02-15:04:05")
	fmt.Println(file)
	f, err := os.Create(file)
	check(err)
	defer f.Close()
	var ue_info Ue_info
	switch test_type {
	case 0: //INSERT
		total_insert_err := 0
		for i := 0; i < times; i++ {
			id := generate_mme_id()
			insert_err := 0
			start := time.Now()
			for insert(id, ue_info) != true {
				insert_err++
				id = generate_mme_id()
			}
			elapsed := time.Since(start)
			f.WriteString(elapsed.String() + "\n")
			total_insert_err = total_insert_err + insert_err
		}
		break
	case 1: //UPDATE
		for i := 0; i < times; i++ {
			id := uint64(86939549076244585)
			start := time.Now()
			update(id, ue_info)
			elapsed := time.Since(start)
			f.WriteString(elapsed.String() + "\n")
		}
		break
	case 2: //GET
		for i := 0; i < times; i++ {
			id := uint64(86939549076244585)
			start := time.Now()
			_ = get(id)
			elapsed := time.Since(start)
			f.WriteString(elapsed.String() + "\n")
		}
		break
	}
	fmt.Println("Test Done!")
}
