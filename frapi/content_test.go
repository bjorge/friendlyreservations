package frapi

import (
	"context"
	"testing"

	"github.com/bjorge/friendlyreservations/models"
)

func TestDefaultContent(t *testing.T) {
	property, _, _, _, _ := initAndCreateTestProperty(context.Background(), t)

	// t.Log("create a single reservation")
	// property, _ = createReservation(t, ctx, resolver, property, me.UserId(), today.AddDays(1).ToString(), today.AddDays(3).ToString())
	// property, _ = createReservation(t, ctx, resolver, property, me.UserId(), today.AddDays(4).ToString(), today.AddDays(6).ToString())

	contents, err := property.Contents()
	if err != nil {
		t.Fatal(err)
	}

	if len(contents) != 2 {
		t.Fatalf("expected 2 contents back")
	}

	content := contents[1]

	if content.Template() == "" {
		t.Fatalf("expected some Template")
	}
	t.Logf("Template: %+v", content.Template())

	rendered, err := content.Rendered()
	if err != nil {
		t.Fatal(err)
	}
	if rendered == "" {
		t.Fatalf("expected some Rendered")
	}
	t.Logf("Rendered: %+v", rendered)
	//t.Fatalf("test")

}

func TestCustomContent(t *testing.T) {
	property, ctx, resolver, _, _ := initAndCreateTestProperty(context.Background(), t)

	t.Logf("before update version is %+v", property.EventVersion())

	testTemplate := "template1"
	testComment := "comment1"
	property, err := resolver.CreateContent(ctx, &struct {
		PropertyID string
		Input      *models.NewContentInput
	}{
		PropertyID: property.PropertyID(),
		Input: &models.NewContentInput{
			ForVersion: property.EventVersion(),
			Name:       models.ADMIN_HOME,
			Template:   testTemplate,
			Comment:    testComment,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("after update version is now %+v", property.EventVersion())

	contents, _ := property.Contents()

	for _, content := range contents {
		if content.Name() == models.ADMIN_HOME {
			if content.Comment() != testComment {
				t.Fatalf("comment does not match, expected \n%+v\nbut got \n%+v", testComment, content.Comment())
			}
			if content.Template() != testTemplate {
				t.Fatalf("template does not match, expected \n%+v\nbut got \n%+v", testTemplate, content.Template())
			}
		}
	}

	// now update the template
	testTemplate = "template2"
	testComment = "comment2"
	property, err = resolver.CreateContent(ctx, &struct {
		PropertyID string
		Input      *models.NewContentInput
	}{
		PropertyID: property.PropertyID(),
		Input: &models.NewContentInput{
			ForVersion: property.EventVersion(),
			Name:       models.ADMIN_HOME,
			Template:   testTemplate,
			Comment:    testComment,
		},
	})

	contents, _ = property.Contents()

	for _, content := range contents {
		if content.Name() == models.ADMIN_HOME {
			if content.Template() != testTemplate {
				t.Fatalf("template does not match")
			}
			if content.Comment() != testComment {
				t.Fatalf("comment does not match")
			}
		}
	}

}
