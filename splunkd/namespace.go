package splunkd

import "github.com/prigio/splunk-go-sdk/v2/errors"

type SplunkSharing string

const (
	SplunkSharingUser   SplunkSharing = "user"
	SplunkSharingApp    SplunkSharing = "app"
	SplunkSharingSystem SplunkSharing = "system"
	SplunkSharingGlobal SplunkSharing = "global"
)

type Namespace struct {
	owner   string
	app     string
	sharing SplunkSharing
}

// GetNamespace instantiates a new Splunk namespace
func NewNamespace(owner, app string, sharing SplunkSharing) (*Namespace, error) {
	if sharing != "" && sharing != SplunkSharingUser && sharing != SplunkSharingApp && sharing != SplunkSharingSystem && sharing != SplunkSharingGlobal {
		return nil, errors.NewErrInvalidParam("newNamespace", nil, "'sharing', must be one of: %s, %s, %s, %s. provided: \"%s\"", SplunkSharingUser, SplunkSharingApp, SplunkSharingSystem, SplunkSharingGlobal, sharing)
	}

	if owner == "" {
		owner = "-"
	}
	if app == "" {
		app = "-"
	}
	if sharing == "" {
		sharing = SplunkSharingUser
	}

	ns := &Namespace{
		owner:   owner,
		app:     app,
		sharing: sharing,
	}

	return ns, nil
}

func (ns *Namespace) GetServicesNSUrl() string {
	o := ns.owner
	a := ns.app
	if o == "" {
		o = "-"
	}
	if a == "" {
		a = "-"
	}
	return "/servicesNS/" + o + "/" + a + "/"
}
