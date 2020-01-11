package utilities

// SetTestSystemUser is called to setup the default system user during testing
func SetTestSystemUser(email string) {
	SystemEmail = email
}

// SetTestUser is used to setup the logged in user during testing
func SetTestUser(email string) {
	TestUserEmail = email
}

// SetAllowCreateProperty is called to allow creating a property during testing
func SetAllowCreateProperty() {
	AllowNewProperty = true
}
