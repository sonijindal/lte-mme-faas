package main

import (
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

type book struct {
    ISBN   string `json:"isbn"`
    Title  string `json:"title"`
    Author string `json:"author"`
}

func show() (events.APIGatewayProxyResponse, error) {
/*    bk := &book{
        ISBN:   "978-1420931693",
        Title:  "The Republic",
        Author: "Plato",
    }
*/	return events.APIGatewayProxyResponse{
		Body:       "Hello Books",
		StatusCode: 200,
	}, nil
}
func main() {
    lambda.Start(show)
}
