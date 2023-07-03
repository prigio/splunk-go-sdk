package client

import (
	"fmt"
	"net/url"
	"strings"
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

func (col *CredentialsCollection) CreateCred(user, realm, password string) (CredentialResource, AccessControlList, error) {

	fullUrl, _ := url.JoinPath("services", col.path)

	credPostParams := url.Values{}
	credPostParams.Set("name", user)
	credPostParams.Set("password", password)
	if realm != "" {
		credPostParams.Set("realm", realm)
	}

	httpCode, respBody, err := col.splunkd.doHttpRequest("POST", fullUrl, nil, strings.NewReader(credPostParams.Encode()))

	if err != nil {
		return CredentialResource{}, AccessControlList{}, fmt.Errorf("%s create: %w", col.name, err)
	}
	cred := NewCredentialsCollection(col.splunkd)
	err = cred.parseResponse(httpCode, respBody)
	if err != nil {
		return CredentialResource{}, AccessControlList{}, fmt.Errorf("%s create: %w", col.name, err)
	}

	return cred.Entries[0].Content, cred.Entries[0].ACL, nil
}

func (col *CredentialsCollection) GetCred(user, realm string) (CredentialResource, AccessControlList, error) {

	entryId := urlEncodeCredential(user, realm)

	cred, err := col.Get(entryId)

	if err != nil {
		return CredentialResource{}, AccessControlList{}, fmt.Errorf("%s GetCredential: %w", col.name, err)
	}
	return cred.Content, cred.ACL, nil
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

/*
func (ss *SplunkService) setCredential(user, realm, password string) (Credential, error) {
	var fullUrl string
	var resp *http.Response
	//var httpCode int
	//var respBody []byte
	var err error
	credParams := url.Values{}

	_, err = ss.getCredential(user, realm)
	if err != nil {
		// no credential present, let's create one
		// Submit credentials form
		credParams.Set("name", user)
		credParams.Set("password", password)
		if realm != "" {
			credParams.Set("realm", realm)
		}
		if resp, err = ss.httpClient.PostForm(ss.buildUrl(pathStoragePasswords), credParams); err != nil {
			return Credential{}, fmt.Errorf("splunk service setCredential: %s", err.Error())
		} else {
			fmt.Println(resp.StatusCode)
		}
	} else {
		// credential IS present, need to update it
		if realm != "" {
			fullUrl = fmt.Sprintf("%s/%s:%s:", pathStoragePasswords, realm, user)
		} else {
			fullUrl = fmt.Sprintf("%s/%s", pathStoragePasswords, user)
		}

		credParams.Set("password", password)

		if _, err = ss.httpClient.PostForm(ss.buildUrl(fullUrl), credParams); err != nil {
			return Credential{}, fmt.Errorf("splunk service setCredential: %s", err.Error())
		}
	}
	return Credential{}, nil
}
*/
