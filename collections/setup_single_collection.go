package collections

import (
	"fmt"
	"log"
	"pocketforge/superuser"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type RulesConfig struct {
	ListRule   *string `mapstructure:"listRule" json:"listRule"`
	ViewRule   *string `mapstructure:"viewRule" json:"viewRule"`
	CreateRule *string `mapstructure:"createRule" json:"createRule"`
	DeleteRule *string `mapstructure:"deleteRule" json:"deleteRule"`
	UpdateRule *string `mapstructure:"updateRule" json:"updateRule"`
	AuthRule   *string `mapstructure:"authRule" json:"authRule"`
	ManageRule *string `mapstructure:"manageRule" json:"manageRule"`
}

type CollectionConfig struct {
	ID                       string        `mapstructure:"id" json:"id"`
	Name                     string        `mapstructure:"name" json:"name"`
	Type                     string        `mapstructure:"type" json:"type"`
	Editable                 bool          `mapstructure:"editable" json:"editable"`
	Rules                    RulesConfig   `mapstructure:"rules" json:"rules"`
	AddDefaultFields         bool          `mapstructure:"addDefaultFields" json:"addDefaultFields"`
	RetainUnconfiguredFields bool          `mapstructure:"retainUnconfiguredFields" json:"retainUnconfiguredFields"`
	Fields                   []FieldConfig `mapstructure:"fields" json:"fields"`
	collection               *core.Collection
}

func (configuration *CollectionConfig) CreateOrUpdateCollection(app *pocketbase.PocketBase) {

	if configuration.ID == "" {
		log.Panicf("Collection %s has no ID", configuration.Name)
		return
	}

	if configuration.Name == "" {
		log.Panicf("Collection %s has no Title", configuration.ID)
		return
	}

	if configuration.Type == "" {
		configuration.Type = "base"
	}

	if configuration.Type != "base" && configuration.Type != "view" && configuration.Type != "auth" {
		log.Panicf("Collection %s has invalid type %s", configuration.ID, configuration.Type)
		return
	}

	_, err := app.FindCollectionByNameOrId(configuration.ID)
	if err != nil {
		if configuration.Type == "base" {
			configuration.collection = core.NewBaseCollection(configuration.Name)
		} else if configuration.Type == "view" {
			configuration.collection = core.NewViewCollection(configuration.Name)
		} else if configuration.Type == "auth" {
			configuration.collection = core.NewAuthCollection(configuration.Name)
		}

		// Whether the collection is a system collection is always set to false
		// As setting to true means it cannot be removed or updated which is problematic.
		configuration.collection.System = false
		configuration.collection.Id = configuration.ID

		app.Save(configuration.collection)
	}

	configuration.updateCollectionSettings(app)
	configuration.updateRules(app)
	configuration.lockCollection(app)

}

// getCollection retrieves a collection from the PocketBase application based on the
// CollectionConfig's ID. If the collection is already cached in the configuration,
// it returns the cached collection. Otherwise, it attempts to find the collection
// by its name or ID using the PocketBase instance. If the collection is not found,
// the function logs a panic with the collection ID.
//
// Parameters:
//   - app: A pointer to the PocketBase instance.
//
// Returns:
//   - A pointer to the core.Collection instance.
func (configuration *CollectionConfig) getCollection(app *pocketbase.PocketBase) *core.Collection {
	if configuration.collection != nil {
		return configuration.collection
	}

	collection, err := app.FindCollectionByNameOrId(configuration.ID)
	if err != nil {
		log.Panicf("Collection %s not found", configuration.ID)
	}
	configuration.collection = collection
	return collection
}

// updateCollectionSettings updates the settings of a collection in the PocketBase application.
// It retrieves the collection using the configuration and updates its name if it differs from the configuration's name.
// If the name is updated, the changes are saved back to the PocketBase application.
//
// Parameters:
//   - app: A pointer to the PocketBase application instance.
//
// Note: This function assumes that the CollectionConfig struct has a method getCollection that retrieves the collection from the PocketBase application.
func (configuration *CollectionConfig) updateCollectionSettings(app *pocketbase.PocketBase) {

	collection := configuration.getCollection(app)

	if collection.Name != configuration.Name {
		collection.Name = configuration.Name
		app.Save(collection)
	}

}

// updateRules updates the rules of a collection in the PocketBase application.
// It compares the current rules of the collection with the new rules provided
// in the CollectionConfig and updates the collection if there are any changes.
//
// Parameters:
// - app: A pointer to the PocketBase application instance.
//
// The function checks each rule (ListRule, ViewRule, DeleteRule, CreateRule, UpdateRule)
// and updates the collection's rules if they differ from the new rules. If the collection
// type is "auth", it also checks and updates the AuthRule and ManageRule.
//
// If any rule is updated, the function saves the updated collection in the PocketBase application.
//
// The function panics if the collection in the configuration is nil.
func (configuration *CollectionConfig) updateRules(app *pocketbase.PocketBase) {

	if configuration.collection == nil {
		log.Panicf("Collection %s has no collection", configuration.Name)
	}

	changed := false

	if configuration.Rules.ListRule != configuration.collection.ListRule {
		configuration.collection.ListRule = configuration.Rules.ListRule
		changed = true
	}
	if configuration.Rules.ViewRule != configuration.collection.ViewRule {
		configuration.collection.ViewRule = configuration.Rules.ViewRule
		changed = true
	}
	if configuration.Rules.DeleteRule != configuration.collection.DeleteRule {
		configuration.collection.DeleteRule = configuration.Rules.DeleteRule
		changed = true
	}
	if configuration.Rules.CreateRule != configuration.collection.CreateRule {
		configuration.collection.CreateRule = configuration.Rules.CreateRule
		changed = true
	}
	if configuration.Rules.UpdateRule != configuration.collection.UpdateRule {
		configuration.collection.UpdateRule = configuration.Rules.UpdateRule
		changed = true
	}

	if configuration.Type == "auth" {
		if configuration.Rules.AuthRule != configuration.collection.AuthRule {
			configuration.collection.AuthRule = configuration.Rules.AuthRule
			changed = true
		}
		if configuration.Rules.ManageRule != configuration.collection.ManageRule {
			configuration.collection.ManageRule = configuration.Rules.ManageRule
			changed = true
		}
	}

	if changed {
		app.Save(configuration.collection)
	}
}

func (configuration *CollectionConfig) lockCollection(app *pocketbase.PocketBase) {

	if configuration.Editable {
		return
	}

	override := superuser.CollectionOverrides{
		Name:                    configuration.Name,
		PreventCollectionUpdate: true,
		PreventCollectionCreate: true,
		PreventCollectionDelete: true,
		PreventRecordCreate:     false,
		PreventRecordUpdate:     false,
		PreventRecordDelete:     false,
	}

	override.ProcessCollectionOverride(app)

}

func (configuration *CollectionConfig) UpdateFields(app *pocketbase.PocketBase) {

	collection := configuration.getCollection(app)
	for _, fieldConfig := range configuration.Fields {
		fieldConfig.CreateOrUpdate(app, collection)
	}

	if configuration.AddDefaultFields {
		defaultFields := []FieldConfig{
			{
				Id:       fmt.Sprintf("%s_autodate_created", collection.Name),
				Name:     "created",
				Type:     "autodate",
				OnCreate: true,
			},
			{
				Id:       fmt.Sprintf("%s_autodate_updated", collection.Name),
				Name:     "updated",
				Type:     "autodate",
				OnUpdate: true,
				OnCreate: true,
			},
		}

		for _, fieldConfig := range defaultFields {
			fieldConfig.CreateOrUpdate(app, collection)
		}

	}

	configuration.removeUnusedFields(app)
}

func (configuration *CollectionConfig) removeUnusedFields(app *pocketbase.PocketBase) {

	if configuration.RetainUnconfiguredFields {
		return
	}

	var fields_to_retain []string
	var fieldIds_in_config []string

	for _, fieldConfig := range configuration.Fields {
		fieldIds_in_config = append(fieldIds_in_config, fieldConfig.Id)
		fields_to_retain = append(fields_to_retain, fieldConfig.Name)
	}

	fields := configuration.collection.Fields

	changed := false
	var default_fields []string

	if configuration.AddDefaultFields {
		default_fields = []string{"created", "updated"}
	} else {
		default_fields = []string{}
	}

	for _, field := range fields {
		found := false

		if field.GetSystem() {
			continue
		}

		for _, defaultField := range default_fields {
			if defaultField == field.GetName() {
				found = true
				break
			}
		}

		for _, fieldName := range fields_to_retain {
			if fieldName == field.GetName() {
				found = true
				break
			}
		}

		if found {
			continue
		}

		for _, fieldId := range fieldIds_in_config {
			if fieldId == field.GetId() {
				found = true
				break
			}
		}

		if !found {
			configuration.collection.Fields.RemoveById(field.GetId())
			changed = true
		}
	}

	if changed {
		app.Save(configuration.collection)
	}
}
