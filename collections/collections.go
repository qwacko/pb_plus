package collections

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

// TODO : Make View Collections Work
// TODO : Make Auth Collections Work
// TODO : Create JSON Schema for collections to allow more complex validation.

func SetupConfiguredCollections(app *pocketbase.PocketBase, vAll *viper.Viper) {

	v := vAll.Sub("collections")

	if v == nil {
		return
	}

	v.SetDefault("enabled", true)

	if !v.GetBool("enabled") {
		return
	}

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {

		SetupCollections(app, v)

		return e.Next()
	})

}
