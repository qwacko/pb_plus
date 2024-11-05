package collections

import (
	"log"

	"github.com/pocketbase/pocketbase"
)

type RulesConfig struct {
	ListRule   *string `mapstructure:"listRule" json:"listRule"`
	ViewRule   *string `mapstructure:"viewRule" json:"viewRule"`
	CreateRule *string `mapstructure:"createRule" json:"createRule"`
	DeleteRule *string `mapstructure:"deleteRule" json:"deleteRule"`
	UpdateRule *string `mapstructure:"updateRule" json:"updateRule"`
}

type CollectionConfig struct {
	ID               string      `mapstructure:"id" json:"id"`
	Title            string      `mapstructure:"title" json:"title"`
	Rules            RulesConfig `mapstructure:"rules" json:"rules"`
	AddDefaultFields bool        `mapstructure:"addDefaultFields" json:"addDefaultFields"`
}

func SetupSingleCollection(app *pocketbase.PocketBase, configuration CollectionConfig) {

	log.Panicln("SetupSingleCollection not implemented")
}
