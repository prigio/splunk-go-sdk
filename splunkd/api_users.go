package splunkd

import (
	"github.com/google/go-querystring/query"
	"github.com/prigio/splunk-go-sdk/utils"
)

// https://docs.splunk.com/Documentation/Splunk/9.1.0/RESTREF/RESTaccess#authentication.2Fusers.2F.7Bname.7D

type UserResource struct {
	Realname   string `json:"realname" url:"realname,omitempty"`
	Email      string `json:"email" url:"email,omitempty"`
	DefaultApp string `json:"defaultApp" url:"defaultApp,omitempty"`
	// Password is only needed when creating a new user, or updating its password
	Password string `url:"password,omitempty"`
	// OldPassword is only needed when changing password of an existing user
	OldPassword string   `url:"oldpassword,omitempty"`
	Roles       []string `json:"roles" url:"roles,omitempty"`
	Lang        string   `json:"lang" url:"lang,omitempty"`
	Tz          string   `json:"tz" url:"tz,omitempty"`
	// Indicates whether to force user password change.
	ForceChangePass bool `url:"force-change-pass,omitempty"`
	//DefaultAppIsUserOverride bool     `json:"defaultAppIsUserOverride",url:""`
	//DefaultAppSourceRole     string   `json:"defaultAppSourceRole",url:""`
	LastSuccessfulLogin int64    `json:"last_successful_login" url:"-"`
	LockedOut           bool     `json:"locked-out" url:"-"`
	Capabilities        []string `json:"capabilities" url:"-"`
	UserType            string   `json:"type" url:"-"`
}

type UsersCollection struct {
	collection[UserResource]
}

func NewUsersCollection(ss *Client) *UsersCollection {
	var col = &UsersCollection{}
	col.name = "users"
	col.path = "/authentication/users"
	col.splunkd = ss
	return col
}

func (col *UsersCollection) CreateUser(name string, details UserResource) (*entry[UserResource], error) {
	if name == "" {
		return nil, utils.NewErrInvalidParam(col.name+" createUser", nil, "entryName cannot be empty")
	}
	urlValues, err := query.Values(details)
	if err != nil {
		return nil, utils.NewErrInvalidParam(col.name+" createUser", err, "details cannot be converted to url.Values")
	}
	return col.Create(name, &urlValues)
}
