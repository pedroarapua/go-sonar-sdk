// Copyright 2013 The techlead-metrics AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sonar

import (
	"net/url"
	"strconv"
	"strings"
)

// ProjectsService provides access to the project related functions
// in the Sonar API.
//
// Sonar API docs: https://sonarcloud.io/web_api/api/projects
type ProjectsService struct {
	client *Client
}

// ResponseProjects object.
type ResponseProjects struct {
	Paging     *ResponsePaging `json:"paging,omitempty"`
	Components *[]Project      `json:"components"`
}

// ResponsePaging object.
type ResponsePaging struct {
	Index int `json:"pageIndex"`
	Size  int `json:"pageSize"`
	Total int `json:"total"`
}

// Project represents a Sonar project.
type Project struct {
	Organization     string `json:"organization,omitempty"`
	ID               string `json:"id,omitempty"`
	Key              string `json:"key,omitempty"`
	Name             string `json:"name,omitempty"`
	Qualifier        string `json:"qualifier,omitempty"`
	Visibility       string `json:"visibility,omitempty"`
	LastAnalysisDate string `json:"lastAnalysisDate,omitempty"`
	Revision         string `json:"revision,omitempty"`
}

// ProjectsOptParams specifies the optional parameters to the
type ProjectsOptParams struct {
	Page     int
	Size     int
	Projects []string
}

// List lists all projects, in the order that they were created on sonar.
//
// Sonar API docs: https://sonarcloud.io/web_api/api/projects
func (p *ProjectsService) List(opt *ProjectsOptParams) (*ResponseProjects, *string, error) {

	params := url.Values{}

	if opt != nil {
		if opt.Page != 0 {
			params.Add("p", strconv.Itoa(opt.Page))
		}
		if opt.Size != 0 {
			params.Add("ps", strconv.Itoa(opt.Size))
		}
		if opt.Projects != nil && len(opt.Projects) > 0 {
			params.Add("projects", strings.Join(opt.Projects[:], ","))
		}
	}

	req, err := p.client.NewRequest("GET", "projects/search", params.Encode())
	if err != nil {
		return nil, nil, err
	}

	response := &ResponseProjects{}
	resp, err := p.client.Do(req, &response)

	if err != nil {
		return nil, nil, err
	}

	return response, resp.BodyStrPtr, err
}
