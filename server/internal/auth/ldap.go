package auth

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

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

var LDAP_BASE_DN string
var LDAP_ACCESS_GROUP_CN string
var LDAP_ADMIN_GROUP_CN string
var LDAP_ACCESS_GROUP_DN string
var LDAP_ADMIN_GROUP_DN string

func TestConnection() {
	l := connectReadonly()
	defer l.Close()
	LDAP_BASE_DN = os.Getenv("LDAP_BASE_DN")
	LDAP_ACCESS_GROUP_CN = os.Getenv("LDAP_ACCESS_GROUP")
	LDAP_ADMIN_GROUP_CN = os.Getenv("LDAP_ADMIN_GROUP")
	LDAP_ACCESS_GROUP_DN = getGroupDN(l, LDAP_ACCESS_GROUP_CN)
	LDAP_ADMIN_GROUP_DN = getGroupDN(l, LDAP_ADMIN_GROUP_CN)
	getGroupMembers(l, LDAP_ACCESS_GROUP_CN)
}

func GetDisplayName(userID string) string {
	l := connectReadonly()
	defer l.Close()
	user := getLdapUserEntry(l, getIDFilter(userID))
	return getDisplayName(user)
}

func GetMailAddress(userID string) (string, error) {
	l := connectReadonly()
	defer l.Close()
	user := getLdapUserEntryWithAttributes(l, getIDFilter(userID), []string{"dn", "mail"})
	if user == nil {
		return "", fmt.Errorf("ldap: failed to find user with ID %s", userID)
	}
	mail := user.GetAttributeValue("mail")
	if mail == "" {
		return "", fmt.Errorf("ldap: user with ID %s has no email address", userID)
	}
	return mail, nil
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
	isAccessMember := isGroupMember(l, user.DN, LDAP_ACCESS_GROUP_CN)
	isAdminMember := isGroupMember(l, user.DN, LDAP_ADMIN_GROUP_CN)
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
	l := connectReadonly()
	defer l.Close()
	accessUserDns, err := getGroupMembersRecursive(l, LDAP_ACCESS_GROUP_DN, make(map[string]bool))
	if err != nil {
		panic(err)
	}
	adminUserDns, err := getGroupMembersRecursive(l, LDAP_ADMIN_GROUP_DN, make(map[string]bool))
	if err != nil {
		panic(err)
	}
	var filteredAccessUserDns []string
	for _, userDN := range accessUserDns {
		if !slices.Contains(adminUserDns, userDN) {
			filteredAccessUserDns = append(filteredAccessUserDns, userDN)
		}
	}
	accessUsers := getUserEntries(l, filteredAccessUserDns)
	userEntries := make([]userEntry, 0)
	for _, user := range accessUsers {
		permissions := permissions{
			Admin: false,
		}
		userEntry := userEntry{
			ID:          getID(user),
			DisplayName: getDisplayName(user),
			Permissions: &permissions,
		}
		userEntries = append(userEntries, userEntry)
	}
	adminUsers := getUserEntries(l, adminUserDns)
	for _, user := range adminUsers {
		permissions := permissions{
			Admin: true,
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

func getGroupMembers(l *ldap.Conn, groupCn string) []string {
	searchRequest := ldap.NewSearchRequest(
		LDAP_BASE_DN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(cn=%s)(objectClass=group))", ldap.EscapeFilter(groupCn)),
		[]string{"member"},
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}
	if len(sr.Entries) != 1 {
		panic("ldap group not found: " + groupCn)
	}
	return sr.Entries[0].GetAttributeValues("member")
}

func getGroupMembersRecursive(l *ldap.Conn, groupDN string, visited map[string]bool) ([]string, error) {
	// prevent infinite loops
	if visited[groupDN] {
		return nil, nil
	}
	visited[groupDN] = true
	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=group)",
		[]string{"member"},
		nil,
	)
	result, err := l.Search(searchRequest)
	if err != nil || len(result.Entries) == 0 {
		return nil, err
	}
	var allMembers []string
	members := result.Entries[0].GetAttributeValues("member")
	for _, dn := range members {
		// Lookup this DN to see if it's a user or a group
		entry, err := getEntryByDN(l, dn)
		if err != nil {
			continue
		}
		if isGroup(entry) {
			nestedMembers, _ := getGroupMembersRecursive(l, dn, visited)
			allMembers = append(allMembers, nestedMembers...)
		} else {
			allMembers = append(allMembers, dn)
		}
	}
	return allMembers, nil
}

func getEntryByDN(l *ldap.Conn, dn string) (*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		dn,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0, 0, false,
		"(objectClass=*)",
		[]string{"objectClass"},
		nil,
	)
	result, err := l.Search(searchRequest)
	if err != nil || len(result.Entries) == 0 {
		return nil, err
	}
	return result.Entries[0], nil
}

func isGroup(entry *ldap.Entry) bool {
	for _, oc := range entry.GetAttributeValues("objectClass") {
		if strings.ToLower(oc) == "group" {
			return true
		}
	}
	return false
}

func getUserEntries(l *ldap.Conn, userDnList []string) []*ldap.Entry {
	var filters []string
	for _, dn := range userDnList {
		filters = append(
			filters,
			fmt.Sprintf(
				"(&(objectClass=organizationalPerson)(distinguishedName=%s))",
				ldap.EscapeFilter(dn)),
		)
	}
	filter := fmt.Sprintf("(|%s)", strings.Join(filters, ""))
	config := getConfiguration()
	attributes := []string{"dn", config.IDAttribute, config.DisplayNameAttribute}
	searchRequest := ldap.NewSearchRequest(
		LDAP_BASE_DN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter,
		attributes,
		nil,
	)
	sr, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}
	return sr.Entries
}

// connectReadonly connects to the LDAP server and binds with readonly
// credentials.
//
// Users are responsible to close the connection with `l.Close()` afterwards.
func connectReadonly() *ldap.Conn {
	url := os.Getenv("LDAP_URL")
	skipTLSVerify := os.Getenv("LDAP_TLS_INSECURE_SKIP_VERIFY") == "true"
	l, err := ldap.DialURL(
		url,
		ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: skipTLSVerify}),
	)
	if err != nil {
		panic(err)
	}

	if cs, ok := l.TLSConnectionState(); !ok || !cs.HandshakeComplete {
		// Reconnect with TLS
		serverName := strings.TrimPrefix(url, "ldap://")
		reg := regexp.MustCompile(`:[0-9]+$`)
		serverName = reg.ReplaceAllString(serverName, "")
		if err = l.StartTLS(&tls.Config{InsecureSkipVerify: skipTLSVerify, ServerName: serverName}); err != nil {
			l.Close()
			panic(err)
		}
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
		LDAP_BASE_DN,
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

// getLdapUserEntryWithAttributes searches for a user with the given filter and
// attributes.
//
// `l` should be an open LDAP connection with readonly access.
//
// Returns nil when the user could not be found.
func getLdapUserEntryWithAttributes(l *ldap.Conn, filter string, attributes []string) *ldap.Entry {
	searchRequest := ldap.NewSearchRequest(
		LDAP_BASE_DN,
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

func getGroupDN(l *ldap.Conn, groupCn string) string {
	searchRequest := ldap.NewSearchRequest(
		LDAP_BASE_DN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 1, 0, false,
		fmt.Sprintf("(&(objectClass=group)(cn=%s))", groupCn),
		[]string{"distinguishedName"},
		nil,
	)
	searchResult, err := l.Search(searchRequest)
	if err != nil {
		panic(err)
	}
	if len(searchResult.Entries) != 1 {
		panic("more than one entry for group")
	}
	return searchResult.Entries[0].DN
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
