package collections

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type FieldConfig struct {
	Type                string  `mapstructure:"type" json:"type"`
	Id                  string  `mapstructure:"id" json:"id"`
	Name                string  `mapstructure:"name" json:"name"`
	Required            bool    `mapstructure:"required" json:"required"`
	Hidden              bool    `mapstructure:"hidden" json:"hidden"`
	Min                 int     `mapstructure:"min" json:"min"`
	Max                 int     `mapstructure:"max" json:"max"`
	MinFloat            float64 `mapstructure:"min_float" json:"min_float"`
	MaxFloat            float64 `mapstructure:"max_float" json:"max_float"`
	MaxSize             int64   `mapstructure:"max_size" json:"max_size"`
	Presentable         bool    `mapstructure:"presentable" json:"presentable"`
	Pattern             string  `mapstructure:"pattern" json:"pattern"`
	AutogeneratePattern string  `mapstructure:"autogenerate_pattern" json:"autogenerate_pattern"`
	OnCreate            bool    `mapstructure:"on_create" json:"on_create"`
	OnUpdate            bool    `mapstructure:"on_update" json:"on_update"`
	OnlyInt             bool    `mapstructure:"only_int" json:"only_int"`
	MinSelect           int     `mapstructure:"min_select" json:"min_select"`
	MaxSelect           int     `mapstructure:"max_select" json:"max_select"`

	// File Specific
	MimeTypes []string `mapstructure:"mime_types" json:"mime_types"`
	Thumbs    []string `mapstructure:"thumbs" json:"thumbs"`
	Protected bool     `mapstructure:"protected" json:"protected"`

	// Email and URL Specific
	ExceptDomains []string `mapstructure:"except_domains" json:"except_domains"`
	OnlyDomains   []string `mapstructure:"only_domains" json:"only_domains"`

	// Date Specific
	MinDate string `mapstructure:"min_date" json:"min_date"`
	MaxDate string `mapstructure:"max_date" json:"max_date"`

	// Editor Specific
	ConvertURLs bool `mapstructure:"convert_urls" json:"convert_urls"`

	// Select Specific
	Values []string `mapstructure:"values" json:"values"`

	// Password Specific
	Cost int `mapstructure:"cost" json:"cost"`

	// Relation Specific
	CollectionId  string `mapstructure:"collection_id" json:"collection_id"`
	CascadeDelete bool   `mapstructure:"cascade_delete" json:"cascade_delete"`
}

func (f *FieldConfig) CreateOrUpdate(app *pocketbase.PocketBase, collection *core.Collection) {
	field := f.getExistingField(collection)

	if field != nil && field.Type() != f.Type {
		collection.Fields.RemoveById(f.getId(collection))
		app.Save(collection)
		field = nil
	}

	if field == nil {
		f.createField(app, collection)
	}

	field = f.getExistingField(collection)

	if field == nil {
		log.Panicf("Failed to create field %s", f.Name)
	}

	//Creating the field actually updates it if the id or name already exists.
	f.createField(app, collection)

}

func (f *FieldConfig) createField(app *pocketbase.PocketBase, collection *core.Collection) {
	switch f.Type {
	case "text":
		f.createTextField(app, collection)
	case "json":
		f.createJSONField(app, collection)
	case "autodate":
		f.createAutodateField(app, collection)
	case "file":
		f.createFileField(app, collection)
	case "email":
		f.createEmailField(app, collection)
	case "url":
		f.createURLField(app, collection)
	case "date":
		f.createDateField(app, collection)
	case "editor":
		f.createEditorField(app, collection)
	case "select":
		f.createSelectField(app, collection)
	case "password":
		f.createPasswordField(app, collection)
	case "relation":
		f.createRelationField(app, collection)
	case "number":
		f.createNumberField(app, collection)
	default:
		log.Panicf("Unknown field type %s", f.Type)
	}
}

func (f *FieldConfig) getExistingField(collection *core.Collection) core.Field {
	id := f.getId(collection)
	for _, field := range collection.Fields {
		if field.GetId() == id {
			return field
		}
	}
	return nil
}

func (f *FieldConfig) getId(collection *core.Collection) string {
	if f.Id == "" {
		return fmt.Sprintf("%s_%s_%s", collection.Name, f.Type, f.Name)
	}
	return f.Id
}

func (f *FieldConfig) createTextField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.TextField{
		Id:                  f.getId(collection),
		Name:                f.Name,
		Required:            f.Required,
		Hidden:              f.Hidden,
		Min:                 f.Min,
		Max:                 f.Max,
		Pattern:             f.Pattern,
		Presentable:         f.Presentable,
		AutogeneratePattern: f.AutogeneratePattern,
	})

	app.Save(collection)
}

func (f *FieldConfig) createJSONField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.JSONField{
		Id:          f.getId(collection),
		Name:        f.Name,
		Required:    f.Required,
		Hidden:      f.Hidden,
		MaxSize:     f.MaxSize,
		Presentable: f.Presentable,
	})

	app.Save(collection)
}

func (f *FieldConfig) createAutodateField(app *pocketbase.PocketBase, collection *core.Collection) {

	collection.Fields.Add(&core.AutodateField{
		Id:          f.getId(collection),
		Name:        f.Name,
		OnCreate:    f.OnCreate,
		OnUpdate:    f.OnUpdate,
		Hidden:      f.Hidden,
		Presentable: f.Presentable,
	})

	app.Save(collection)
}

func (f *FieldConfig) createFileField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.FileField{
		Id:          f.getId(collection),
		Name:        f.Name,
		Required:    f.Required,
		Hidden:      f.Hidden,
		MaxSize:     f.MaxSize,
		MaxSelect:   f.MaxSelect,
		Presentable: f.Presentable,
		Protected:   f.Protected,
		MimeTypes:   f.MimeTypes,
		Thumbs:      f.Thumbs,
	})

	app.Save(collection)
}

func (f *FieldConfig) createEmailField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.EmailField{
		Id:            f.getId(collection),
		Name:          f.Name,
		Required:      f.Required,
		Hidden:        f.Hidden,
		Presentable:   f.Presentable,
		ExceptDomains: f.ExceptDomains,
		OnlyDomains:   f.OnlyDomains,
	})

	app.Save(collection)
}

func (f *FieldConfig) createURLField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.URLField{
		Id:            f.getId(collection),
		Name:          f.Name,
		Required:      f.Required,
		Hidden:        f.Hidden,
		Presentable:   f.Presentable,
		ExceptDomains: f.ExceptDomains,
		OnlyDomains:   f.OnlyDomains,
	})

	app.Save(collection)
}

func (f *FieldConfig) createDateField(app *pocketbase.PocketBase, collection *core.Collection) {

	maxDateTime, err := types.ParseDateTime(f.MaxDate)
	if err != nil {
		log.Panicf("Failed to parse MaxDate %s: %v", f.MaxDate, err)
	}

	minDateTime, err := types.ParseDateTime(f.MinDate)
	if err != nil {
		log.Panicf("Failed to parse MinDate %s: %v", f.MinDate, err)
	}

	collection.Fields.Add(&core.DateField{
		Id:          f.getId(collection),
		Name:        f.Name,
		Required:    f.Required,
		Hidden:      f.Hidden,
		Min:         minDateTime,
		Max:         maxDateTime,
		Presentable: f.Presentable,
	})

	app.Save(collection)
}

func (f *FieldConfig) createEditorField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.EditorField{
		Id:          f.getId(collection),
		Name:        f.Name,
		Required:    f.Required,
		Hidden:      f.Hidden,
		Presentable: f.Presentable,
		MaxSize:     f.MaxSize,
		ConvertURLs: f.ConvertURLs,
	})

	app.Save(collection)
}

func (f *FieldConfig) createSelectField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.SelectField{
		Id:          f.getId(collection),
		Name:        f.Name,
		Required:    f.Required,
		Hidden:      f.Hidden,
		Presentable: f.Presentable,
		Values:      f.Values,
		MaxSelect:   f.MaxSelect,
	})

	app.Save(collection)
}

func (f *FieldConfig) createPasswordField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.PasswordField{
		Id:          f.getId(collection),
		Name:        f.Name,
		Required:    f.Required,
		Hidden:      f.Hidden,
		Presentable: f.Presentable,
		Cost:        f.Cost,
		Pattern:     f.Pattern,
		Min:         f.Min,
		Max:         f.Max,
	})

	app.Save(collection)
}

func (f *FieldConfig) createRelationField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.RelationField{
		Id:            f.getId(collection),
		Name:          f.Name,
		Required:      f.Required,
		Hidden:        f.Hidden,
		Presentable:   f.Presentable,
		CollectionId:  f.CollectionId,
		CascadeDelete: f.CascadeDelete,
		MinSelect:     f.MinSelect,
		MaxSelect:     f.MaxSelect,
	})

	app.Save(collection)
}

func (f *FieldConfig) createNumberField(app *pocketbase.PocketBase, collection *core.Collection) {
	collection.Fields.Add(&core.NumberField{
		Id:          f.getId(collection),
		Name:        f.Name,
		Required:    f.Required,
		Hidden:      f.Hidden,
		Presentable: f.Presentable,
		Min:         &f.MinFloat,
		Max:         &f.MaxFloat,
		OnlyInt:     f.OnlyInt,
	})

	app.Save(collection)
}
