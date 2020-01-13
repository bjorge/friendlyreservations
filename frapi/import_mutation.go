package frapi

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"strconv"

	"github.com/bjorge/friendlyreservations/utilities"
)

// ImportProperty is called to import a property from a file
func (r *Resolver) ImportProperty(ctx context.Context) (string, error) {

	if utilities.ImportFileName == "" {
		return "", errors.New("import file name not specified")
	}

	data, err := ioutil.ReadFile(utilities.ImportFileName)

	if err != nil {
		return "", err
	}

	return internalImportProperty(ctx, data)
}

func internalImportProperty(ctx context.Context, data []byte) (string, error) {

	dec := gob.NewDecoder(bytes.NewBuffer(data))
	decoded := &PropertyExport{}
	err := dec.Decode(decoded)
	if err != nil {
		return "", err
	}

	nextPropertyTransactionKey, err := PersistedPropertyList.GetNextVersion(ctx)
	if err != nil {
		return "", err
	}
	propertyID := strconv.Itoa(nextPropertyTransactionKey)

	// don't create the property with too many events...
	numEvents := len(decoded.Events)
	events := decoded.Events
	nextVersion := 0
	maxEvents := 5000
	if numEvents > maxEvents {
		events = decoded.Events[nextVersion:maxEvents]
	}
	nextVersion, err = PersistedVersionedEvents.CreateProperty(ctx, propertyID, events, PersistedPropertyList, nextPropertyTransactionKey)
	if err != nil {
		return "", err
	}

	// update with additional events
	if numEvents > maxEvents {
		for nextVersion < numEvents {
			high := nextVersion + maxEvents
			if high > numEvents {
				high = numEvents
			}
			events = decoded.Events[nextVersion:high]
			nextVersion, err = PersistedVersionedEvents.NewPropertyEvents(ctx, propertyID, nextVersion, events, false)
			if err != nil {
				return "", err
			}
		}
	}

	PersistedEmailStore.RestoreEmails(ctx, propertyID, decoded.EmailMap)

	return propertyID, nil
}
