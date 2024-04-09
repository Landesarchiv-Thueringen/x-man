package auth

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/go-ldap/ldap/v3"
)

type ldapConfiguration struct {
	IDAttribute          string
	IDAttributeIsBinary  bool
	UsernameAttribute    string
	EmailAttribute       string
	DisplayNameAttribute string
}

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
	ID string `json:"id"`
	// DisplayName is a human-readable string to identify the user in a GUI
	DisplayName string `json:"displayName"`
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

func GetDisplayName(userID string) string {
	l := connectReadonly()
	defer l.Close()
	user := getLdapUserEntry(l, getIDFilter(userID))
	return getDisplayName(user)
}

func GetMailAddress(userID string) string {
	l := connectReadonly()
	defer l.Close()
	user := getLdapUserEntryWithAttributes(l, getIDFilter(userID), []string{"dn", "mail"})
	if user == nil {
		panic("could not find user with ID " + string(userID))
	}
	mail := user.GetAttributeValue("mail")
	if mail == "" {
		dn := user.GetAttributeValue("DN")
		panic("user " + dn + " has no mail address configured")
	}
	return mail
}

// authorizeUser connects to the LDAP server and checks the given users
// credentials.
//
// Environment variables are used to identify the LDAP server and attributes.
//
// We return a value indicating whether the provided credentials are valid, and
// if so, what level of access should be grated to the user.
func authorizeUser(username string, password string) authorizationResult {
	l := connectReadonly()
	defer l.Close()

	// Search for the given username
	filter := getUserNameFilter(username)
	user := getLdapUserEntry(l, filter)
	if user == nil {
		return authorizationResult{Predicate: INVALID}
	}

	// Bind as the user to verify their password
	err := l.Bind(user.DN, password)
	if err != nil {
		return authorizationResult{Predicate: INVALID}
	}

	// Rebind as the read only user for any further queries
	if err = l.Bind(os.Getenv("LDAP_USER"), os.Getenv("LDAP_PASSWORD")); err != nil {
		panic(err)
	}

	// Check basic access rights
	isAccessMember := isGroupMember(l, user.DN, os.Getenv("LDAP_ACCESS_GROUP"))
	isAdminMember := isGroupMember(l, user.DN, os.Getenv("LDAP_ADMIN_GROUP"))
	if !isAccessMember && !isAdminMember {
		return authorizationResult{Predicate: DENIED}
	}

	// At this point, the user has proven basic access authorization (i.e.: Predicate: GRANTED)
	//
	// Check further permissions
	permissions := permissions{
		Admin: isAdminMember,
	}
	userEntry := userEntry{
		ID:          getID(user),
		DisplayName: getDisplayName(user),
		Permissions: &permissions,
	}

	return authorizationResult{
		Predicate: GRANTED,
		UserEntry: &userEntry,
	}
}

// ListUsers lists all users with basic access rights and their permissions
func ListUsers() []userEntry {
	if os.Getenv("LDAP_URL") == "" {
		return []userEntry{}
	}
	l := connectReadonly()
	defer l.Close()

	// Get relevant groups
	accessGroupMembers := getGroupMembers(l, os.Getenv("LDAP_ACCESS_GROUP"))
	adminGroupMembers := getGroupMembers(l, os.Getenv("LDAP_ADMIN_GROUP"))

	// Get all users
	users := getLdapUserEntries(l)

	userEntries := make([]userEntry, 0)
	for _, user := range users {
		isAccessGroupMember := false
		for _, userDn := range accessGroupMembers {
			if user.DN == userDn {
				isAccessGroupMember = true
				break
			}
		}
		isAdminGroupMember := false
		for _, userDn := range adminGroupMembers {
			if user.DN == userDn {
				isAdminGroupMember = true
				break
			}
		}
		if !isAccessGroupMember && !isAdminGroupMember {
			continue
		}
		permissions := permissions{
			Admin: isAdminGroupMember,
		}
		userEntry := userEntry{
			ID:          getID(user),
			DisplayName: getDisplayName(user),
			Permissions: &permissions,
		}
		userEntries = append(userEntries, userEntry)
	}
	return userEntries
}

func getGroupMembers(l *ldap.Conn, groupName string) []string {
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAP_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(cn=%s)(objectClass=group))", ldap.EscapeFilter(groupName)),
		[]string{"member"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}
	if len(sr.Entries) != 1 {
		panic("ldap group not found: " + groupName)
	}
	return sr.Entries[0].GetAttributeValues("member")
}

// connectReadonly connects to the LDAP server and binds with readonly
// credentials.
//
// Users are responsible to close the connection with `l.Close()` afterwards.
func connectReadonly() *ldap.Conn {
	l, err := ldap.DialURL(os.Getenv("LDAP_URL"))
	if err != nil {
		panic(err)
	}

	// Reconnect with TLS
	if err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		l.Close()
		panic(err)
	}

	// First bind with a read only user
	if err := l.Bind(os.Getenv("LDAP_USER"), os.Getenv("LDAP_PASSWORD")); err != nil {
		l.Close()
		panic(err)
	}
	return l
}

// isGroupMember returns `true` if the given user is member of the given group.
//
// `l` should be an open LDAP connection with readonly access.
func isGroupMember(l *ldap.Conn, userDn string, groupCn string) bool {
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAP_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(member=%s)(objectClass=group)(cn=%s))", ldap.EscapeFilter(userDn), ldap.EscapeFilter(groupCn)),
		[]string{},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}
	return len(sr.Entries) == 1
}

// getLdapUserEntry searches for a user with the given filter.
//
// It requests all attributes needed to create a userEntry object.
//
// `l` should be an open LDAP connection with readonly access.
//
// Returns nil when the user could not be found.
func getLdapUserEntry(l *ldap.Conn, filter string) *ldap.Entry {
	config := getConfiguration()
	attributes := []string{"dn", config.IDAttribute, config.DisplayNameAttribute}
	return getLdapUserEntryWithAttributes(l, filter, attributes)
}

func getLdapUserEntries(l *ldap.Conn) []*ldap.Entry {
	config := getConfiguration()
	attributes := []string{"dn", config.IDAttribute, config.DisplayNameAttribute}
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAP_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=organizationalPerson)",
		attributes,
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}
	return sr.Entries
}

// getLdapUserEntryWithAttributes searches for a user with the given filter and
// attributes.
//
// `l` should be an open LDAP connection with readonly access.
//
// Returns nil when the user could not be found.
func getLdapUserEntryWithAttributes(l *ldap.Conn, filter string, attributes []string) *ldap.Entry {
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("LDAP_BASE_DN"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 1, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)%s)", filter),
		attributes,
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}
	if len(sr.Entries) != 1 {
		return nil
	}
	return sr.Entries[0]
}

func getFilter(key string, value string) string {
	if value == "" {
		panic("called getFilter with empty value")
	}
	return fmt.Sprintf("(%s=%s)", key, ldap.EscapeFilter(value))
}

func getConfiguration() ldapConfiguration {
	switch ldapType := os.Getenv("LDAP_CONFIG"); ldapType {
	case "active-directory":
		return ldapConfiguration{
			IDAttribute:          "objectGUID",
			IDAttributeIsBinary:  true,
			UsernameAttribute:    "sAMAccountName",
			EmailAttribute:       "mail",
			DisplayNameAttribute: "displayName",
		}
	default:
		return ldapConfiguration{
			IDAttribute:          "uid",
			IDAttributeIsBinary:  false,
			UsernameAttribute:    "uid",
			EmailAttribute:       "mail",
			DisplayNameAttribute: "cn",
		}
	}
}

// getID returns a value to be used as unique ID.
func getID(entry *ldap.Entry) string {
	config := getConfiguration()
	if config.IDAttributeIsBinary {
		rawValue := entry.GetRawAttributeValue(config.IDAttribute)
		return base64.StdEncoding.EncodeToString(rawValue)
	} else {
		return entry.GetAttributeValue(config.IDAttribute)
	}
}

// getIDFilter returns a filter string to be used in an LDAP search for an
// ID obtained with GetID.
func getIDFilter(id string) string {
	config := getConfiguration()
	if config.IDAttributeIsBinary {
		bytes, err := base64.StdEncoding.DecodeString(id)
		if err != nil {
			panic(err)
		}
		return getFilter(config.IDAttribute, string(bytes))
	} else {
		return getFilter(config.IDAttribute, id)
	}
}

// getUserNameFilter returns a filter string to be used in an LDAP search
// for a username used in a login form.
func getUserNameFilter(username string) string {
	config := getConfiguration()
	return getFilter(config.UsernameAttribute, username)
}

func getDisplayName(user *ldap.Entry) string {
	if user == nil {
		return ""
	}
	config := getConfiguration()
	return user.GetAttributeValue(config.DisplayNameAttribute)
}
