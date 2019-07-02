package frapi

import (
	"fmt"
)

// BUG(bjorge): change to real query accessible from gql,
// that way client can show for each membership all the users' status in a single table
// for now only used by ledger rollup

func (r *PropertyResolver) membershipInfo(restrictionID string, userID string, maxVersion int32) (MembershipState, error) {

	// rollup
	r.rollupMembershipStatus()

	versionedRollups := r.getRollups(&rollupArgs{maxVersion: &maxVersion, id: &restrictionID}, membershipStatusRollupType)

	if len(versionedRollups) == 0 {
		return OPEN, fmt.Errorf("restriction id is not valid")
	}

	rollup := versionedRollups[0].(*MembershipRollupRecord)

	if status, ok := rollup.Users[userID]; ok {
		return status, nil
	}

	return OPEN, nil
}
