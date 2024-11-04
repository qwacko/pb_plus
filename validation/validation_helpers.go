package validation

import (
	"github.com/pocketbase/pocketbase/core"
)

type RulesConfig struct {
	ListRule   *string
	ViewRule   *string
	CreateRule *string
	DeleteRule *string
	UpdateRule *string
}

func createOrUpdateCollectionRules(collection *core.Collection, rules RulesConfig, changed *bool) {
	if rules.ListRule != collection.ListRule {
		collection.ListRule = rules.ListRule
		*changed = true
	}
	if rules.ViewRule != collection.ViewRule {
		collection.ViewRule = rules.ViewRule
		*changed = true
	}
	if rules.DeleteRule != collection.DeleteRule {
		collection.DeleteRule = rules.DeleteRule
		*changed = true
	}
	if rules.CreateRule != collection.CreateRule {
		collection.CreateRule = rules.CreateRule
		*changed = true
	}
	if rules.UpdateRule != collection.UpdateRule {
		collection.UpdateRule = rules.UpdateRule
		*changed = true
	}
}

func createOrUpdateTextField(collection *core.Collection, fieldName string, configuration *core.TextField, changed *bool) {
	field := collection.Fields.GetByName(fieldName)
	if field == nil {
		*changed = true
		collection.Fields.Add(configuration)
	} else {
		textField, ok := field.(*core.TextField)
		if !ok {
			*changed = true
			collection.Fields.RemoveByName(fieldName)
			collection.Fields.Add(configuration)
		} else {
			if textField.Hidden != configuration.Hidden {
				textField.Hidden = configuration.Hidden
				*changed = true
			}
			if textField.Required != configuration.Required {
				textField.Required = configuration.Required
				*changed = true
			}

			if textField.Min != configuration.Min {
				textField.Min = configuration.Min
				*changed = true
			}

			if textField.Max != configuration.Max {
				textField.Max = configuration.Max
				*changed = true
			}

			if textField.Presentable != configuration.Presentable {
				textField.Presentable = configuration.Presentable
				*changed = true
			}
		}
	}
}

func createOrUpdateJSONField(collection *core.Collection, fieldName string, configuration *core.JSONField, changed *bool) {

	field := collection.Fields.GetByName(fieldName)
	if field == nil {
		*changed = true
		collection.Fields.Add(configuration)
	} else {
		jsonField, ok := field.(*core.JSONField)
		if !ok {
			*changed = true
			collection.Fields.RemoveByName(fieldName)
			collection.Fields.Add(configuration)
		} else {
			if jsonField.Hidden != configuration.Hidden {
				jsonField.Hidden = configuration.Hidden
				*changed = true
			}
			if jsonField.Required != configuration.Required {
				jsonField.Required = configuration.Required
				*changed = true
			}
			if jsonField.Presentable != configuration.Presentable {
				jsonField.Presentable = configuration.Presentable
				*changed = true
			}
			if jsonField.MaxSize != configuration.MaxSize {
				jsonField.MaxSize = configuration.MaxSize
				*changed = true
			}
		}
	}
}

func createOrUpdateAutodateField(collection *core.Collection, fieldName string, configuration *core.AutodateField, changed *bool) {

	field := collection.Fields.GetByName(fieldName)
	if field == nil {
		*changed = true
		collection.Fields.Add(configuration)
	} else {
		autodateField, ok := field.(*core.AutodateField)
		if !ok {
			*changed = true
			collection.Fields.RemoveByName(fieldName)
			collection.Fields.Add(configuration)
		} else {
			if autodateField.Hidden != configuration.Hidden {
				autodateField.Hidden = configuration.Hidden
				*changed = true
			}
			if autodateField.Presentable != configuration.Presentable {
				autodateField.Presentable = configuration.Presentable
				*changed = true
			}
			if autodateField.OnCreate != configuration.OnCreate {
				autodateField.OnCreate = configuration.OnCreate
				*changed = true
			}
			if autodateField.OnUpdate != configuration.OnUpdate {
				autodateField.OnUpdate = configuration.OnUpdate
				*changed = true
			}
		}
	}
}
