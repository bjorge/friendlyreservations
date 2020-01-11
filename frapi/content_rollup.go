package frapi

import (
	"github.com/bjorge/friendlyreservations/models"
)

// ContentRollup is the rollup record for (html) content, exported for memcache
type ContentRollup struct {
	// orginal reservation
	Input        *models.NewContentInput
	EventVersion int32
}

// GetEventVersion returns the version of the rollup record
func (r *ContentRollup) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *PropertyResolver) rollupContent() {
	r.rollupMutexes[contentsRollupType].Lock()
	defer r.rollupMutexes[contentsRollupType].Unlock()

	if !r.rollupsExists(contentsRollupType) {

		for _, event := range r.property.Events {
			if newContentInput, ok := event.(*models.NewContentInput); ok {

				contentRollup := &ContentRollup{}
				contentRollup.Input = newContentInput
				contentRollup.EventVersion = newContentInput.EventVersion

				r.addRollup(string(newContentInput.Name),
					contentRollup, contentsRollupType)
			}
		}
		cacheError := r.cacheRollup(contentsRollupType)
		if cacheError != nil {
			Logger.LogWarningf("cache write content rollups error: %+v", cacheError)
		}
	}

}
