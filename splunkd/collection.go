package splunkd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/prigio/splunk-go-sdk/utils"
)

// collectionEntry represents one entry returned by a collection after invoking the API
type collectionEntry[T any] struct {
	Name   string `json:"name"`
	Id     string `json:"id"`
	Author string `json:"author"`
	Links  struct {
		// more info on the provided links under:
		//    https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTUM/RESTusing#Atom_Feed_response
		List   string `json:"list"`
		Remove string `json:"remove"`
	} `json:"links"`
	ACL     AccessControlList `json:"acl"`
	Content T                 `json:"content"`
}

func (ce *collectionEntry[T]) Delete(ss *Client) error {
	if ce.Links.Remove == "" {
		return fmt.Errorf("%T delete: '%s' cannot be deleted", ce, ce.Name)
	}
	if err := doSplunkdHttpRequest(ss, "DELETE", ce.Links.Remove, nil, nil, "", &discardBody{}); err != nil {
		return fmt.Errorf("%T delete: '%s' cannot be deleted: %w", ce, ce.Name, err)
	}
	return nil
}

func (ce *collectionEntry[T]) setSharing(ss *Client, sharing SplunkSharing) error {
	fullUrl, _ := url.JoinPath(ce.Id, "acl")
	ce.ACL.Sharing = string(sharing)
	tmp := collectionEntry[T]{}
	if err := doSplunkdHttpRequest(ss, "POST", fullUrl, ce.ACL.ToURL(), nil, "", &tmp); err != nil {
		return fmt.Errorf("setSharing: cannot share '%s' to '%s'. %w", ce.Name, sharing, err)
	}
	ce.ACL = tmp.ACL
	return nil
}

func (ce *collectionEntry[T]) SetSharingGlobal(ss *Client) error {
	if ce.ACL.Sharing == string(SplunkSharingGlobal) {
		return nil
	}
	return ce.setSharing(ss, SplunkSharingGlobal)
}

func getUrl(collectionPath, entry string) string {
	var fullUrl string

	if strings.HasPrefix(collectionPath, "/services") {
		fullUrl, _ = url.JoinPath(collectionPath, entry)
	} else {
		fullUrl, _ = url.JoinPath("/services", collectionPath, entry)
	}
	return fullUrl
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
	path string

	Entries []collectionEntry[T] `json:"entry"`

	splunkd *Client
}

func (col *collection[T]) isInitialized() error {
	if col.path == "" || col.name == "" || col.splunkd == nil {
		return fmt.Errorf("uninitialized collection. Use a New... method to properly initialize internal parameters")
	}
	return nil
}

func (col *collection[T]) List() ([]collectionEntry[T], error) {
	if err := col.isInitialized(); err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	fullUrl := getUrl(col.path, "")

	if err := doSplunkdHttpRequest(col.splunkd, "GET", fullUrl, nil, nil, "", &col); err != nil {
		return nil, fmt.Errorf("%s list: %w", col.name, err)
	}
	return col.Entries, nil
}

func (col *collection[T]) Get(entryName string) (*collectionEntry[T], error) {
	if err := col.isInitialized(); err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	fullUrl := getUrl(col.path, entryName)
	tmpCol := collection[T]{}
	if err := doSplunkdHttpRequest(col.splunkd, "GET", fullUrl, nil, nil, "", &tmpCol); err != nil {
		return nil, fmt.Errorf("%s get: %w", col.name, err)
	}
	return &tmpCol.Entries[0], nil
}

func (col *collection[T]) Create(entryName string, params *url.Values) (*collectionEntry[T], error) {
	if err := col.isInitialized(); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	if entryName == "" {
		return nil, utils.NewErrInvalidParam(col.name+" create", nil, "entryName cannot be empty")
	}
	if params == nil || len(*params) == 0 {
		return nil, utils.NewErrInvalidParam(col.name+" create", nil, "params for '%s' cannot be empty", entryName)
	}

	fullUrl := getUrl(col.path, "")
	if !params.Has("name") {
		params.Set("name", entryName)
	}
	tmpCol := collection[T]{}
	if err := doSplunkdHttpRequest(col.splunkd, "POST", fullUrl, nil, []byte(params.Encode()), "", &tmpCol); err != nil {
		return nil, fmt.Errorf("%s create: %w", col.name, err)
	}
	return &tmpCol.Entries[0], nil
}

func (col *collection[T]) CreateNS(ns *Namespace, entryName string, params *url.Values) (*collectionEntry[T], error) {
	if err := col.isInitialized(); err != nil {
		return nil, fmt.Errorf("createNS: %w", err)
	}
	if entryName == "" {
		return nil, utils.NewErrInvalidParam(col.name+" createNS", nil, "entryName cannot be empty")
	}
	if params == nil || len(*params) == 0 {
		return nil, utils.NewErrInvalidParam(col.name+" createNS", nil, "params for '%s' cannot be empty", entryName)
	}
	if ns == nil {
		return nil, utils.NewErrInvalidParam(col.name+" createNS", nil, "namespace for '%s' cannot be nil", entryName)
	}

	var fullUrl string
	if strings.HasPrefix(col.path, "/servicesNS/") {
		//col.path is like  "/servicesNS/user/app/some/other/stuff"
		//i want to have a result like: "" servicesNS, user, app, some/other/stuff
		path := strings.SplitAfterN(col.path, "/", 5)[4]
		fullUrl, _ = url.JoinPath(ns.GetServicesNSUrl(), path)
	} else {
		fullUrl, _ = url.JoinPath(ns.GetServicesNSUrl(), col.path)
	}
	tmpCol := collection[T]{}

	if err := doSplunkdHttpRequest(col.splunkd, "POST", fullUrl, nil, []byte(params.Encode()), "", &tmpCol); err != nil {
		return nil, fmt.Errorf("%s createNS: %w", col.name, err)
	}

	return &tmpCol.Entries[0], nil
}

func (col *collection[T]) Update(entryName string, params *url.Values) error {
	if err := col.isInitialized(); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	if entryName == "" {
		return utils.NewErrInvalidParam(col.name+" update", nil, "entryName cannot be empty")
	}
	if params == nil || len(*params) == 0 {
		return utils.NewErrInvalidParam(col.name+" update", nil, "params for '%s' cannot be empty", entryName)
	}

	fullUrl := getUrl(col.path, entryName)

	if err := doSplunkdHttpRequest(col.splunkd, "POST", fullUrl, nil, []byte(params.Encode()), "", &discardBody{}); err != nil {
		return fmt.Errorf("%s update: %w", col.name, err)
	}
	return nil
}

func (col *collection[T]) Delete(entryName string) error {
	if err := col.isInitialized(); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if entryName == "" {
		return utils.NewErrInvalidParam(col.name+" delete", nil, "entryName cannot be empty")
	}

	fullUrl := getUrl(col.path, entryName)
	if err := doSplunkdHttpRequest(col.splunkd, "DELETE", fullUrl, nil, nil, "", &discardBody{}); err != nil {
		return fmt.Errorf("%s delete: %w", col.name, err)
	}

	return nil
}

// https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTUM/RESTusing#Access_Control_List
func (col *collection[T]) UpdateACL(entryName string, acl AccessControlList) error {
	if err := col.isInitialized(); err != nil {
		return fmt.Errorf("updateACL: %w", err)
	}
	if entryName == "" {
		return utils.NewErrInvalidParam(col.name+" UpdateACL", nil, "entryName cannot be empty")
	}

	fullUrl := getUrl(col.path, entryName) + "/acl"

	aclParams := acl.ToURL()

	currentEntry, err := col.Get(entryName)
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

	if err := doSplunkdHttpRequest(col.splunkd, "POST", fullUrl, nil, []byte(aclParams.Encode()), "", &discardBody{}); err != nil {
		return fmt.Errorf("%s updateACL: %w", col.name, err)
	}
	return nil
}
