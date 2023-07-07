package client

import (
	"fmt"
	"net/url"
)

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

type CredentialResource struct {
	Realm         string `json:"realm"`
	Username      string `json:"username"`
	ClearPassword string `json:"clear_password"`
	EncrPassword  string `json:"encr_password"`
}

type CredentialsCollection struct {
	collection[CredentialResource]
}

func urlEncodeCredential(user, realm string) string {
	if realm == "" {
		return ":" + user + ":"
	} else {
		return realm + ":" + user + ":"
	}
}

func NewCredentialsCollection(ss *SplunkService) *CredentialsCollection {
	var col = &CredentialsCollection{}
	col.name = "credentials"
	col.path = "storage/passwords/"
	col.splunkd = ss
	return col

}

func (col *CredentialsCollection) CreateCred(user, realm, password string) (*collectionEntry[CredentialResource], error) {
	credPostParams := url.Values{}
	credPostParams.Set("name", user)
	credPostParams.Set("password", password)
	if realm != "" {
		credPostParams.Set("realm", realm)
	}
	entryId := urlEncodeCredential(user, realm)
	return col.Create(entryId, &credPostParams)
}

func (col *CredentialsCollection) GetCred(user, realm string) (*collectionEntry[CredentialResource], error) {
	entryId := urlEncodeCredential(user, realm)
	return col.Get(entryId)
}

func (col *CredentialsCollection) UpdateCred(user, realm, newPassword string) error {

	entryId := urlEncodeCredential(user, realm)

	credPostParams := url.Values{}
	credPostParams.Set("password", newPassword)

	if err := col.Update(entryId, &credPostParams); err != nil {
		return fmt.Errorf("%s UpdateCred: %w", col.name, err)
	}

	return nil
}

// https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTUM/RESTusing#Access_Control_List
func (col *CredentialsCollection) UpdateCredACL(user, realm string, aclParams *url.Values) error {
	entryId := urlEncodeCredential(user, realm)
	return col.UpdateACL(entryId, aclParams)
}