package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type JiraResponse struct {
	Issues []JiraIssue `json:"issues"`
}

type JiraIssue struct {
	Id         string `json:"id"`
	Key        string `json:"key"`
	JiraFields `json:"fields"`
}

type JiraFields struct {
	Summary string `json:"summary"`
}

type Jira struct {
	Username string `config:"username,optional"`
	Token    string `config:"api-token,optional"`
}

func (jira *Jira) BasicAuth() string {
	out := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, out)
	fmt.Fprintf(enc, "%s:%s", jira.Username, jira.Token)
	enc.Close()
	return out.String()
}

func (jira *Jira) QueryTickets() ([]SelectOption, error) {
	target := new(url.URL)
	target.Scheme = "https"
	target.Host = "sainsburys-tech.atlassian.net"
	target.Path = "/rest/api/2/search"
	q := target.Query()
	q.Add("jql", "assignee=currentUser() and sprint IN openSprints()")
	target.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", jira.BasicAuth()))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	contents, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	parsedResponse := new(JiraResponse)
	err = json.Unmarshal(contents, &parsedResponse)

	if err != nil {
		return nil, err
	}

	arr := make([]SelectOption, len(parsedResponse.Issues))
	for i, issue := range parsedResponse.Issues {
		arr[i] = SelectOption{fmt.Sprintf("%s: %s", issue.Key, issue.Summary), issue.Key}
	}
	return arr, nil
}
