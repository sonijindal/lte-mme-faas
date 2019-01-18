// dynamo_access.go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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
func generate_mme_id() uint64 {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return (uint64(r1.Uint32())<<32 + uint64(r1.Uint32()))
}

func Handle(request events.APIGatewayProxyRequest) (string, error) {
	/*id := generate_mme_id()

	var ue_info Ue_info

	for insert(id, ue_info) != true {

		id = generate_mme_id()
	}*/
	var ue_info Ue_info
	id := uint64(86939549076244585)
	start := time.Now()
	update(id, ue_info)
	elapsed := time.Since(start)
	fmt.Println("Time for update:", elapsed)
	start = time.Now()
	_ = get(id)
	elapsed = time.Since(start)
	fmt.Println("Time for read:", elapsed)
	//res, err := getItem(4705687789088032224)

	return "Inserted", nil
}

func main() {
	lambda.Start(Handle)
}
