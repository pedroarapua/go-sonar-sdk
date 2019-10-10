// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sonar

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestProjectsService_List(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/projects/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"paging": {
				"pageIndex": 1,
				"pageSize": 5,
				"total": 2
			},
			"components": [{
				"organization": "default-organization",
				"id": "AW1qZC7qcyZaoJOWxgwC",
				"key": "teste",
				"name": "teste",
				"qualifier": "TRK",
				"visibility": "public",
				"lastAnalysisDate": "2019-09-26T14:18:52+0000",
				"revision": "a37e0a446129120538489350be32f97aa7e5893f"
			}]
		}`)
	})

	response, _, err := client.Projects.List(context.Background(), nil)
	if err != nil {
		t.Errorf("Projects.List returned error: %v", err)
	}

	want := &ResponseProjects{
		Paging: &ResponsePaging{
			Index: Int(1),
			Size:  Int(5),
			Total: Int(2),
		},
		Components: &[]Project{{
			Organization:     String("default-organization"),
			ID:               String("AW1qZC7qcyZaoJOWxgwC"),
			Key:              String("teste"),
			Name:             String("teste"),
			Qualifier:        String("TRK"),
			Visibility:       String("public"),
			LastAnalysisDate: String("2019-09-26T14:18:52+0000"),
			Revision:         String("a37e0a446129120538489350be32f97aa7e5893f"),
		}},
	}
	if !reflect.DeepEqual(response, want) {
		t.Errorf("Projects.List returned %+v, want %+v", response, want)
	}
}
