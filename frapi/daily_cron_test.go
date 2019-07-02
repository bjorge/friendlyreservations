package frapi

import (
	"context"
	"testing"
	"time"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/templates"
	"github.com/bjorge/friendlyreservations/utilities"
)

func TestDailyCronNotifications(t *testing.T) {
	_, ctx, resolver, me, _ := initAndCreateTestProperty(context.Background(), t)
	utilities.DebugLog(ctx, "testing")

	// now run the cron job - nothing should happen because balance not negative
	if err := DailyCron(ctx); err != nil {
		t.Fatal(err)
	}

	property := getUpdatedProperty(ctx, t, resolver)

	settings, _ := property.Settings(&settingsArgs{})
	dateBuilder := frdate.MustNewDateBuilder(settings.Timezone())

	lastNotifications := property.lastNotifications(dateBuilder)

	t.Log("test for base case of no notifications")
	logToday(t, property)
	if _, ok := lastNotifications[templates.LowBalanceNotification]; ok {
		t.Fatalf("there should be no balance notifications yet")
	}

	// ok, now lower the balance and see if we get the balance notification
	property = createPayment(ctx, t, resolver, property, 10000, false, property.EventVersion())

	if err := DailyCron(ctx); err != nil {
		t.Fatal(err)
	}
	property = getUpdatedProperty(ctx, t, resolver)

	// make sure the body resolves without issues
	notifications, _ := property.Notifications(&notificationArgs{})
	for _, notification := range notifications {
		_, err := notification.Body()
		if err != nil {
			t.Fatal(err)
		}
		//t.Logf("body '%+v'", body)
	}

	t.Log("test for an immediate notification")
	logToday(t, property)
	lastNotifications = property.lastNotifications(dateBuilder)
	timeDate1, ok := lastNotifications[templates.LowBalanceNotification][me.UserID()]
	if !ok {
		t.Fatalf("there should be a balance notification now")
	}

	// cool, now let's fast forward one day and see make sure we don't get another notification
	t.Log("test for no notification inside wait interval")
	timeOffset := 1
	frdate.TestTimeOffsetDays = &timeOffset
	logToday(t, property)

	if err := DailyCron(ctx); err != nil {
		t.Fatal(err)
	}
	property = getUpdatedProperty(ctx, t, resolver)

	lastNotifications = property.lastNotifications(dateBuilder)
	timeDate2, _ := lastNotifications[templates.LowBalanceNotification][me.UserID()]
	if !timeDate1.Equal(timeDate2) {
		t.Fatalf("expected no new balance notification")
	}

	t.Log("test for new notification after wait interval")
	// maxFrequencyDuration := fmt.Sprintf("%dh", settings.BalanceReminderIntervalDays()*24+1)
	timeOffset = int(settings.BalanceReminderIntervalDays())
	frdate.TestTimeOffsetDays = &timeOffset
	logToday(t, property)

	if err := DailyCron(ctx); err != nil {
		t.Fatal(err)
	}
	property = getUpdatedProperty(ctx, t, resolver)

	lastNotifications = property.lastNotifications(dateBuilder)
	timeDate3, _ := lastNotifications[templates.LowBalanceNotification][me.UserID()]
	if timeDate2.Equal(timeDate3) {
		t.Logf("timeDate1 %+v timeDate2 %+v timeDate3 %+v", timeDate1.ToString(), timeDate2.ToString(), timeDate3.ToString())
		t.Fatalf("expected a new balance notification")
	}
}

func TestDailyCronTrialExpiration(t *testing.T) {
	_, ctx, resolver, _, _ := initAndCreateTestProperty(context.Background(), t)
	utilities.DebugLog(ctx, "testing trial expiration")

	// now run the cron job - nothing should happen because trial not turned on
	t.Logf("run cron with defaults")
	if err := DailyCron(ctx); err != nil {
		t.Fatal(err)
	}

	property := getUpdatedProperty(ctx, t, resolver)

	// ok, now set the trial period for 2 days and allow deletion flag...
	utilities.AllowDeleteProperty = true
	utilities.TrialDuration, _ = time.ParseDuration("48h")

	// go forward one day, should not delete
	t.Logf("move time 1 day forward")
	timeOffset := 1
	frdate.TestTimeOffsetDays = &timeOffset
	logToday(t, property)
	if err := DailyCron(ctx); err != nil {
		t.Fatal(err)
	}
	property = getUpdatedProperty(ctx, t, resolver)

	// go forward three days, should delete
	t.Logf("move time 3 days forward")
	timeOffset = 3
	frdate.TestTimeOffsetDays = &timeOffset
	logToday(t, property)
	if err := DailyCron(ctx); err != nil {
		t.Fatal(err)
	}

	properties, err := resolver.Properties(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(properties) > 0 {
		t.Fatalf("property should be deleted")
	}

}

func getUpdatedProperty(ctx context.Context, t *testing.T, resolver *Resolver) *PropertyResolver {
	properties, err := resolver.Properties(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return properties[0]
}

func logToday(t *testing.T, property *PropertyResolver) {
	settings, err := property.Settings(&settingsArgs{})
	if err != nil {
		t.Fatal(err)
	}
	dateBuilder := frdate.MustNewDateBuilder(settings.Timezone())
	today := dateBuilder.MustNewDateTime(frdate.CreateDateTimeUTC())
	t.Logf("today is: %+v", today.ToString())
}
