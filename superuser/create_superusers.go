package superuser

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

type SuperuserAccount struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func createSuperusers(app *pocketbase.PocketBase, v *viper.Viper) error {

	if !v.IsSet("accounts") {
		return nil
	}

	var accounts []SuperuserAccount
	if err := v.UnmarshalKey("accounts", &accounts); err != nil {
		log.Fatalf("Error unmarshalling superuser accounts: %v", err)
	}

	superusersCollection, err := app.FindCollectionByNameOrId("_superusers")
	if err != nil {
		return fmt.Errorf("Error finding superusers collection: %v", err)
	}

	for _, account := range accounts {

		//Find if the superuser account already exists
		existingAccount, err := app.FindFirstRecordByData(superusersCollection, "email", account.Email)

		if err != nil {
			if err == sql.ErrNoRows {

				record := core.NewRecord(superusersCollection)
				record.Set("email", account.Email)
				record.Set("password", account.Password)
				record.Set("passwordConfirm", account.Password)

				if err := app.Save(record); err != nil {
					return fmt.Errorf("Error creating superuser account: %v", err)
				}
				continue
			} else {
				return fmt.Errorf("Error finding superuser account: %v", err)
			}
		}

		existingAccount.Set("password", account.Password)
		existingAccount.Set("passwordConfirm", account.Password)

		if err := app.Save(existingAccount); err != nil {
			return fmt.Errorf("Error updating superuser account: %v", err)
		}

	}

	return nil
}
