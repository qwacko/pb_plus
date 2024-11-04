package validation

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func validateSchemaTableColumns(app *pocketbase.PocketBase, collection *core.Collection, viewRule *string) (*core.Collection, bool) {

	changed := false

	createOrUpdateCollectionRules(collection, RulesConfig{
		ListRule:   viewRule,
		ViewRule:   viewRule,
		CreateRule: nil,
		DeleteRule: nil,
		UpdateRule: nil,
	}, &changed)

	createOrUpdateTextField(collection, "table", &core.TextField{
		Name:        "table",
		Required:    true,
		Hidden:      false,
		Min:         0,
		Max:         0,
		Presentable: true,
	}, &changed)

	createOrUpdateTextField(collection, "column", &core.TextField{
		Name:        "column",
		Required:    true,
		Hidden:      false,
		Min:         0,
		Max:         0,
		Presentable: true,
	}, &changed)

	createOrUpdateTextField(collection, "hash", &core.TextField{
		Name:        "hash",
		Required:    true,
		Hidden:      false,
		Min:         0,
		Max:         0,
		Presentable: false,
	}, &changed)

	createOrUpdateJSONField(collection, "schema", &core.JSONField{
		Name:        "schema",
		Required:    true,
		Hidden:      false,
		MaxSize:     0,
		Presentable: false,
	}, &changed)

	createOrUpdateAutodateField(collection, "updated", &core.AutodateField{
		Name:        "updated",
		OnCreate:    true,
		OnUpdate:    true,
		Hidden:      false,
		Presentable: false,
	}, &changed)

	createOrUpdateAutodateField(collection, "created", &core.AutodateField{
		Name:        "created",
		OnCreate:    true,
		OnUpdate:    false,
		Hidden:      false,
		Presentable: false},
		&changed)

	return collection, changed

}

func getOrCreateSchemaCollection(app *pocketbase.PocketBase, schemaTable string, viewRule *string) *core.Collection {
	collection, err := app.FindCollectionByNameOrId(schemaTable)
	if err != nil {
		collection = core.NewBaseCollection(schemaTable)
		app.Save(collection)
	}

	collection, changed := validateSchemaTableColumns(app, collection, viewRule)
	if changed {
		app.Save(collection)
	}
	return collection
}
