package client

import (
	"fmt"
)

type SplunkSharing string

const (
	SplunkSharingUser   SplunkSharing = "user"
	SplunkSharingApp    SplunkSharing = "app"
	SplunkSharingSystem SplunkSharing = "system"
	SplunkSharingGlobal SplunkSharing = "global"
)

type NameSpace struct {
	owner   string
	app     string
	sharing SplunkSharing
}

// GetNamespace instantiates a new Splunk namespace
func NewNamespace(owner, app string, sharing SplunkSharing) (*NameSpace, error) {
	if sharing != "" && sharing != SplunkSharingUser && sharing != SplunkSharingApp && sharing != SplunkSharingSystem && sharing != SplunkSharingGlobal {
		return nil, fmt.Errorf("invalid 'sharing', must be one of: %s, %s, %s, %s. provided: \"%s\"", SplunkSharingUser, SplunkSharingApp, SplunkSharingSystem, SplunkSharingGlobal, sharing)
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

	ns := &NameSpace{
		owner:   owner,
		app:     app,
		sharing: sharing,
	}

	return ns, nil
}

func (ns *NameSpace) GetServicesNSUrl() string {
	return "/servicesNS/" + ns.owner + "/" + ns.app + "/"
}
