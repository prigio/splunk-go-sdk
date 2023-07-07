package client

import (
	"net/url"
	"strings"
)

// AccessControlList represents the access rights of a splunk resource
// See: https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTUM/RESTusing#Access_Control_List

type AccessControlList struct {
	App     string `json:"app"`
	Owner   string `json:"owner"`
	Sharing string `json:"sharing"`
	Perms   struct {
		Read  []string `json:"read"`
		Write []string `json:"write"`
	}
}

// ServicesNSPath returns the user/app path which needs be added after ../servcesNS/ in an API call to splunkd
// Returns a url-fragment in the form '<user>/app/'. The fragment DOES NOT start with a slash and ENDS with a slash.
// See: https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTUM/RESTusing#Namespace
func (acl *AccessControlList) ServicesNSPath() string {
	var appSect, ownerSect string
	if acl.App == "" || acl.App == "*" || acl.App == "-" {
		appSect = "-"
	} else {
		appSect = acl.App
	}
	if acl.Owner == "" || acl.Owner == "*" || acl.Owner == "-" {
		ownerSect = "-"
	} else {
		ownerSect = acl.Owner
	}
	return ownerSect + "/" + appSect + "/"
}

// ToURL encodes the ACL information into URL values which can be used to set properties on the splunkd API
func (acl *AccessControlList) ToURL() *url.Values {
	v := url.Values{}
	v.Set("app", acl.App)
	v.Set("owner", acl.Owner)
	v.Set("sharing", acl.Sharing)
	if len(acl.Perms.Read) > 0 {
		v.Set("perms.read", strings.Join(acl.Perms.Read, ", "))
	}
	if len(acl.Perms.Write) > 0 {
		v.Set("perms.write", strings.Join(acl.Perms.Write, ", "))
	}
	return &v
}