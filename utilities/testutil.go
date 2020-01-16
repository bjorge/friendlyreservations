package utilities

// SetTestSystemUser is called to setup the default system user during testing
func SetTestSystemUser(email string) {
	SystemEmail = email
}

// SetAllowCreateProperty is called to allow creating a property during testing
func SetAllowCreateProperty() {
	AllowNewProperty = true
}
