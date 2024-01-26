package auth

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"github.com/go-ldap/ldap/v3"
)

// authorizationPredicate represents the outcome of a user authorization
type authorizationPredicate int

const (
	// INVALID indicates an invalid username / password combination
	INVALID authorizationPredicate = iota // 0
	// DENIED denies any user access based on missing LDAP group membership
	DENIED // 1
	// GRANTED grants basic access rights
	GRANTED // 2
)

// permissions grants additional permissions based an LDAP group memberships
type permissions struct {
	Admin bool `json:"admin,omitempty"`
}

// userEntry represents an LDAP user
type userEntry struct {
	// ID is a persistent, unique user ID to reliably identify the user by the system
	ID []byte `ldap:"objectGUID" json:"id"`
	// DisplayName is a human-readable string to identify the user in a GUI
	DisplayName string `ldap:"displayName" json:"displayName"`
	// Permissions are the users extended permissions within x-man
	Permissions *permissions `json:"permissions"`
}

// authorizationResult represents all data returned by the user authorization.
//
// If Predicate is anything but GRANTED, the other attributes will be omitted.
type authorizationResult struct {
	Predicate authorizationPredicate
	UserEntry *userEntry
}

// authorizeUser connects to the LDAP server and checks the given users
// credentials.
//
// Environment variables are used to identify the LDAP server and attributes.
//
// We return a value indicating whether the provided credentials are valid, and
// if so, what level of access should be grated to the user.
func authorizeUser(username string, password string) (authorizationResult, error) {
	l, err := connectReadonly()
	if err != nil {
		return authorizationResult{}, err
	}
	defer l.Close()

	// Search for the given username
	user, err := getLdapUserEntry(l, "sAMAccountName", username)
	if err != nil {
		return authorizationResult{}, err
	}
	if user == nil {
		return authorizationResult{Predicate: INVALID}, nil
	}

	// Bind as the user to verify their password
	err = l.Bind(user.DN, password)
	if err != nil {
		return authorizationResult{Predicate: INVALID}, nil
	}

	// Rebind as the read only user for any further queries
	if err = l.Bind(os.Getenv("AD_USER"), os.Getenv("AD_PASS")); err != nil {
		return authorizationResult{}, err
	}

	// Check basic access rights
	if hasAccess, err := isGroupMember(l, user.DN, os.Getenv("AD_ACCESS_GROUP")); err != nil {
		return authorizationResult{}, err
	} else if !hasAccess {
		return authorizationResult{Predicate: DENIED}, nil
	}
	// At this point, the user has proven basic access authorization (i.e.: Predicate: GRANTED)

	userEntry := userEntry{}
	if err := user.Unmarshal(&userEntry); err != nil {
		return authorizationResult{}, err
	}

	// Check further permissions
	permissions, err := getUserPermissions(l, user.DN)
	if err != nil {
		return authorizationResult{}, err
	}
	userEntry.Permissions = &permissions

	return authorizationResult{
		Predicate: GRANTED,
		UserEntry: &userEntry,
	}, nil
}

func listUsers() ([]userEntry, error) {
	l, err := connectReadonly()
	if err != nil {
		return []userEntry{}, err
	}
	defer l.Close()

	// Get all members of AD_ACCESS_GROUP
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("AD_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(cn=%s)(objectClass=group))", ldap.EscapeFilter(os.Getenv("AD_ACCESS_GROUP"))),
		[]string{"member"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return []userEntry{}, err
	}
	if len(sr.Entries) != 1 {
		return []userEntry{}, errors.New("ldap group not found: " + os.Getenv("AD_ACCESS_GROUP"))
	}
	members := sr.Entries[0].GetAttributeValues("member")

	userEntries := make([]userEntry, 0)
	for _, userDn := range members {
		user, err := getLdapUserEntry(l, "distinguishedName", userDn)
		if err != nil {
			return userEntries, err
		} else if user == nil {
			continue
		}
		userEntry := userEntry{}
		if err := user.Unmarshal(&userEntry); err != nil {
			return userEntries, err
		}
		permissions, err := getUserPermissions(l, user.DN)
		if err != nil {
			return userEntries, err
		}
		userEntry.Permissions = &permissions
		userEntries = append(userEntries, userEntry)
	}
	return userEntries, nil
}

// connectReadonly connects to the LDAP server and binds with readonly
// credentials.
//
// Users are responsible to close the connection with `l.Close()` afterwards.
func connectReadonly() (*ldap.Conn, error) {
	l, err := ldap.DialURL(os.Getenv("AD_URL"))
	if err != nil {
		return nil, err
	}

	// Reconnect with TLS
	if err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		l.Close()
		return nil, err
	}

	// First bind with a read only user
	if err := l.Bind(os.Getenv("AD_USER"), os.Getenv("AD_PASS")); err != nil {
		l.Close()
		return nil, err
	}
	return l, nil
}

// isGroupMember returns `true` if the given user is member of the given group.
//
// `l` should be an open LDAP connection with readonly access.
func isGroupMember(l *ldap.Conn, userDn string, groupCn string) (bool, error) {
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("AD_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(member=%s)(objectClass=group)(cn=%s))", ldap.EscapeFilter(userDn), ldap.EscapeFilter(groupCn)),
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return false, err
	}
	return len(sr.Entries) == 1, nil
}

// getLdapUserEntry searches for a user with the given LDAP key and value.
//
// `l` should be an open LDAP connection with readonly access.
//
// Returns nil when the user could not be found.
func getLdapUserEntry(l *ldap.Conn, key string, value string) (*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("AD_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(%s=%s))", key, ldap.EscapeFilter(value)),
		[]string{"dn", "displayName", "objectGUID"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) != 1 {
		return nil, nil
	}
	return sr.Entries[0], nil
}

func getUserPermissions(l *ldap.Conn, userDn string) (permissions, error) {
	permissions := permissions{}
	if isAdmin, err := isGroupMember(l, userDn, os.Getenv("AD_ADMIN_GROUP")); err != nil {
		return permissions, err
	} else if isAdmin {
		permissions.Admin = true
	}
	return permissions, nil
}
