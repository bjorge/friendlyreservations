package frapi

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestReservationNotification(t *testing.T) {
	property, ctx, resolver, me, today := initAndCreateTestProperty(context.Background(), t)

	property, _ = createReservation(ctx, t, resolver, property, me.UserID(), today.AddDays(1).ToString(), today.AddDays(3).ToString())

	notifications, err := property.Notifications(&notificationArgs{})

	if err != nil {
		t.Fatal(err)
	}

	if len(notifications) != 2 {
		t.Fatalf("expected 2 notifications")
	}

	t.Log("got 2 notifications")

	reservationNotification := notifications[1]

	t.Logf("reservation notification: %+v", reservationNotification)
	body, err := reservationNotification.Body()
	t.Logf("body: %+v", body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(body, today.AddDays(1).ToString()) {
		t.Fatalf("wrong body")
	}
}

func TestBalanceUpdateNotifications(t *testing.T) {
	t.Logf("TestBalanceUpdateNotifications: create property")
	property, ctx, resolver, _, _ := initAndCreateTestProperty(context.Background(), t)

	t.Logf("TestBalanceUpdateNotifications: first payment")
	property = createPayment(ctx, t, resolver, property, int32(1111), true, property.EventVersion())

	notifications, _ := property.Notifications(&notificationArgs{})

	for i, notification := range notifications {
		body, _ := notification.Body()
		item := fmt.Sprintf("i: %+v body: %+v", i, body)
		t.Logf(item)
	}

	if len(notifications) != 2 {
		t.Fatalf("expected 2 notifications, one for create property and one for payment")
	}

	body, _ := notifications[1].Body()
	if !strings.Contains(body, "11.11") {
		t.Fatalf("body not correct in first payment notification")
	}

	t.Logf("TestBalanceUpdateNotifications: second payment")
	property = createPayment(ctx, t, resolver, property, int32(2222), true, property.EventVersion())

	notifications, _ = property.Notifications(&notificationArgs{})

	for i, notification := range notifications {
		body, _ := notification.Body()
		item := fmt.Sprintf("i: %+v body: %+v", i, body)
		t.Logf(item)
	}

	if len(notifications) != 3 {
		t.Fatalf("expected 3 notifications, one for create property and two for payments")
	}

	body, _ = notifications[2].Body()
	if !strings.Contains(body, "22.22") {
		t.Fatalf("body not correct in second payment notification")
	}

	body, _ = notifications[1].Body()
	if !strings.Contains(body, "11.11") {
		t.Fatalf("body not correct in first payment notification")
	}

}

func TestNotifications(t *testing.T) {

	property, ctx, resolver, me, _ := initAndCreateTestProperty(context.Background(), t)

	notifications, err := property.Notifications(&notificationArgs{})

	if err != nil {
		t.Fatal(err)
	}

	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification")
	}
	t.Logf("notification: %+v", *notifications[0].rollup.Input)

	author := notifications[0].Author().Nickname()
	t.Logf("author: %+v", author)

	subject, err := notifications[0].Subject()
	t.Logf("subject: %+v", subject)
	if err != nil {
		t.Fatal(err)
	}
	if subject != "Test Property: New property created" {
		t.Fatalf("wrong subject")
	}

	body, err := notifications[0].Body()
	t.Logf("body: %+v", body)
	if err != nil {
		t.Fatal(err)
	}
	if body != "New property created" {
		t.Fatalf("wrong body '%+v'", body)
	}

	if notifications[0].Read() {
		t.Fatalf("should not be read yet")
	}

	property, err = resolver.NotificationRead(ctx, &struct {
		PropertyID     string
		NotificationID string
		ForVersion     int32
	}{
		PropertyID:     property.PropertyID(),
		NotificationID: notifications[0].NotificationID(),
		ForVersion:     property.EventVersion(),
	})
	if err != nil {
		t.Fatal(err)
	}

	userID := me.UserID()
	notifications, err = property.Notifications(&notificationArgs{UserID: &userID})

	if !notifications[0].Read() {
		t.Fatalf("should be read now")
	}

}
