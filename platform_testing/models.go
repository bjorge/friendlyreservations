package platformtesting

type testEvent1 struct {
	Value        int
	VersionValue int
}

func (r *testEvent1) GetEventVersion() int {
	return r.VersionValue
}

func (r *testEvent1) SetEventVersion(Version int) {
	r.VersionValue = Version
}

type testEvent2 struct {
	Value        string
	VersionValue int
}

func (r *testEvent2) GetEventVersion() int {
	return r.VersionValue
}

func (r *testEvent2) SetEventVersion(Version int) {
	r.VersionValue = Version
}

type testEvent3 struct {
	Value        bool
	VersionValue int
}

func (r *testEvent3) GetEventVersion() int {
	return r.VersionValue
}

func (r *testEvent3) SetEventVersion(Version int) {
	r.VersionValue = Version
}
