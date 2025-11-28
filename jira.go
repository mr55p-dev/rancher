package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/charmbracelet/huh"
)

type JiraResponse struct {
	Issues []JiraIssue `json:"issues"`
}

type JiraIssue struct {
	Id     string     `json:"id"`
	Key    string     `json:"key"`
	Fields JiraFields `json:"fields"`
}

type JiraFields struct {
	Summary string `json:"summary"`
}

type Jira struct {
	Username       string `config:"username,optional"`
	Token          string `config:"token,optional"`
	Query          string `config:"query,optional"`
	Host           string `config:"host,optional"`
	SuffixTicketID bool   `config:"suffix-ticket-id,optional"`
}

func (jira *Jira) BasicAuth() string {
	out := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, out)
	fmt.Fprintf(enc, "%s:%s", jira.Username, jira.Token)
	enc.Close()
	return out.String()
}

func (jira *Jira) QueryTickets(debug bool) ([]huh.Option[string], error) {
	target := new(url.URL)
	target.Scheme = "https"
	target.Host = jira.Host
	target.Path = "/rest/api/3/search/jql"
	q := target.Query()
	q.Add("jql", jira.Query)
	q.Add("fields", "key,summary")
	target.RawQuery = q.Encode()

	if debug {
		fmt.Printf("Jira Query: %s\n", jira.Query)
		fmt.Printf("Jira URL: %s\n", target.String())
	}

	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	if debug {
		fmt.Printf("Username: %s\n", jira.Username)
		fmt.Printf("Token: %s\n", jira.Token)
		fmt.Printf("Authorization token: %s\n", jira.BasicAuth())
	}
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", jira.BasicAuth()))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira API returned status %d: %s", res.StatusCode, res.Status)
	}

	contents, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if debug {
		fmt.Printf("Jira API Response:\n%s\n", string(contents))
	}
	parsedResponse := new(JiraResponse)
	err = json.Unmarshal(contents, &parsedResponse)

	if err != nil {
		return nil, err
	}

	if len(parsedResponse.Issues) == 0 {
		return nil, fmt.Errorf("no tickets returned from Jira (query may be too restrictive or no tickets match)")
	}

	arr := make([]huh.Option[string], len(parsedResponse.Issues))
	for i, issue := range parsedResponse.Issues {
		arr[i] = huh.Option[string]{
			Key:   fmt.Sprintf("%s: %s", issue.Key, issue.Fields.Summary),
			Value: issue.Key,
		}
	}
	return arr, nil
}
