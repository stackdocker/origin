package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// Auth system gets identity name and provider
// POST to UserIdentityMapping, get back error or a filled out UserIdentityMapping object

type User struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	FullName string `json:"fullname,omitempty"`

	Identities []string `json:"identities,omitempty"`

	Groups []string `json:"groups,omitempty"`
}

type UserList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`
	Items                []User `json:"items,omitempty"`
}

type Identity struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	// ProviderName is the source of identity information
	ProviderName string `json:"providerName,omitempty"`

	// ProviderUserName uniquely represents this identity in the scope of the provider
	ProviderUserName string `json:"providerUserName,omitempty"`

	// User is a reference to the user this identity is associated with
	// Both Name and UID must be set
	User kapi.ObjectReference `json:"user,omitempty"`

	Extra map[string]string `json:"extra,omitempty"`
}

type IdentityList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`
	Items                []Identity `json:"items,omitempty"`
}

type UserIdentityMapping struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	Identity kapi.ObjectReference `json:"identity,omitempty"`
	User     kapi.ObjectReference `json:"user,omitempty"`
}

// Group represents a referenceable set of Users
type Group struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	Users []string `json:"users,omitempty"`
}

type GroupList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`
	Items                []Group `json:"items,omitempty"`
}
