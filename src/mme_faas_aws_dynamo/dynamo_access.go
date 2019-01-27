// dynamo_access.go
package main

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Item struct {
	Key  uint64  `json:"key"`
	Info Ue_info `json:"info"`
}

// Declare a new DynamoDB instance. Note that this is safe for concurrent
// use.
// var db *DynamoDB
// = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-2"))

func getItem(key int, db *dynamodb.DynamoDB) (*Ue_info, error) {
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

	item := new(Ue_info)

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	return item, err
}
func insert(id uint64, ue_info Ue_info, db *dynamodb.DynamoDB) bool {
	item := new(Item)
	item.Key = id
	item.Info = ue_info
	av, err := dynamodbattribute.MarshalMap(item)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("ue_info"),
	}
	if err != nil {
		fmt.Println("MarshalMap failure")
		fmt.Println(err.Error())
		return false
	}
	_, err = db.PutItem(input)

	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		return false
	}

	return true
}

func get(id uint64, db *dynamodb.DynamoDB) (Ue_info, error) {
	//info := Ue_info{}
	item := Item{}
	var err error
	err = nil
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
	result, err1 := db.GetItem(input)
	if err1 != nil {
		fmt.Println("Got error calling GetItem:")
		fmt.Println(err1.Error())
		return item.Info, err1
	}
	if result.Item == nil {
		return item.Info, nil
	}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	return item.Info, err
}

type Ue_info_update struct {
	Info Ue_info `json:":in"`
}

func update(id uint64, ue_info Ue_info, db *dynamodb.DynamoDB) bool {
	//var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-2"))

	//ue_info.
	update := Ue_info_update{ue_info}

	update_info, err := dynamodbattribute.MarshalMap(update)
	if err != nil {
		fmt.Println("During update, error in MarshalMap:")
		fmt.Println(err.Error())
		return false
	}
	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"key": {
				N: aws.String(strconv.FormatUint(id, 10)),
			},
		},
		TableName: aws.String("ue_info"),

		ExpressionAttributeNames: map[string]*string{
			"#IN": aws.String("info"),
		},
		ExpressionAttributeValues: update_info,

		UpdateExpression: aws.String("SET #IN = :in"),
	}

	_, err = db.UpdateItem(input)

	if err != nil {
		fmt.Println("During update, got error calling UpdateItem:")
		fmt.Println(err.Error())
		return false
	}

	return true
}
