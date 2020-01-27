package frapi

import "github.com/bjorge/friendlyreservations/models"

// AdminSchema is the gql schema for admin requests
var AdminSchema = `
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
		# create reservation
		createReservation(propertyId: String!, input: NewReservationInput!) : Property
		# cancel reservation
		cancelReservation(propertyId: String!, forVersion: Int!, reservationId: String!, adminRequest: Boolean) : Property
		# create restriction
		createRestriction(propertyId: String!, input: NewRestrictionInput!) : Property
		# create user
		createUser(propertyId: String!, input: NewUserInput!) : Property
		# update user
		updateUser(propertyId: String!, userId: String!, input: UpdateUserInput!) : Property
		# update system user
		updateSystemUser(propertyId: String!, userId: String!, input: UpdateSystemUserInput!) : Property
		# update user balance
		updateBalance(propertyId: String!, input: UpdateBalanceInput!) : Property
		# mark notification read
		notificationRead(propertyId: String!, notificationId: String!) : Property
		# create content
		createContent(propertyId: String!, input: NewContentInput!) : Property
		# accept or reject an invitation to join a property
		acceptInvitation(propertyId: String!, input: AcceptInvitationInput!) : Property
		# update settings
		updateSettings(propertyId: String!, input: UpdateSettingsInput!) : Property
		# update membership
		updateMembershipStatus(propertyId: String!, input: UpdateMembershipInput!) : Property
		export(propertyId: String!) : Property
		deleteProperty(propertyId: String!) : Boolean!

	}

	# QUERY RESULTS

	# all information about a property
	type Property {
		propertyId: String!
		eventVersion: Int!
		createDateTime: String!
		reservations(userId: String, reservationId: String, order: OrderDirection = ASCENDING): [Reservation]!
		settings(maxVersion: Int): Settings!
		users(userId: String, email: String): [User!]!
		me: User!
		restrictions(restrictionId: String, maxVersion: Int): [RestrictionRecord]!
		ledgers(userId: String, last: Int, reverse: Boolean): [Ledger]!
		notifications(userId: String, reverse: Boolean): [Notification]!
		contents: [Content]!
		updateSettingsConstraints: UpdateSettingsConstraints!
		membershipStatusConstraints(userId: String): [MembershipStatusConstraints]!
		newReservationConstraints(userId: String, userType: ConstraintsUserType!): NewReservationConstraints!
		cancelReservationConstraints(userId: String, userType: ConstraintsUserType!): CancelReservationConstraints!
		updateUserConstraints(userId: String): UpdateUserConstraints!
		updateBalanceConstraints(): UpdateBalanceConstraints!

	}


` + models.NewRestrictionInputGQL + models.BlackoutRestrictionInputGQL + models.MembershipRestrictionInputGQL + models.UpdateSettingsInputGQL + models.AcceptInvitationInputGQL + models.NewReservationInputGQL + settingsGQL + reservationGQL + userGQL + ledgerQueryGQL + membershipStatusConstraintsGQL + restrictionGQL + notificationGQL + contentGQL + reservationConstraintsGQL + settingsConstraintsGQL + updateUserConstraintsGQL + cancelReservationConstraintsGQL + models.LedgerMutationGQL + models.NewUserInputGQL + models.UpdateUserInputGQL + updateBalanceConstraintsGQL + models.NewContentInputGQL + models.UpdateMembershipStatusInputGQL + models.UpdateSystemUserInputGQL
