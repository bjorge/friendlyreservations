package frapi

import (
	"context"
	"os"
	"testing"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/platform"
	"github.com/bjorge/friendlyreservations/utilities"

	"github.com/bjorge/friendlyreservations/models"
)

// TestLoadExport loads up the database and then exports it
func TestLoadExport(t *testing.T) {

	property, ctx, resolver, me, _ := initAndCreateTestProperty(context.Background(), t)

	// now let's create a bunch of updates
	userID := me.UserID()
	version := property.EventVersion()
	forVersion := version
	events := []platform.VersionedEvent{}
	numEvents := int32(100000)
	for i := int32(0); i < numEvents; i++ {
		updateBalance := createUpdateBalanceRecord(userID, 1, forVersion)
		events = append(events, updateBalance)
		forVersion++
	}

	property, err := commitChanges(ctx, property.PropertyID(), version, events...)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := resolver.exportInternal(ctx, property, me)
	if err != nil {
		t.Fatal(err)
	}

	if len(msg.Attachments) != 1 {
		t.Fatalf("expected an attachment")
	}

	fi, err := os.Create("/tmp/exportloadtest.bin")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fi.Write(msg.Attachments[0].Data)
	if err != nil {
		t.Fatal(err)
	}
	defer fi.Close()

	//resolver.Export(ctx, &struct{ PropertyID string }{PropertyID: property.PropertyID()})

	importedPropertyID, err := internalImportProperty(ctx, msg.Attachments[0].Data)

	if err != nil {
		t.Fatal(err)
	}

	if property.PropertyID() == importedPropertyID {
		t.Fatalf("wrong property id")
	}

	importedProperty, err := resolver.Property(ctx, &struct{ ID string }{ID: importedPropertyID})

	if err != nil {
		t.Fatal(err)
	}

	settings, err := property.Settings(&settingsArgs{})

	importedSettings, err := importedProperty.Settings(&settingsArgs{})

	if settings.PropertyName() != importedSettings.PropertyName() {
		t.Fatalf("property not imported correctly")
	}

	args := &ledgersArgs{UserID: &userID}
	//args.UserID := &userID
	ledgers, err := property.Ledgers(args)
	if err != nil {
		t.Fatal(err)
	}
	records := ledgers[0].Records()
	balance := records[len(records)-1].balanceInternal().Raw()
	if balance != numEvents {
		t.Fatalf("balance is %+v, expected %+v", balance, numEvents)
	}
}

func createUpdateBalanceRecord(userID string, amount int32, version int32) *models.UpdateBalanceInput {
	updateBalance := &models.UpdateBalanceInput{}
	updateBalance.Amount = amount
	updateBalance.AuthorUserId = userID
	updateBalance.CreateDateTime = frdate.CreateDateTimeUTC()
	updateBalance.PaymentId = utilities.NewGUID()
	updateBalance.Description = "load test"
	updateBalance.ForVersion = version
	updateBalance.Increase = true
	updateBalance.UpdateForUserId = userID
	return updateBalance
}
