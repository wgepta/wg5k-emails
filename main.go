package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rbriski/wg5k/racemine"
)

func main() {
	rc := racemine.NewClient(os.Getenv("RM_USERNAME"), os.Getenv("RM_PASSWORD"))
	links, err := rc.GetAllExports()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	fmt.Printf("%v\n", links)
}

// func main() {
// 	apiKey := os.Getenv("CC_API_KEY")
// 	accessToken := os.Getenv("CC_ACCESS_TOKEN")
// 	context := context.Background()

// 	cc := constantcontact.NewClient(nil, apiKey, accessToken)
// 	lists, _, err := cc.Lists.GetAll(context)
// 	if err != nil {
// 		log.Errorf("%v\n", err)
// 	}
// 	for _, list := range lists {
// 		fmt.Printf("%s [%s]: %d\n", list.Name, list.Status, list.ContactCount)
// 	}
// }
