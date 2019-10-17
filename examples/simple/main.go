// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The simple command demonstrates a simple functionality which
// prompts the user for a GitHub username and lists all the public
// organization memberships of the specified username.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pedroarapua/go-sonar-sdk/sonar"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	fmt.Print("Sonar Token: ")
	token, _ := r.ReadString('\n')

	tp := sonar.BasicAuthTransport{
		Username: strings.TrimSpace(token),
	}

	url := string("http://localhost:9000/api/")
	client := sonar.NewClient(tp.Client(), url)
	options := sonar.NewProjectsListOptions(0, 0, nil)
	ctx := context.Background()

	resp, _, err := client.Projects.List(ctx, &options)

	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
	}
}
