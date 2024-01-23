package auth

import (
	"crypto/tls"
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
	// DisplayName is a human-readable string to identify the user in a GUI
	DisplayName string `ldap:"displayName"`
	// GUID is a persistent, unique user ID to reliably identify the user by the system
	GUID []byte `ldap:"objectGUID"`
}

// authorizationResult represents all data returned by the user authorization.
//
// If Predicate is anything but GRANTED, the other attributes will be omitted.
type authorizationResult struct {
	Predicate   authorizationPredicate
	Permissions *permissions
	UserEntry   *userEntry
}

// authorizeUser connects to the LDAP server and checks the given users
// credentials.
//
// Environment variables are used to identify the LDAP server and attributes.
//
// We return a value indicating whether the provided credentials are valid, and
// if so, what level of access should be grated to the user.
func authorizeUser(username string, password string) (authorizationResult, error) {
	l, err := ldap.DialURL(os.Getenv("AD_URL"))
	if err != nil {
		return authorizationResult{}, err
	}
	defer l.Close()

	// Reconnect with TLS
	if err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		return authorizationResult{}, err
	}

	// First bind with a read only user
	if err = l.Bind(os.Getenv("AD_USER"), os.Getenv("AD_PASS")); err != nil {
		return authorizationResult{}, err
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("AD_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(sAMAccountName=%s))", ldap.EscapeFilter(username)),
		[]string{"dn", "displayName", "objectGUID"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return authorizationResult{}, err
	}
	if len(sr.Entries) != 1 {
		return authorizationResult{Predicate: INVALID}, nil
	}

	userDn := sr.Entries[0].DN

	// Bind as the user to verify their password
	err = l.Bind(userDn, password)
	if err != nil {
		return authorizationResult{Predicate: INVALID}, nil
	}

	// Rebind as the read only user for any further queries
	if err = l.Bind(os.Getenv("AD_USER"), os.Getenv("AD_PASS")); err != nil {
		return authorizationResult{}, err
	}

	// Check basic access rights
	if hasAccess, err := isGroupMember(l, userDn, os.Getenv("AD_ACCESS_GROUP")); err != nil {
		return authorizationResult{}, err
	} else if !hasAccess {
		return authorizationResult{Predicate: DENIED}, nil
	}
	// At this point, the user has proven basic access authorization (i.e.: Predicate: GRANTED)

	user := userEntry{}
	if err := sr.Entries[0].Unmarshal(&user); err != nil {
		return authorizationResult{}, err
	}

	// Check further permissions
	permissions := permissions{}
	if isAdmin, err := isGroupMember(l, userDn, os.Getenv("AD_ADMIN_GROUP")); err != nil {
		return authorizationResult{}, err
	} else if isAdmin {
		permissions.Admin = true
	}

	return authorizationResult{
		Predicate:   GRANTED,
		UserEntry:   &user,
		Permissions: &permissions,
	}, nil
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
