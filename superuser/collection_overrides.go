package superuser

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

type CollectionOverrides struct {
	Name                    string `mapstructure:"name"`
	PreventCollectionUpdate bool   `mapstructure:"prevent_collection_update"`
	PreventCollectionCreate bool   `mapstructure:"prevent_collection_create"`
	PreventCollectionDelete bool   `mapstructure:"prevent_collection_delete"`
	PreventRecordCreate     bool   `mapstructure:"prevent_record_create"`
	PreventRecordUpdate     bool   `mapstructure:"prevent_record_update"`
	PreventRecordDelete     bool   `mapstructure:"prevent_record_delete"`
}

func overrideCollections(app *pocketbase.PocketBase, v *viper.Viper) error {

	if !v.IsSet("collections") {
		return nil
	}

	var overrides []CollectionOverrides
	if err := v.UnmarshalKey("collections", &overrides); err != nil {
		log.Fatalf("Error unmarshalling collection overrides: %v", err)
	}

	for _, override := range overrides {
		if err := override.ProcessCollectionOverride(app); err != nil {
			return fmt.Errorf("error processing collection override: %v", err)
		}
	}

	return nil
}

func (override CollectionOverrides) ProcessCollectionOverride(app *pocketbase.PocketBase) error {

	if override.PreventCollectionUpdate {
		app.OnCollectionUpdateRequest().BindFunc(func(e *core.CollectionRequestEvent) error {
			if e.Collection.Name == override.Name {
				if e.HasSuperuserAuth() {
					return apis.NewForbiddenError(fmt.Sprintf("Collection %s cannot be updated", override.Name), nil)
				}
				return e.Next()
			}
			return e.Next()
		})
	}

	if override.PreventCollectionCreate {
		app.OnCollectionCreateRequest().BindFunc(func(e *core.CollectionRequestEvent) error {
			if e.Collection.Name == override.Name && e.HasSuperuserAuth() {
				return apis.NewForbiddenError(fmt.Sprintf("Collection named %s cannot be created.", override.Name), nil)
			}
			return e.Next()
		})
	}

	if override.PreventCollectionDelete {
		app.OnCollectionDeleteRequest().BindFunc(func(e *core.CollectionRequestEvent) error {
			if e.Collection.Name == override.Name && e.HasSuperuserAuth() {
				return apis.NewForbiddenError(fmt.Sprintf("Collection %s cannot be deleted.", override.Name), nil)
			}
			return e.Next()
		})
	}

	if override.PreventRecordCreate {
		app.OnRecordCreateRequest().BindFunc(func(e *core.RecordRequestEvent) error {
			if e.Collection.Name == override.Name && e.HasSuperuserAuth() {
				return apis.NewForbiddenError(fmt.Sprintf("Collection %s is doesn't allow superusers to create records.", override.Name), nil)
			}
			return e.Next()
		})
	}

	if override.PreventRecordUpdate {
		app.OnRecordUpdateRequest().BindFunc(func(e *core.RecordRequestEvent) error {
			if e.Collection.Name == override.Name && e.HasSuperuserAuth() {
				return apis.NewForbiddenError(fmt.Sprintf("Collection %s is doesn't allow superusers to update records.", override.Name), nil)
			}
			return e.Next()
		})
	}

	if override.PreventRecordDelete {
		app.OnRecordDeleteRequest().BindFunc(func(e *core.RecordRequestEvent) error {
			if e.Collection.Name == override.Name && e.HasSuperuserAuth() {
				return apis.NewForbiddenError(fmt.Sprintf("Collection %s is doesn't allow superusers to delete records.", override.Name), nil)
			}
			return e.Next()
		})
	}

	return nil
}
