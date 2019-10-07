// Copyright 2013 The techlead-metrics AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sonar

import (
	"context"
	"time"
)

// ProjectsService provides access to the project related functions
// in the Sonar API.
//
// Sonar API docs: https://sonarcloud.io/web_api/api/projects
type ProjectsService service

// Project represents a Sonar project.
type Project struct {
	Organization     *string    `json:"organization,omitempty"`
	ID               *string    `json:"id,omitempty"`
	Key              *string    `json:"key,omitempty"`
	Name             *string    `json:"name,omitempty"`
	Qualifier        *string    `json:"qualifier,omitempty"`
	Visibility       *string    `json:"visibility,omitempty"`
	LastAnalysisDate *time.Time `json:"lastAnalysisDate,omitempty"`
	Revision         *string    `json:"revision,omitempty"`
}

func (p Project) String() string {
	return Stringify(p)
}

// ProjectsListOptions specifies the optional parameters to the
// ProjectsService.ListAll method.
type ProjectsListOptions struct {
	ListOptions
}

// ListAll lists all projects, in the order that they were created on sonar.
//
// Sonar API docs: https://sonarcloud.io/web_api/api/projects
func (p *ProjectsService) ListAll(ctx context.Context, opt *ProjectsListOptions) ([]*Project, *Response, error) {
	u, err := addOptions("projects", opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := p.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	projects := []*Project{}
	resp, err := p.client.Do(ctx, req, &projects)
	if err != nil {
		return nil, resp, err
	}
	return projects, resp, nil
}
