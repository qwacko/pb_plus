package collections

import (
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/viper"
)

type CollectionPluginConfig struct {
	Enabled                       bool               `mapstructure:"enabled" json:"enabled"`
	RetainUnconfiguredCollections bool               `mapstructure:"retainUnconfiguredCollections" json:"retainUnconfiguredCollections"`
	FilterPrefix                  string             `mapstructure:"filterPrefix" json:"filterPrefix"`
	Collections                   []CollectionConfig `mapstructure:"collections" json:"collections"`
}

func SetupCollections(app *pocketbase.PocketBase, v *viper.Viper) {

	if v == nil {
		return
	}

	v.SetDefault("enabled", true)
	v.SetDefault("retainUnconfiguredCollections", false)
	v.SetDefault("retainUnconfiguredFields", false)
	v.SetDefault("filterPrefix", "_")

	if !v.GetBool("enabled") {
		return
	}

	pluginConfig := CollectionPluginConfig{}
	err := v.Unmarshal(&pluginConfig)
	if err != nil {
		panic(err)
	}

	log.Println("Collections Plugin Configuration : ", pluginConfig)

	// First create collections and then create fields to ensure the collections exist prior to creating any reference fields.
	for _, collectionConfig := range pluginConfig.Collections {
		collectionConfig.CreateOrUpdateCollection(app)
	}

	for _, collectionConfig := range pluginConfig.Collections {
		collectionConfig.UpdateFields(app)
	}

	pluginConfig.removeUnusedCollections(app)

}

func (config *CollectionPluginConfig) removeUnusedCollections(app *pocketbase.PocketBase) {

	if config.RetainUnconfiguredCollections {
		return
	}

	collections_to_retain := []string{"users"}

	var collections_in_config []string
	var collectionIds_in_config []string
	for _, collectionConfig := range config.Collections {
		collections_in_config = append(collections_in_config, collectionConfig.Name)
		collectionIds_in_config = append(collectionIds_in_config, collectionConfig.ID)
	}

	collections, err := app.FindAllCollections()
	if err != nil {
		log.Panicf("Failed to find collections: %v", err)
	}

	for _, collection := range collections {
		found := false

		// Do not remove collections in the list of collections in the config.
		for _, collectionName := range collections_in_config {
			if collectionName == collection.Name {
				found = true
				break
			}
		}

		// Do not remove collections in the list of items to retain.
		for _, collectionName := range collections_to_retain {
			if collectionName == collection.Name {
				found = true
				break
			}
		}

		// Do not remove collections in the list of items to retain.
		for _, collectionId := range collectionIds_in_config {
			if collectionId == collection.Id {
				found = true
				break
			}
		}

		// Do not remove collections with the filter prefix
		if strings.HasPrefix(collection.Name, config.FilterPrefix) {
			found = true
		}

		// Do not remove system collections
		if collection.System {
			found = true
		}

		if !found {
			collection, err := app.FindCollectionByNameOrId(collection.Name)
			if err != nil {
				log.Panicf("Failed to find collection %s: %v", collection.Name, err)
			}
			app.Delete(collection)
			log.Printf("Removed unused collection %s", collection.Name)
		}
	}
}
