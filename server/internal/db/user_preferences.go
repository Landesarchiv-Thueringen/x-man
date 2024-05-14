package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserPreferences struct {
	// ID is the value of the LDAP attribute used as identifier.
	UserID string `bson:"user_id" json:"-"`
	// MessageEmailNotifications is the user's setting to receive e-mail notifications
	// regarding new messages from x-man.
	MessageEmailNotifications bool `bson:"message_email_notifications" json:"messageEmailNotifications"`
	// ReportByEmail is the user's setting to receive the generated report after
	// successfully archiving a message.
	ReportByEmail bool `bson:"report_by_email" json:"reportByEmail"`
	// ErrorEmailNotifications is a setting for users with administration
	// privileges to receive e-mails for new processing errors.
	ErrorEmailNotifications bool `bson:"error_email_notifications" json:"errorEmailNotifications"`
}

var defaultUserPreferences = UserPreferences{
	MessageEmailNotifications: true,
	ReportByEmail:             true,
	ErrorEmailNotifications:   false,
}

// FindUserPreferences returns the preferences for the given user or the default
// preferences if no entry could be found for this user.
func FindUserPreferences(ctx context.Context, userID string) UserPreferences {
	coll := mongoDatabase.Collection("user_preferences")
	filter := bson.D{{"user_id", userID}}
	var p UserPreferences
	err := coll.FindOne(ctx, filter).Decode(&p)
	if err == mongo.ErrNoDocuments {
		return defaultUserPreferences
	} else if err != nil {
		panic(err)
	}
	return p
}

// UpsertUserPreferences saves the preferences for the given user to the
// database.
//
// The entry for the user preferences is created if it does not exist yet.
func UpsertUserPreferences(p UserPreferences) {
	coll := mongoDatabase.Collection("user_preferences")
	filter := bson.D{{"user_id", p.UserID}}
	opts := options.Replace().SetUpsert(true)
	_, err := coll.ReplaceOne(context.Background(), filter, p, opts)
	if err != nil {
		panic(err)
	}
}
