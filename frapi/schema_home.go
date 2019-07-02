package frapi

// HomeSchema is the gql schema for choosing a property
var HomeSchema = `
	schema {
		query: Query
		mutation: Mutation
	}
	
	# The query type, represents all of the entry points into our object graph
	type Query {
		# normal users query which properties they have
		properties(): [Property]!
		# return a logout url passing in dest for final destination after the logout
		logoutURL(dest: String!): String
		# settings constraints for creating a new property
		updateSettingsConstraints: UpdateSettingsConstraints!
		# user constraints for creating a new property
		updateUserConstraints: UpdateUserConstraints!

	}

	# The mutation type, represents all updates we can make to our data
	type Mutation {
		# create a new property
		createProperty(input: NewPropertyInput!): Property
		importProperty(): String!
	}

	# MUTATION INPUT
	input NewPropertyInput {
		propertyName: String!
		currency: Currency!
		memberRate: Int!
		allowNonMembers: Boolean!
		nonMemberRate: Int!
		isMember: Boolean!
		nickname: String!
		timezone: String!
	}
	
	# QUERY RESULTS

	# all information about a property
	type Property {
		propertyId: String!
		eventVersion: Int!
		createDateTime: String!
		settings(maxVersion: Int): Settings!
		me: User!

	}

` + settingsGQL + userGQL + settingsConstraintsGQL + updateUserConstraintsGQL
