package db

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AgencyV1 struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `json:"name"`
	Abbreviation string             `json:"abbreviation"`
	// Prefix is the agency prefix as by xdomea.
	Prefix string `json:"prefix"`
	// Code is the agency code as by xdomea.
	Code string `json:"code"`
	// ContactEmail is the e-mail address to use to contact the agency.
	ContactEmail string `bson:"contact_email" json:"contactEmail"`
	// TransferDirURL contains the protocol, host, username and password needed to access a file share.
	// Possible values for the protocol are file, webdav, webdavs.
	// The username and password are optional.
	TransferDirURL string `json:"transferDirURL"`
	// Users are users responsible for processes of this Agency.
	Users        []string            `json:"users"`
	CollectionID *primitive.ObjectID `bson:"collection_id" json:"collectionId"`
}

func MigrateAgencies(ctx context.Context) error {
	filter := bson.M{
		"schema_version": bson.M{
			"$exists": false,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"schema_version": 1,
		},
	}
	coll := mongoDatabase.Collection("agencies")
	_, err := coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}
	var agenciesV1 []AgencyV1
	cursor, err := coll.Find(
		ctx,
		bson.M{
			"schema_version": 1,
		})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &agenciesV1)
	if err != nil {
		return err
	}
	var notMigrated []string
	for _, aV1 := range agenciesV1 {
		aV2, err := migrateAgencyV1(aV1)
		if err != nil {
			notMigrated = append(notMigrated, aV1.Name)
			continue
		}
		result, err := coll.ReplaceOne(ctx, bson.M{"_id": aV1.ID}, aV2)
		if err != nil {
			notMigrated = append(notMigrated, aV1.Name)
			continue
		}
		if result.ModifiedCount != 1 {
			notMigrated = append(notMigrated, aV1.Name)
		}
	}
	if len(notMigrated) != 0 {
		return fmt.Errorf(
			"migration failed for agencies: %s",
			strings.Join(notMigrated, ", "),
		)
	}
	return nil
}

func migrateAgencyV1(agencyV1 AgencyV1) (Agency, error) {
	transferDir := TransferDir{}
	url, err := url.Parse(agencyV1.TransferDirURL)
	if err != nil {
		log.Println(err)
		return Agency{}, err
	}
	transferDir.Protocol = TransferProtocol(url.Scheme)
	transferDir.Host = url.Host
	transferDir.Path = url.Path
	if url.User != nil {
		transferDir.User = url.User.Username()
		password, ok := url.User.Password()
		if ok {
			transferDir.Password = password
		}
	}
	agency := Agency{
		ID:            agencyV1.ID,
		SchemaVersion: 2,
		Name:          agencyV1.Name,
		Abbreviation:  agencyV1.Abbreviation,
		Prefix:        agencyV1.Prefix,
		Code:          agencyV1.Code,
		ContactEmail:  agencyV1.ContactEmail,
		Users:         agencyV1.Users,
		CollectionID:  agencyV1.CollectionID,
		TransferDir:   transferDir,
	}
	return agency, nil
}
