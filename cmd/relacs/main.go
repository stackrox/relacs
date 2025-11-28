package main

import (
	"context"
	"fmt"
	"log"

	"github.com/stackrox/rel-acs/pkg/client"
)

var (
	org  = "stackrox"
	repo = "stackrox"
	tag  = "4.9.1"
)

func main() {
	client := client.NewClient()

	commit, err := client.GetCommitForTag(context.Background(), org, repo, tag)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(commit)
}
