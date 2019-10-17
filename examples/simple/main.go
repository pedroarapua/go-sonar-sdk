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
	client := sonar.NewClient(url, tp.Client())
	options := sonar.ProjectsOptParams{
		Page:     1,
		Size:     10,
		Projects: []string{},
	}

	_, _, err := client.Projects.List(&options)

	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
	}
}
