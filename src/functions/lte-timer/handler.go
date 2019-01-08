package function

import (
	"fmt"
	"strconv"
	"time"
)

// Handle a serverless request
func Handle(req []byte) string {
	timer_val, _ := strconv.Atoi(string(req))
	fmt.Println("Going to start timer for:", timer_val)
	time.Sleep(time.Duration(timer_val) * time.Second)
	return "success"
}
