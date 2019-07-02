package frapi

import (
	"context"
	"testing"
)

func TestExport(t *testing.T) {

	property, ctx, resolver, me, _ := initAndCreateTestProperty(context.Background(), t)

	msg, err := resolver.exportInternal(ctx, property, me)
	if err != nil {
		t.Fatal(err)
	}

	if len(msg.Attachments) != 1 {
		t.Fatalf("expected an attachment")
	}
	//resolver.Export(ctx, &struct{ PropertyID string }{PropertyID: property.PropertyId()})

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

}
