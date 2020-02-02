package frapi

import (
	"context"
	"testing"
)

func TestExportCSV(t *testing.T) {

	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	userID := me.UserID()

	// Create reservation
	t.Logf("TestExportCSV: add a reservation")
	property, _ = createReservation(ctx, t, resolver, property, userID, today.AddDays(1).ToString(), today.AddDays(3).ToString())

	t.Logf("TestExportCSV: add a payment")
	paymentAmount := int32(200)
	property = createPayment(ctx, t, resolver, property, paymentAmount, true, property.EventVersion())

	t.Logf("TestExportCSV: create the csv files")
	msg, err := resolver.exportCSVInternal(ctx, property, me)
	if err != nil {
		t.Fatal(err)
	}

	if len(msg.Attachments) != 2 {
		t.Fatalf("expected 2 attachments")
	}

	t.Logf("ledgers csv contents:\n%+v", string(msg.Attachments[0].Data))
	t.Logf("reservations csv contents:\n%+v", string(msg.Attachments[1].Data))

	// t.Fatal("testing")
}
