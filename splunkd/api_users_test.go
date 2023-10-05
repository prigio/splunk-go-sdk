package splunkd

import (
	"testing"

	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	"github.com/prigio/splunk-go-sdk/v2/utils"
)

func TestUsersList(t *testing.T) {
	t.Log("INFO Connecting to Splunk")

	ss := mustLoginToSplunk(t)

	users := ss.GetUsers()

	uList, err := users.List()

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(uList) == 0 {
		t.Errorf("No users returned")
	}

	names := utils.ListOfVals(uList, func(e *entry[UserResource]) string { return e.Name })
	t.Logf("API returned %d users: %+v", len(uList), names)

}

func TestUsersCreate(t *testing.T) {
	ss := mustLoginToSplunk(t)

	users := ss.GetUsers()
	newUser := uuid.New().String()[0:8]
	newUserPass := uuid.New().String()[0:16]
	newEmail := newUser + "@test.com"
	roles := []string{"user", "power"}

	t.Logf("INFO Creating user='%s' roles='%v'", newUser, roles)

	ur := UserResource{Password: newUserPass, Roles: roles, Email: newEmail}

	//urlValues, err := query.Values(ur)
	//t.Logf("urlValues: %+v", urlValues)

	u, err := users.CreateUser(newUser, ur)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !utils.IsSubset(roles, u.Content.Roles) {
		t.Errorf("roles of created user do not match. Found: '%v', expected: '%v", u.Content.Roles, roles)
	}
	if newEmail != u.Content.Email {
		t.Errorf("email of created user does not match. Found: '%s', expected: '%s'", u.Content.Email, newEmail)
	}
}

func TestUsersUpdate(t *testing.T) {
	ss := mustLoginToSplunk(t)

	users := ss.GetUsers()
	newUser := uuid.New().String()[0:8]
	newRealname := newUser + " Lastname"
	newUserPass := uuid.New().String()[0:16]
	newEmail := newUser + "@test.com"
	roles := []string{"user", "power"}
	changedRealname := newUser + " Changed Lastname"
	changedEmail := newUser + "@newEmail.com"
	changedRoles := []string{"admin", "user", "power"}
	changedPassword := uuid.New().String()[0:16]

	t.Logf("INFO Creating user='%s' roles='%v'", newUser, roles)

	_, err := users.CreateUser(newUser, UserResource{Password: newUserPass, Realname: newRealname, Roles: roles, Email: newEmail})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	updatedUser := UserResource{Realname: changedRealname, Roles: changedRoles, Email: changedEmail}
	urlValues, _ := query.Values(updatedUser)
	if err := users.Update(newUser, &urlValues); err != nil {
		t.Error(err)
		t.FailNow()
	}

	u, _ := users.Get(newUser)

	if changedEmail != u.Content.Email {
		t.Errorf("email of updated user does not match. Found: '%s', expected: '%s'", u.Content.Email, changedEmail)
	}
	if changedEmail != u.Content.Email {
		t.Errorf("realname of updated user does not match. Found: '%s', expected: '%s'", u.Content.Realname, changedRealname)
	}

	t.Logf("INFO Updating password for user='%s' to password='%s'", newUser, changedPassword)
	updatedUser = UserResource{Password: changedPassword}
	urlValues, _ = query.Values(updatedUser)
	if err := users.Update(newUser, &urlValues); err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log("INFO Testing API login with the new user password")
	if err := ss.Login(newUser, changedPassword, ""); err != nil {
		t.Errorf("New password for user='%s' does not work. password='%s'", newUser, changedPassword)
	}
}

func TestUsersDelete(t *testing.T) {
	ss := mustLoginToSplunk(t)

	users := ss.GetUsers()
	newUser := uuid.New().String()[0:8]
	newUserPass := uuid.New().String()[0:16]
	newEmail := newUser + "@test.com"
	roles := []string{"user", "power"}

	t.Logf("INFO Creating user='%s' roles='%v'", newUser, roles)
	ur := UserResource{Password: newUserPass, Roles: roles, Email: newEmail}
	_, err := users.CreateUser(newUser, ur)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("INFO Deleting user='%s'", newUser)
	err = users.Delete(newUser)
	if err != nil {
		t.Error(err)
	}
	t.Logf("INFO Checking wheter user='%s' still exists", newUser)
	uList, err := users.List()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	uNames := utils.ListOfVals(uList, func(e *entry[UserResource]) string { return e.Name })
	if utils.In(newUser, uNames) {
		t.Errorf("user '%v' was not deleted. It is within the list of existing users '%v'", newUser, uNames)
	}
}
