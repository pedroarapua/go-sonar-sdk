// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sonar

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestProjectsService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/projects/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"paging": {
				"pageIndex": 1,
				"pageSize": 10,
				"total": 1
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

	response, _, err := client.Projects.List(nil)
	if err != nil {
		t.Errorf("Projects.List returned error: %v", err)
	}

	want := &ResponseProjects{
		Paging: &ResponsePaging{
			Index: 1,
			Size:  10,
			Total: 1,
		},
		Components: &[]Project{{
			Organization:     "default-organization",
			ID:               "AW1qZC7qcyZaoJOWxgwC",
			Key:              "teste",
			Name:             "teste",
			Qualifier:        "TRK",
			Visibility:       "public",
			LastAnalysisDate: "2019-09-26T14:18:52+0000",
			Revision:         "a37e0a446129120538489350be32f97aa7e5893f",
		}},
	}
	if !reflect.DeepEqual(response, want) {
		t.Errorf("Projects.List returned %+v, want %+v", response, want)
	}
}

func TestProjectsService_ListEmpty(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/projects/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"paging": {
				"pageIndex": 1,
				"pageSize": 10,
				"total": 0
			},
			"components": []
		}`)
	})

	response, _, err := client.Projects.List(nil)
	if err != nil {
		t.Errorf("Projects.List returned error: %v", err)
	}

	want := &ResponseProjects{
		Paging: &ResponsePaging{
			Index: 1,
			Size:  10,
			Total: 0,
		},
		Components: &[]Project{},
	}
	if !reflect.DeepEqual(response, want) {
		t.Errorf("Projects.List returned %+v, want %+v", response, want)
	}
}
