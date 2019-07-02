package frapi

import "github.com/bjorge/friendlyreservations/models"

// MemberSchema is the gql schema for member requests
var MemberSchema = `
	schema {
		query: Query
		mutation: Mutation
	}
	
	# The query type, represents all of the entry points into our object graph
	type Query {
		# get info about a property
		property(id: String!): Property
	}

	# The mutation type, represents all updates we can make to our data
	type Mutation {
		# Create a reservation.
		createReservation(propertyId: String!, input: NewReservationInput!) : Property
		cancelReservation(propertyId: String!, forVersion: Int!, reservationId: String!, adminRequest: Boolean) : Property
		# Accept or reject an invitation to join a property.
		acceptInvitation(propertyId: String!, input: AcceptInvitationInput!) : Property
		# update membership status
		updateMembershipStatus(propertyId: String!, input: UpdateMembershipInput!) : Property
	}

	# QUERY RESULTS

	# all information about a property
	type Property {
		propertyId: String!
		eventVersion: Int!
		createDateTime: String!
		reservations(userId: String, reservationId: String, order: OrderDirection = ASCENDING): [Reservation]!
		restrictions(restrictionId: String, maxVersion: Int): [RestrictionRecord]!
		settings(maxVersion: Int): Settings!
		users(userId: String, email: String, maxVersion: Int): [User!]!
		me: User!
		ledgers(userId: String, last: Int, reverse: Boolean): [Ledger]!
		# ranges of dates that are disabled for the calendar view
		notifications(userId: String, reverse: Boolean): [Notification]!
		contents: [Content]!
		membershipStatusConstraints(userId: String): [MembershipStatusConstraints]!
		newReservationConstraints(userId: String, userType: ConstraintsUserType!): NewReservationConstraints!
		cancelReservationConstraints(userId: String, userType: ConstraintsUserType!): CancelReservationConstraints!

	}


` + models.AcceptInvitationInputGQL + models.NewReservationInputGQL + settingsGQL + reservationGQL + restrictionGQL + userGQL + ledgerQueryGQL + notificationGQL + contentGQL + membershipStatusConstraintsGQL + reservationConstraintsGQL + cancelReservationConstraintsGQL + models.UpdateMembershipStatusInputGQL
