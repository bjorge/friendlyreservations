package frapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bjorge/friendlyreservations/utilities"

	"github.com/bjorge/friendlyreservations/models"
)

func TestSettingsResolver(t *testing.T) {

	property, _, _, _, _ := initAndCreateTestProperty(context.Background(), t)

	settingsResolver, err := property.Settings(&settingsArgs{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("settings: %+v", settingsResolver.settings)

	if settingsResolver.memberRateInternal() != 4000 {
		t.Fatal(errors.New("wrong MemberRate"))
	}

	//call one more time to get cached value
	settingsResolver, err = property.Settings(&settingsArgs{})
	if err != nil {
		t.Fatal(err)
	}

}

func TestSettings(t *testing.T) {
	property, ctx, resolver, _, _ := initAndCreateTestProperty(context.Background(), t)

	firstEventVersion := property.EventVersion()

	settings, err := property.Settings(&settingsArgs{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("property name: %+v", settings.PropertyName())
	t.Logf("default member rate: %+v", settings.memberRateInternal())

	//t.Fatalf("exit")

	if settings.PropertyName() != "Test Property" {
		t.Fatalf("expected valid property name")
	}

	input := &models.UpdateSettingsInput{}

	testPropertyName := "Test Update Name"

	input.AllowNonMembers = settings.AllowNonMembers()
	input.BalanceReminderIntervalDays = settings.BalanceReminderIntervalDays()
	currency, _ := settings.Currency(ctx, &struct {
		Format currencyFormat
	}{
		Format: acronym,
	})
	input.Currency = models.Currency(currency)
	input.MaxOutDays = settings.MaxOutDays()
	input.MemberRate = settings.memberRateInternal()
	//input.MinBalance = settings.MinBalanceInternal()
	input.MinInDays = settings.MinInDays()
	input.NonMemberRate = settings.nonMemberRateInternal()
	input.PropertyName = testPropertyName
	input.ReservationReminderDaysBefore = settings.ReservationReminderDaysBefore()

	input.ForVersion = property.EventVersion()

	property, err = resolver.UpdateSettings(ctx, &struct {
		PropertyID string
		Input      *models.UpdateSettingsInput
	}{
		PropertyID: property.PropertyID(),
		Input:      input,
	})

	if err != nil {
		t.Fatal(err)
	}

	secondEventVersion := property.EventVersion()

	settings, _ = property.Settings(&settingsArgs{})
	if settings.PropertyName() != testPropertyName {
		t.Fatalf("update property name failed, returned %+v expected %+v", settings.PropertyName(), testPropertyName)
	}

	if firstEventVersion == secondEventVersion {
		t.Fatal("expected the event version to change after the mutation")
	}

	t.Log("test duplicate settings")

	property, err = resolver.UpdateSettings(ctx, &struct {
		PropertyID string
		Input      *models.UpdateSettingsInput
	}{
		PropertyID: property.PropertyID(),
		Input:      input,
	})

	if err != nil {
		t.Fatal(err)
	}

	thirdEventVersion := property.EventVersion()

	if secondEventVersion != thirdEventVersion {
		t.Fatal("expected duplicate supression to return the same event version")
	}

}

func TestSettingsConstraints(t *testing.T) {
	_, ctx, resolver, _, _ := initAndCreateTestProperty(context.Background(), t)

	constraints, err := resolver.UpdateSettingsConstraints(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// test the default for trial
	if utilities.TrialDuration.Hours() != 0 {
		t.Fatalf("default trial duration expected 0 but got %+v", utilities.TrialDuration.Hours())
	}

	if constraints.TrialOn() {
		t.Fatalf("trial should be off by default")
	}

	if constraints.TrialDays() != 0 {
		t.Fatalf("trial days should be zero by default")
	}

	utilities.TrialDuration, err = time.ParseDuration("48h")
	if err != nil {
		t.Fatal(err)
	}

	if !constraints.TrialOn() {
		t.Fatalf("trial should be on")
	}

	if constraints.TrialDays() != 2 {
		t.Fatalf("trial days should be 2")
	}
}
