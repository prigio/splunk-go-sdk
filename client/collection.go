package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// collectionEntry represents one entry returned by a collection after invoking the API
type collectionEntry[T any] struct {
	Name    string            `json:"name"`
	Id      string            `json:"id"`
	Author  string            `json:"author"`
	ACL     AccessControlList `json:"acl"`
	Content T                 `json:"content"`
}

// collection represents a collection of entries regarding an API endpoint
type collection[T any] struct {
	Origin  string `json:"origin"`
	Link    string `json:"link"`
	Updated string `json:"updated"`
	Paging  struct {
		Total   int `json:"total"`
		PerPage int `json:"perPage"`
		Offset  int `json:"offset"`
	}
	// name is the internal name used to refer to this collection[T]. Mostly used for error management purposes
	name string
	// path represents the part of URL after services/ or servicesNS/user/app/ to access the resources of the collection
	path    string
	Entries []collectionEntry[T] `json:"entry"`

	splunkd *SplunkService
}

func (col *collection[T]) List() ([]collectionEntry[T], error) {
	if col.path == "" || col.name == "" || col.splunkd == nil {
		return nil, fmt.Errorf("list: uninitialized collection. Use a New... method to properly initialize the collection")
	}

	httpCode, respBody, err := col.splunkd.doHttpRequest("GET", "services/"+col.path, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%s list: %w", col.name, err)
	}
	err = col.parseResponse(httpCode, respBody)
	if err != nil {
		return nil, fmt.Errorf("%s list: %w", col.name, err)
	}
	return col.Entries, nil
}

func (col *collection[T]) Get(entryId string) (collectionEntry[T], error) {
	if col.path == "" || col.name == "" || col.splunkd == nil {
		return collectionEntry[T]{}, fmt.Errorf("get: uninitialized collection. Use a New... method to properly initialize the collection")
	}
	var fullUrl string

	if strings.HasPrefix(entryId, "http") {
		// entryId is an absolute URL, including protocol, server
		fullUrl = entryId
	} else {
		fullUrl, _ = url.JoinPath("services", col.path, entryId)
	}
	httpCode, respBody, err := col.splunkd.doHttpRequest("GET", fullUrl, nil, nil)

	if err != nil {
		return collectionEntry[T]{}, fmt.Errorf("%s get: %w", col.name, err)
	}

	tmpCol := collection[T]{}
	err = tmpCol.parseResponse(httpCode, respBody)
	if err != nil {
		return collectionEntry[T]{}, fmt.Errorf("%s get: %w", col.name, err)
	}

	return tmpCol.Entries[0], nil
}

func (col *collection[T]) Create(entry string, params *url.Values) (*collectionEntry[T], error) {
	if col.path == "" || col.name == "" || col.splunkd == nil {
		return nil, fmt.Errorf("create: uninitialized collection. Use a New... method to properly initialize the collection")
	}
	if params == nil || len(*params) == 0 {
		return nil, fmt.Errorf("%s create: cannot create entry without any properties. entry='%s'", col.name, entry)
	}

	fullUrl, _ := url.JoinPath("services", col.path)

	httpCode, respBody, err := col.splunkd.doHttpRequest("POST", fullUrl, nil, strings.NewReader(params.Encode()))

	if err != nil {
		return nil, fmt.Errorf("%s create: %w", col.name, err)
	}
	tmpCol := collection[T]{}
	err = tmpCol.parseResponse(httpCode, respBody)
	if err != nil {
		return nil, fmt.Errorf("%s create: %w", col.name, err)
	}

	return &tmpCol.Entries[0], nil
}

func (col *collection[T]) Update(entryId string, params *url.Values) error {
	if col.path == "" || col.name == "" || col.splunkd == nil {
		return fmt.Errorf("update: uninitialized collection. Use a New... method to properly initialize the collection")
	}

	var fullUrl string

	if strings.HasPrefix(entryId, "http") {
		// entryId is an absolute URL, including protocol, server
		fullUrl = entryId
	} else {
		fullUrl, _ = url.JoinPath("services", col.path, entryId)
	}

	httpCode, respBody, err := col.splunkd.doHttpRequest("POST", fullUrl, nil, strings.NewReader(params.Encode()))

	if err != nil {
		return fmt.Errorf("%s update: %w", col.name, err)
	}
	if httpCode >= 400 {
		return fmt.Errorf("%s update: HTTP %v - %s", col.name, httpCode, string(respBody))
	}
	return nil
}

func (col *collection[T]) Delete(entryId string) error {
	if col.path == "" || col.name == "" || col.splunkd == nil {
		return fmt.Errorf("delete: uninitialized collection. Use a New... method to properly initialize the collection")
	}

	var fullUrl string

	if strings.HasPrefix(entryId, "http") {
		// entryId is an absolute URL, including protocol, server
		fullUrl = entryId
	} else {
		fullUrl, _ = url.JoinPath("services", col.path, entryId)
	}

	httpCode, respBody, err := col.splunkd.doHttpRequest("DELETE", fullUrl, nil, nil)
	if err != nil {
		return fmt.Errorf("%s delete: %w", col.name, err)
	}
	if httpCode >= 400 {
		return fmt.Errorf("%s delete: HTTP %v - %s", col.name, httpCode, string(respBody))
	}
	return nil
}

// https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTUM/RESTusing#Access_Control_List
func (col *collection[T]) UpdateACL(entryId string, aclParams *url.Values) error {
	if col.path == "" || col.name == "" || col.splunkd == nil {
		return fmt.Errorf("updateACL: uninitialized collection. Use a New... method to properly initialize the collection")
	}

	var fullUrl string

	if strings.HasPrefix(entryId, "http") {
		// entryId is an absolute URL, including protocol, server
		fullUrl, _ = url.JoinPath(entryId, "acl")
	} else {
		fullUrl, _ = url.JoinPath("services", col.path, entryId, "acl")
	}

	currentEntry, err := col.Get(entryId)
	if err != nil {
		return fmt.Errorf("%s updateACL: %w", col.name, err)
	}

	// sharing and owner are required by the API, and cannot be empty
	if !aclParams.Has("sharing") || aclParams.Get("sharing") == "" {
		aclParams.Set("sharing", currentEntry.ACL.Sharing)
	}

	if !aclParams.Has("owner") || aclParams.Get("owner") == "" {
		aclParams.Set("owner", currentEntry.ACL.Owner)
	}

	// this is actually only required for savedsearches
	if strings.Contains(fullUrl, "/saved/searches/") && (!aclParams.Has("app") || aclParams.Get("app") == "") {
		aclParams.Set("app", currentEntry.ACL.App)
	}
	// the "app" setting is only used by savedsearches. It is not accepted by the splunkd api.
	// to change the "app", the ".../move" endpoint must be used
	if !strings.Contains(fullUrl, "/saved/searches/") && aclParams.Has("app") {
		aclParams.Del("app")
	}

	// permissions need be updated for both read and write ´, otherwise the other is set to empty by splunkd
	// use the current ones if one of them is set but the other is not.
	if aclParams.Has("perms.read") && !aclParams.Has("perms.write") {
		aclParams.Set("perms.write", strings.Join(currentEntry.ACL.Perms.Write, ", "))
	}

	if aclParams.Has("perms.write") && !aclParams.Has("perms.read") {
		aclParams.Set("perms.read", strings.Join(currentEntry.ACL.Perms.Read, ", "))
	}

	// splunkd api does not support multiple permission parameters, so they get joined into a single string
	if perms, ok := (*aclParams)["perms.read"]; ok && len(perms) > 0 {
		aclParams.Set("perms.read", strings.Join((*aclParams)["perms.read"], ", "))
	}

	if perms, ok := (*aclParams)["perms.write"]; ok && len(perms) > 0 {
		aclParams.Set("perms.write", strings.Join((*aclParams)["perms.write"], ", "))
	}

	httpCode, respBody, err := col.splunkd.doHttpRequest("POST", fullUrl, nil, strings.NewReader(aclParams.Encode()))

	if err != nil {
		return fmt.Errorf("%s updateACL: %w", col.name, err)
	}
	if httpCode >= 400 {
		return fmt.Errorf("%s updateACL: HTTP %v - %s", col.name, httpCode, string(respBody))
	}
	return nil
}

func (col *collection[T]) parseResponse(httpCode int, httpRespBody []byte) error {
	if httpCode >= 400 {
		// HTTP 401
		// {"messages":[{"type":"WARN","text":"call not properly authenticated"}]}%
		return fmt.Errorf("%s: HTTP %v - %s", col.name, httpCode, string(httpRespBody))
	}

	if err := json.Unmarshal(httpRespBody, col); err != nil {
		return fmt.Errorf("%s: %w", col.name, err)
	}
	return nil
}
