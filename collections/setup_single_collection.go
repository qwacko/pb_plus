package collections

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type RulesConfig struct {
	ListRule   *string `mapstructure:"listRule" json:"listRule"`
	ViewRule   *string `mapstructure:"viewRule" json:"viewRule"`
	CreateRule *string `mapstructure:"createRule" json:"createRule"`
	DeleteRule *string `mapstructure:"deleteRule" json:"deleteRule"`
	UpdateRule *string `mapstructure:"updateRule" json:"updateRule"`
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

	collection, err := app.FindCollectionByNameOrId(configuration.ID)
	configuration.collection = collection
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

	configuration.updateRules(app)

}

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

	if changed {
		app.Save(configuration.collection)
	}
}

func (configuration *CollectionConfig) UpdateFields(app *pocketbase.PocketBase) {

	collection := configuration.getCollection(app)
	for _, fieldConfig := range configuration.Fields {
		fieldConfig.CreateOrUpdate(app, collection)
	}

	if configuration.AddDefaultFields {
		defaultFields := []FieldConfig{
			{
				Id:       fmt.Sprintf("default_%s_created", collection.Name),
				Name:     "created",
				Type:     "autodate",
				OnCreate: true,
			},
			{
				Id:       fmt.Sprintf("default_%s_updated", collection.Name),
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
}
