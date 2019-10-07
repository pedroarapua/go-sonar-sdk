// Copyright 2013 The techlead-metrics AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sonar

import (
	"context"
	"strings"
)

// ProjectsService provides access to the project related functions
// in the Sonar API.
//
// Sonar API docs: https://sonarcloud.io/web_api/api/projects
type ProjectsService service

// ResponseProjects object.
type ResponseProjects struct {
	Paging     ResponsePaging `json:"paging,omitempty"`
	Components []Project      `json:"components"`
}

// ResponsePaging object.
type ResponsePaging struct {
	Index int `json:"pageIndex"`
	Size  int `json:"pageSize"`
	Total int `json:"total"`
}

// Project represents a Sonar project.
type Project struct {
	Organization     *string `json:"organization,omitempty"`
	ID               *string `json:"id,omitempty"`
	Key              *string `json:"key,omitempty"`
	Name             *string `json:"name,omitempty"`
	Qualifier        *string `json:"qualifier,omitempty"`
	Visibility       *string `json:"visibility,omitempty"`
	LastAnalysisDate *string `json:"lastAnalysisDate,omitempty"`
	Revision         *string `json:"revision,omitempty"`
}

func (r ResponseProjects) String() string {
	return Stringify(r)
}

// ProjectsListOptions specifies the optional parameters to the
// ProjectsService.ListAll method.
type ProjectsListOptions struct {
	ListOptions
	Projects string `url:"projects,omitempty"`
}

// NewProjectsListOptions create new instance of ProjectsListOptions
func NewProjectsListOptions(index int, size int, projects []string) ProjectsListOptions {
	if index == 0 {
		index = 1
	}
	if size == 0 {
		size = 10
	}
	if projects == nil {
		projects = []string{}
	}

	return ProjectsListOptions{
		ListOptions: ListOptions{
			PageIndex: index,
			PageSize:  size,
		},
		Projects: strings.Join(projects[:], ","),
	}
}

// List lists all projects, in the order that they were created on sonar.
//
// Sonar API docs: https://sonarcloud.io/web_api/api/projects
func (p *ProjectsService) List(ctx context.Context, opt *ProjectsListOptions) (*ResponseProjects, *Response, error) {
	u, err := addOptions("projects/search", opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := p.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	response := &ResponseProjects{}
	httpResponse, err := p.client.Do(ctx, req, &response)
	if err != nil {
		return nil, httpResponse, err
	}
	return response, httpResponse, nil
}
