package collections

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

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
