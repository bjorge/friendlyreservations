package models

import (
	"encoding/gob"
)

// init intializes the data structures for gob serialization
func init() {
	// todo: just run once
	gob.Register(&NewVersionEvent{})
	gob.Register(&NewPropertyInput{})
	gob.Register(&NewReservationInput{})
	gob.Register(&CancelReservationInput{})
	gob.Register(&UpdateMembershipStatusInput{})
	gob.Register(&UpdateBalanceInput{})
	gob.Register(&UpdateSettingsInput{})
	gob.Register(&NewUserInput{})
	gob.Register(&UpdateUserInput{})
	gob.Register(&UpdateSystemUserInput{})
	gob.Register(&AcceptInvitationInput{})
	gob.Register(&NewRestrictionInput{})
	gob.Register(&NewNotificationInput{})
	gob.Register(&NotificationReadInput{})
	gob.Register(&NewContentInput{})

	gob.Register(&BlackoutRestriction{})
	gob.Register(&MembershipRestriction{})
}
