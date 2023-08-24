package splunkd

import (
	"fmt"
	"net/url"
)

// entry represents one entry returned by a collection after invoking the API
type entry[T any] struct {
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

func (e *entry[T]) setSharing(ss *Client, sharing SplunkSharing) error {
	fullUrl, _ := url.JoinPath(e.Id, "acl")
	e.ACL.Sharing = string(sharing)
	tmp := entry[T]{}
	if err := doSplunkdHttpRequest(ss, "POST", fullUrl, e.ACL.ToURL(), nil, "", &tmp); err != nil {
		return fmt.Errorf("setSharing: cannot share '%s' to '%s'. %w", e.Name, sharing, err)
	}
	e.ACL = tmp.ACL
	return nil
}

func (e *entry[T]) SetSharingGlobal(ss *Client) error {
	if e.ACL.Sharing == string(SplunkSharingGlobal) {
		return nil
	}
	return e.setSharing(ss, SplunkSharingGlobal)
}
