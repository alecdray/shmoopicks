package musicbrainz

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"shmoopicks/src/internal/core/contextx"
	"strconv"
	"strings"
	"time"
)

const (
	origin = "https://musicbrainz.org"
)

type Include string

var (
	IncludeAliases     Include = "aliases"
	IncludeAnnotation  Include = "annotation"
	IncludeTags        Include = "tags"
	IncludeRatings     Include = "ratings"
	IncludeUserTags    Include = "user-tags"
	IncludeUserRatings Include = "user-ratings"
	IncludeGenres      Include = "genres"
	IncludeUserGenres  Include = "user-genres"
)

func (i Include) String() string {
	return string(i)
}

type Client struct {
	appName      string
	appVersion   string
	contactUrl   string
	contactEmail string
}

type ClientOpt func(*Client) *Client

func WithContactUrl(url string) ClientOpt {
	return func(c *Client) *Client {
		c.contactUrl = url
		return c
	}
}

func WithContactEmail(email string) ClientOpt {
	return func(c *Client) *Client {
		c.contactEmail = email
		return c
	}
}

func NewClient(appName, appVersion string, options ...ClientOpt) (*Client, error) {
	client := &Client{
		appName:    appName,
		appVersion: appVersion,
	}

	for _, option := range options {
		option(client)
	}

	if client.appName == "" {
		return nil, errors.New("appName cannot be empty")
	}
	if client.appVersion == "" {
		return nil, errors.New("appVersion cannot be empty")
	}
	if client.contactUrl == "" && client.contactEmail == "" {
		return nil, errors.New("both contactUrl and contactEmail cannot be empty")
	}

	return client, nil
}

func (client *Client) UserAgent() string {
	contact := client.contactUrl
	if contact == "" {
		contact = client.contactEmail
	}

	return fmt.Sprintf("%s/%s ( %s )", client.appName, client.appVersion, contact)
}

func (client *Client) MakeRequest(ctx contextx.ContextX, method string, path string, query url.Values) (*http.Response, error) {
	reqUrl, err := url.Parse(origin + path)
	if err != nil {
		return nil, err
	}
	reqUrl.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", client.UserAgent())
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{}
	return httpClient.Do(req)
}

type QueryProps struct {
	Query    string
	Limit    int
	Offset   int
	Includes []string
}

type SearchResult struct {
	Created       time.Time      `json:"created"`
	Count         int            `json:"count"`
	Offset        int            `json:"offset"`
	ReleaseGroups []ReleaseGroup `json:"release-groups"`
	Recordings    []Recording    `json:"recordings"`
}

func (client *Client) SearchEntities(ctx contextx.ContextX, entity Entity, props QueryProps) (*SearchResult, error) {
	path := fmt.Sprintf("/ws/2/%s", entity.Slug())

	query := url.Values{}
	if props.Query != "" {
		query.Set("query", props.Query)
	}
	if props.Limit > 0 {
		query.Set("limit", strconv.Itoa(props.Limit))
	}
	if props.Offset > 0 {
		query.Set("offset", strconv.Itoa(props.Offset))
	}
	if len(props.Includes) > 0 {
		query.Set("inc", strings.Join(props.Includes, " "))
	}

	resp, err := client.MakeRequest(ctx, http.MethodGet, path, query)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}
