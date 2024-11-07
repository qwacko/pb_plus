package collections

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/viper"
)

type EmailTemplateConfig struct {
	Subject string `mapstructure:"subject" json:"subject"`
	Body    string `mapstructure:"body" json:"body"`
}

type AuthAlertConfig struct {
	Enabled       bool                `mapstructure:"enabled" json:"enabled"`
	EmailTemplate EmailTemplateConfig `mapstructure:"email_template" json:"email_template"`
}

type TokenConfig struct {
	Duration int `mapstructure:"duration" json:"duration"`
}

type MFAConfig struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled"`
	Duration int    `mapstructure:"duration" json:"duration"`
	Rule     string `mapstructure:"rule" json:"rule"`
}

type OTPConfig struct {
	Enabled       bool                `mapstructure:"enabled" json:"enabled"`
	Duration      int                 `mapstructure:"duration" json:"duration"`
	Length        int                 `mapstructure:"length" json:"length"`
	EmailTemplate EmailTemplateConfig `mapstructure:"email_template" json:"email_template"`
}

type PasswordConfig struct {
	Enabled        bool     `mapstructure:"enabled" json:"enabled"`
	IdentityFields []string `mapstructure:"identity_fields" json:"identity_fields"`
}

type OAuth2ProviderConfig struct {
	PKCE *bool `mapstructure:"pkce,omitempty" json:"pkce"`

	Name         string         `mapstructure:"name" json:"name"`
	ClientId     string         `mapstructure:"client_id" json:"client_id"`
	ClientSecret string         `mapstructure:"client_secret,omitempty" json:"client_secret,omitempty"`
	AuthURL      string         `mapstructure:"auth_url,omitempty" json:"auth_url"`
	TokenURL     string         `mapstructure:"token_url,omitempty" json:"token_url"`
	UserInfoURL  string         `mapstructure:"user_info_url,omitempty" json:"user_info_url"`
	DisplayName  string         `mapstructure:"display_name,omitempty" json:"display_name"`
	Extra        map[string]any `mapstructure:"extra,omitempty" json:"extra"`
}

type OAuth2MappedFieldConfig struct {
	AvatarURL string `mapstructure:"avatar_url" json:"avatar_url"`
	Id        string `mapstructure:"id" json:"id"`
	Name      string `mapstructure:"name" json:"name"`
	Username  string `mapstructure:"username" json:"username"`
}

type OAuth2Config struct {
	Enabled      bool                    `mapstructure:"enabled" json:"enabled"`
	MappedFields OAuth2MappedFieldConfig `mapstructure:"mapped_fields" json:"mapped_fields"`
	Providers    []OAuth2ProviderConfig  `mapstructure:"providers" json:"providers"`
}

type AuthConfig struct {
	AuthAlert                  AuthAlertConfig     `mapstructure:"auth_alert" json:"auth_alert"`
	AuthToken                  TokenConfig         `mapstructure:"auth_token" json:"auth_token"`
	ConfirmEmailChangeTemplate EmailTemplateConfig `mapstructure:"confirm_email_change_template" json:"confirm_email_change_template"`
	EmailChangeToken           TokenConfig         `mapstructure:"email_change_token" json:"email_change_token"`
	FileToken                  TokenConfig         `mapstructure:"file_token" json:"file_token"`
	MFA                        MFAConfig           `mapstructure:"mfa" json:"mfa"`
	OTP                        OTPConfig           `mapstructure:"otp" json:"otp"`
	PasswordAuth               PasswordConfig      `mapstructure:"password_auth" json:"password_auth"`
	PasswordResetToken         TokenConfig         `mapstructure:"password_reset_token" json:"password_reset_token"`
	ResetPasswordTemplate      EmailTemplateConfig `mapstructure:"reset_pasword_template" json:"reset_pasword_template"`
	VerificationTemplate       EmailTemplateConfig `mapstructure:"verification_template" json:"verification_template"`
	VerificationToken          TokenConfig         `mapstructure:"verification_token" json:"verification_token"`
	OAuth2                     OAuth2Config        `mapstructure:"oauth" json:"oauth"`
}

type AuthConfigAction struct {
	AuthConfig  AuthConfig
	TableConfig *CollectionConfig
	V           *viper.Viper
	App         *pocketbase.PocketBase
}

func (configuration *CollectionConfig) ConfigAuth(app *pocketbase.PocketBase, v *viper.Viper) {

	_, err := configuration.refreshCollection(app)
	if err != nil {
		log.Panicf("Auth Collection %s not found", configuration.Name)
	}

	if configuration.collection == nil {
		log.Panicf("Auth Collection %s not found", configuration.Name)
	}

	if configuration.collection.Type != "auth" {
		log.Panicf("Collection %s is not an auth collection", configuration.Name)
	}

	// Auth Alert
	processAuthStringItem(v, "auth_alert.email_template.subject", &configuration.collection.AuthAlert.EmailTemplate.Subject)
	processAuthStringItem(v, "auth_alert.email_template.body", &configuration.collection.AuthAlert.EmailTemplate.Body)
	processAuthBoolItem(v, "auth_alert.enabled", &configuration.collection.AuthAlert.Enabled)

	// Auth Token
	processAuthInt64Item(v, "auth_token.duration", &configuration.collection.AuthToken.Duration)

	// Confirm Email Change Template
	processAuthStringItem(v, "confirm_email_change_template.subject", &configuration.collection.ConfirmEmailChangeTemplate.Subject)
	processAuthStringItem(v, "confirm_email_change_template.body", &configuration.collection.ConfirmEmailChangeTemplate.Body)

	// Email Change Token
	processAuthInt64Item(v, "email_change_token.duration", &configuration.collection.EmailChangeToken.Duration)

	// File Token
	processAuthInt64Item(v, "file_token.duration", &configuration.collection.FileToken.Duration)

	// MFA
	processAuthBoolItem(v, "mfa.enabled", &configuration.collection.MFA.Enabled)
	processAuthInt64Item(v, "mfa.duration", &configuration.collection.MFA.Duration)
	processAuthStringItem(v, "mfa.rule", &configuration.collection.MFA.Rule)

	// OTP
	processAuthBoolItem(v, "otp.enabled", &configuration.collection.OTP.Enabled)
	processAuthInt64Item(v, "otp.duration", &configuration.collection.OTP.Duration)
	processAuthIntItem(v, "otp.length", &configuration.collection.OTP.Length)
	processAuthStringItem(v, "otp.email_template.subject", &configuration.collection.OTP.EmailTemplate.Subject)
	processAuthStringItem(v, "otp.email_template.body", &configuration.collection.OTP.EmailTemplate.Body)

	// Password Auth
	processAuthBoolItem(v, "password_auth.enabled", &configuration.collection.PasswordAuth.Enabled)
	processAuthStringSliceItem(v, "password_auth.identity_fields", &configuration.collection.PasswordAuth.IdentityFields)

	// Password Reset Token
	processAuthInt64Item(v, "password_reset_token.duration", &configuration.collection.PasswordResetToken.Duration)

	// Reset Pasword Template
	processAuthStringItem(v, "reset_pasword_template.subject", &configuration.collection.ResetPasswordTemplate.Subject)
	processAuthStringItem(v, "reset_pasword_template.body", &configuration.collection.ResetPasswordTemplate.Body)

	// Verification Template
	processAuthStringItem(v, "verification_template.subject", &configuration.collection.VerificationTemplate.Subject)
	processAuthStringItem(v, "verification_template.body", &configuration.collection.VerificationTemplate.Body)

	// Verification Token
	processAuthInt64Item(v, "verification_token.duration", &configuration.collection.VerificationToken.Duration)

	// Oauth
	processAuthBoolItem(v, "oauth.enabled", &configuration.collection.OAuth2.Enabled)
	processAuthStringItem(v, "oauth.mapped_fields.avatar_url", &configuration.collection.OAuth2.MappedFields.AvatarURL)
	processAuthStringItem(v, "oauth.mapped_fields.id", &configuration.collection.OAuth2.MappedFields.Id)
	processAuthStringItem(v, "oauth.mapped_fields.name", &configuration.collection.OAuth2.MappedFields.Name)
	processAuthStringItem(v, "oauth.mapped_fields.username", &configuration.collection.OAuth2.MappedFields.Username)

	// Oauth Providers
	if v.IsSet("oauth.providers") {
		var providers []core.OAuth2ProviderConfig
		err := v.UnmarshalKey("oauth.providers", &providers)
		if err != nil {
			log.Panicf("Failed to unmarshal oauth providers: %v", err)
		}
		configuration.collection.OAuth2.Providers = providers
	}

	configuration.saveAndRefreshCollection(app)
	if err != nil {
		log.Panicf("Failed to save collection: %v", err)
	}

}

func processAuthStringItem(v *viper.Viper, key string, value *string) {
	if v.IsSet(key) {
		*value = v.GetString(key)
	}
}

func processAuthBoolItem(v *viper.Viper, key string, value *bool) {

	if v.IsSet(key) {
		*value = v.GetBool(key)
	}

}

func processAuthInt64Item(v *viper.Viper, key string, value *int64) {
	if v.IsSet(key) {
		*value = v.GetInt64(key)
	}
}

func processAuthIntItem(v *viper.Viper, key string, value *int) {
	if v.IsSet(key) {
		*value = v.GetInt(key)
	}
}

func processAuthStringSliceItem(v *viper.Viper, key string, value *[]string) {
	if v.IsSet(key) {
		*value = v.GetStringSlice(key)
	}
}
