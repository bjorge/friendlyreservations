package frapi

import (
	"context"
	"errors"
	"testing"
)

func TestLedgerResolvers(t *testing.T) {

	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	userID := me.UserID()

	// get initial ledger records
	t.Logf("intial check ledger")
	checkLedger(ctx, t, property, userID, 1, startLedgerEvent, 0, 0)

	// Create reservation
	property, reservations := createReservation(ctx, t, resolver, property, userID, today.AddDays(1).ToString(), today.AddDays(3).ToString())

	// now there should be two records in the ledger
	reservationExpense := -1 * 2 * defaultPropertyInput.MemberRate
	t.Logf("ledger check after making reservation")
	record, _ := checkLedger(ctx, t, property, userID, 2, reservationLedgerEvent, reservationExpense, reservationExpense)

	// check the reservation reference
	if record.EventDateTime() != reservations[0].CreateDateTime() {
		t.Fatal(errors.New("TestLedgerResolver: wrong event"))
	}

	t.Logf("now add a payment")
	paymentAmount := int32(200)
	property = createPayment(ctx, t, resolver, property, paymentAmount, true, property.EventVersion())

	// check that the payment changed the balance
	// balance should now be the following:
	balance := reservationExpense
	balance += paymentAmount
	checkLedger(ctx, t, property, userID, 3, paymentLedgerEvent, balance, paymentAmount)

	// TODO: check the payment reference

	t.Logf("check single last ledger record")
	last := int32(1)
	userRecords, _ := property.Ledgers(&ledgersArgs{UserID: &userID, Last: &last})

	records := userRecords[0].Records()

	if len(records) != 1 {
		t.Fatalf("wrong number of records: %+v", len(records))
	}

	if records[0].Event() != paymentLedgerEvent {
		t.Fatalf("wrong event expected PAYMENT got %+v", records[0].Event())
	}

}

func TestLedgerReservation(t *testing.T) {
	property, ctx, resolver, me, todayDate := initAndCreateTestProperty(context.Background(), t)

	userID := me.UserID()

	checkin := todayDate.AddDays(1)
	checkout := checkin.AddDays(2)

	t.Log("create a single reservation")
	property, reservations := createReservation(ctx, t, resolver, property, userID, checkin.ToString(), checkout.ToString())

	reservationExpense := 2 * defaultPropertyInput.MemberRate
	checkLedger(ctx, t, property, userID, 2, reservationLedgerEvent, -1*reservationExpense, -1*reservationExpense)

	property, _ = cancelReservation(ctx, t, resolver, property, reservations[0].ReservationID(), false, property.EventVersion())
	checkLedger(ctx, t, property, userID, 3, cancelReservationLedgerEvent, 0, reservationExpense)

}

func TestLedgerDuplicatePayment(t *testing.T) {
	property, ctx, resolver, me, _ := initAndCreateTestProperty(context.Background(), t)

	userID := me.UserID()

	t.Logf("now add a payment")
	paymentAmount := int32(200)
	eventVersion := property.EventVersion()
	property = createPayment(ctx, t, resolver, property, paymentAmount, true, eventVersion)

	// check that the payment changed the balance
	// balance should now be the following:
	balance, amount := paymentAmount, paymentAmount
	checkLedger(ctx, t, property, userID, 2, paymentLedgerEvent, balance, amount)

	// add the payment again (with the same event version), make sure dup is suppressed
	t.Logf("now add a duplicate payment")
	if property.EventVersion() == eventVersion {
		t.Fatal("something wrong with returned version")
	}
	property = createPayment(ctx, t, resolver, property, paymentAmount, true, eventVersion)
	checkLedger(ctx, t, property, userID, 2, paymentLedgerEvent, balance, amount)

	// add a non-duplicate payment, make sure payment is added
	t.Logf("now add a non-duplicate payment")
	paymentAmount = int32(100)
	eventVersion = property.EventVersion()
	property = createPayment(ctx, t, resolver, property, paymentAmount, true, eventVersion)
	balance, amount = balance+paymentAmount, paymentAmount
	checkLedger(ctx, t, property, userID, 3, paymentLedgerEvent, balance, amount)
}

func TestLedgerMembershipUpdate(t *testing.T) {
	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	t.Log("TestLedgerMembershipUpdate test new membership for this year")
	property, membershipRestriction, restrictions := createMembershipRestriction(ctx, t, resolver, property, today)
	balance := int32(0)
	//balance := 0 - membershipRestriction.Amount
	amount := int32(0)
	checkLedger(ctx, t, property, me.UserID(), 1, startLedgerEvent, balance, amount)

	t.Log("TestLedgerMembershipUpdate pay membership")
	property = updateMembership(ctx, t, resolver, property, restrictions[0].RestrictionID(), me.UserID(), true, "set to true")

	balance -= membershipRestriction.Amount
	amount = -1 * membershipRestriction.Amount
	checkLedger(ctx, t, property, me.UserID(), 2, purchaseMembershipLedgerEvent, balance, amount)

	t.Log("TestLedgerMembershipUpdate optout membership")
	property = updateMembership(ctx, t, resolver, property, restrictions[0].RestrictionID(), me.UserID(), false, "set to false")

	balance += membershipRestriction.Amount
	amount = membershipRestriction.Amount
	checkLedger(ctx, t, property, me.UserID(), 3, optoutMembershipLedgerEvent, balance, amount)

}
