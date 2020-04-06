package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/bjorge/friendlyreservations/frapi"
	"github.com/bjorge/friendlyreservations/logger"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/platform"
)

var log = logger.New()
var names = []string{"Liam", "Emma", "Noah", "Olivia", "William", "Ava", "James",
	"Isabella", "Oliver", "Sophia", "Benjamin", "Mia", "Elijah", "Charlotte", "Lucas",
	"Amelia", "Mason", "Evelyn", "Logan", "Abigail"}
var propertyName = "Beach House"
var systemName = "System"
var nonMemberName = "Pat"
var nonMemberInfo = "contact at email pat@test.com"

var admins = []string{}

func main() {

	fmt.Println("Start anonymization")

	data, err := ioutil.ReadFile("frdatav1.bin")

	if err != nil {
		panic(err)
	}

	// decode the original backup
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	decoded := &frapi.PropertyExport{}
	err = dec.Decode(decoded)
	if err != nil {
		panic(err)
	}

	// create the anonymized backup
	anonymizedEvents := []platform.VersionedEvent{}
	anonymizedEmailMap := make(map[string]string)

	// for emailID, email := range decoded.EmailMap {
	// 	log.LogDebugf("name: %v value %v", emailID, email)
	// }

	userIterator := 0
	checkIterator := 100

	for _, versionedEvent := range decoded.Events {

		switch event := versionedEvent.(type) {

		case *models.NewUserInput:
			// log.LogDebugf("models.NewUserInput")
			if event.IsSystem {
				event.Nickname = systemName
				anonymizedEmailMap[event.EmailId] = strings.ToLower(systemName) + "@test.com"
			} else {
				event.Nickname = names[userIterator%20]
				anonymizedEmailMap[event.EmailId] = strings.ToLower(names[userIterator%20]) + strconv.Itoa(userIterator) + "@test.com"
				if event.IsAdmin {
					admins = append(admins, anonymizedEmailMap[event.EmailId])
					// log.LogDebugf("admin email: " + anonymizedEmailMap[event.EmailId])
				}
				userIterator++
			}
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.NewVersionEvent:
			// log.LogDebugf("models.NewVersionEvent")
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.NewPropertyInput:
			// log.LogDebugf("models.NewPropertyInput")
			event.PropertyName = propertyName
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.NewNotificationInput:
			// log.LogDebugf("models.NewNotificationInput")
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.NewRestrictionInput:
			// log.LogDebugf("models.NewRestrictionInput")
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.UpdateMembershipStatusInput:
			// comment := ""
			// if event.Comment != nil {
			// 	comment = *event.Comment
			// }
			// log.LogDebugf("models.UpdateMembershipStatusInput, comment is %v", comment)
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.NewReservationInput:
			// log.LogDebugf("models.NewReservationInput")
			if !event.Member {
				event.NonMemberName = &nonMemberName
				event.NonMemberInfo = &nonMemberInfo
			}
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.UpdateBalanceInput:
			// log.LogDebugf("models.UpdateBalanceInput")
			event.Description = "check number " + strconv.Itoa(checkIterator)
			checkIterator++
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.UpdateSettingsInput:
			// log.LogDebugf("models.UpdateSettingsInput")
			event.PropertyName = propertyName
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.UpdateUserInput:
			// log.LogDebugf("models.UpdateUserInput")
			event.Nickname = names[userIterator%20]
			anonymizedEmailMap[event.EmailId] = strings.ToLower(names[userIterator%20]) + strconv.Itoa(userIterator) + "@test.com"
			if event.IsAdmin {
				admins = append(admins, anonymizedEmailMap[event.EmailId])
				// log.LogDebugf("admin email: " + anonymizedEmailMap[event.EmailId])
			}
			userIterator++
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.AcceptInvitationInput:
			// log.LogDebugf("models.AcceptInvitationInput")
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.CancelReservationInput:
			// log.LogDebugf("models.CancelReservationInput")
			anonymizedEvents = append(anonymizedEvents, event)
		case *models.UpdateSystemUserInput:
			// log.LogDebugf("models.UpdateSystemUserInput")
			event.Nickname = systemName
			anonymizedEmailMap[event.EmailID] = strings.ToLower(systemName) + "@test.com"
			anonymizedEvents = append(anonymizedEvents, event)
		default:
			log.LogDebugf("BUG: need to anonymize this event")
			anonymizedEvents = append(anonymizedEvents, event)
		}
	}

	// for emailID, email := range anonymizedEmailMap {
	// 	log.LogDebugf("name: %v value %v", emailID, email)
	// }

	// package it up
	anonymizedExport := &frapi.PropertyExport{}
	anonymizedExport.Events = anonymizedEvents
	anonymizedExport.EmailMap = anonymizedEmailMap

	// encoded the array into a gob
	dataFile, err := os.Create("anonymized.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// serialize the data
	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(anonymizedExport)

	dataFile.Close()

	log.LogDebugf("wrote out anonymized.bin file")
	log.LogDebugf("admins are: %+v", admins)

}
