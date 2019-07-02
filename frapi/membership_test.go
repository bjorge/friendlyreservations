package frapi

import (
	"context"
	"testing"

	"github.com/bjorge/friendlyreservations/frdate"
)

func TestMembershipStatusBasic(t *testing.T) {
	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	userID := me.UserID()

	t.Log("TestMembershipStatusBasic: create a current membership")
	property, _, _ = createMembershipRestriction(ctx, t, resolver, property, today)

	t.Log("TestMembershipStatusBasic: get membership status OPEN")
	usersStatus, err := property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})

	if err != nil {
		t.Fatal(err)
	}

	if len(usersStatus) != 1 {
		t.Fatalf("expect 1 user status")
	}
}

func TestMembershipStatus(t *testing.T) {
	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	// ledgers, _ = property.Ledgers(&ledgersArgs{Reverse: &bTrue, UserID: &userID})
	// t.Logf("ledger balance is: %v", ledgers[0].Records()[0].balanceInternal().Raw())

	userID := me.UserID()

	t.Log("TestMembershipStatus: create a current membership")
	property, _, _ = createMembershipRestriction(ctx, t, resolver, property, today)

	t.Log("TestMembershipStatus: get membership status OPEN")
	usersStatus, _ := property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})

	daysFromNow := usersStatus[0].Memberships()[0].membership.inDate.Sub(today)
	frdate.TestTimeOffsetDays = &daysFromNow
	expectedStatus(t, usersStatus[0].Memberships()[0], OPEN, true, true, 0)

	t.Log("TestMembershipStatus: update membership to purchased")
	property = updateMembership(ctx, t, resolver, property, usersStatus[0].Memberships()[0].Info().RestrictionID(), me.UserID(), true, "update current")
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], PURCHASED, false, true, 0)

	t.Log("TestMembershipStatus: get last membership status for user")
	maxVersion := property.EventVersion() - 1
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{
		UserID:     &userID,
		MaxVersion: &maxVersion,
	})
	expectedStatus(t, usersStatus[0].Memberships()[0], OPEN, true, true, 0)

	t.Log("TestMembershipStatus: create a second membership period")
	property, _, _ = createMembershipRestriction(ctx, t, resolver, property, today.AddDays(365))
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{
		UserID: &userID,
	})

	if len(usersStatus[0].Memberships()) != 2 {
		t.Fatalf("expected 2 memberships")
	}

	// the first membership should be OPEN and the second one should be PURCHASED
	t.Log("TestMembershipStatus: first record open but no changes allowed")
	expectedStatus(t, usersStatus[0].Memberships()[0], OPEN, false, true, 0)
	t.Log("TestMembershipStatus: second record purchased and only opt out allowed")
	expectedStatus(t, usersStatus[0].Memberships()[1], PURCHASED, false, true, 0)

	t.Log("TestMembershipStatus: check reservation count")
	property, _ = createReservation(ctx, t, resolver, property, userID, today.AddDays(daysFromNow+1).ToString(), today.AddDays(daysFromNow+2).ToString())
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})

	t.Log("TestMembershipStatus: check reservation count 0")
	expectedStatus(t, usersStatus[0].Memberships()[0], OPEN, false, true, 0)
	t.Log("TestMembershipStatus: check reservation count 1")
	expectedStatus(t, usersStatus[0].Memberships()[1], PURCHASED, false, false, 1)

	t.Log("TestMembershipStatus: create a second reservation")
	property, _ = createReservation(ctx, t, resolver, property, userID, today.AddDays(daysFromNow+3).ToString(), today.AddDays(daysFromNow+4).ToString())
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	t.Log("TestMembershipStatus: check reservation count 2")
	expectedStatus(t, usersStatus[0].Memberships()[1], PURCHASED, false, false, 2)

	t.Log("TestMembershipStatus: update second membership to purchased")
	property = updateMembership(ctx, t, resolver, property, usersStatus[0].Memberships()[0].Info().RestrictionID(), me.UserID(), true, "update current")
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], PURCHASED, false, true, 0)

	t.Log("TestMembershipStatus: create a reservation for second membership period")
	daysFromNow = usersStatus[0].Memberships()[0].membership.inDate.Sub(today)
	frdate.TestTimeOffsetDays = &daysFromNow
	property, reservations := createReservation(ctx, t, resolver, property, userID, today.AddDays(daysFromNow+1).ToString(), today.AddDays(daysFromNow+2).ToString())
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], PURCHASED, false, false, 1)

	t.Log("TestMembershipStatus: cancel the reservation during second membership period")
	property, _ = cancelReservation(ctx, t, resolver, property, reservations[0].ReservationID(), false, property.EventVersion())
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], PURCHASED, false, true, 0)

	t.Log("TestMembershipStatus: opt out second membership")
	property = updateMembership(ctx, t, resolver, property, usersStatus[0].Memberships()[0].Info().RestrictionID(), me.UserID(), false, "update current")
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], OPTOUT, true, false, 0)

	t.Log("TestMembershipStatus: test prepay period positive")
	daysFromNow = usersStatus[0].Memberships()[0].membership.prePayStartDate.Sub(today) + 1
	frdate.TestTimeOffsetDays = &daysFromNow
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], OPTOUT, true, false, 0)

	t.Log("TestMembershipStatus: test prepay period negative")
	daysFromNow = usersStatus[0].Memberships()[0].membership.prePayStartDate.Sub(today) - 1
	frdate.TestTimeOffsetDays = &daysFromNow
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], OPTOUT, false, false, 0)

	t.Log("TestMembershipStatus: low balance blocks purchase")
	daysFromNow = usersStatus[0].Memberships()[0].membership.prePayStartDate.Sub(today) + 1
	frdate.TestTimeOffsetDays = &daysFromNow
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], OPTOUT, true, false, 0)
	property, _ = createReservation(ctx, t, resolver, property, userID, today.AddDays(daysFromNow+1).ToString(), today.AddDays(daysFromNow+100).ToString())
	usersStatus, _ = property.MembershipStatusConstraints(&membershipStatusConstraintsArgs{UserID: &userID})
	expectedStatus(t, usersStatus[0].Memberships()[0], OPTOUT, false, false, 1)

	t.Logf("TestMembershipStatus: test query rather than constraint")
	lastStatus := usersStatus[0].Memberships()[0]
	state, err := property.membershipInfo(lastStatus.Info().RestrictionID(), userID, property.EventVersion())
	if err != nil {
		t.Fatal(err)
	}
	if state != OPTOUT {
		t.Fatalf("expected OPTOUT from query but got %+v", state)
	}

}

func expectedStatus(t *testing.T, membership *MembershipRecordResolver, status MembershipState, purchaseAllowed bool, optOutAllowed bool, reservationCount int32) {
	if membership.Status() != status {
		t.Fatalf("membership status expected: %v actual: %v", status, membership.Status())
	}

	if membership.PurchaseAllowed() != purchaseAllowed {
		t.Fatalf("membership purchase allowed expected: %v actual: %v", purchaseAllowed, membership.PurchaseAllowed())
	}

	if membership.OptOutAllowed() != optOutAllowed {
		t.Fatalf("membership opt out allowed expected: %v actual: %v", optOutAllowed, membership.OptOutAllowed())
	}

	if membership.ReservationCount() != reservationCount {
		t.Fatalf("membership reservation count expected: %v actual: %v", reservationCount, membership.ReservationCount())
	}
}
