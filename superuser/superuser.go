package superuser

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

func ConfigureSuperuserOverrides(app *pocketbase.PocketBase, vAll *viper.Viper) {
	v := vAll.Sub("superuser")

	if v == nil {
		return
	}

	v.SetDefault("enabled", true)

	if !v.GetBool("enabled") {
		log.Println("Superuser overrides disabled")
		return
	}

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		if err := createSuperusers(app, v); err != nil {
			return fmt.Errorf("error Creating Superusers: %v", err)
		}

		return e.Next()
	})

	overrideCollections(app, v)
}
