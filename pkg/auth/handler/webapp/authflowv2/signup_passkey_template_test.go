package authflowv2

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

// Authany: the combined signup/login page (signup.html, used by the signup_login flow) must offer
// "Continue with passkey" when the flow allows passkey login, as login.html does.
func TestSignupPagePasskeyButton(t *testing.T) {
	Convey("signup.html passkey button", t, func() {
		builtin := afero.NewBasePathFs(afero.NewOsFs(), "../../../../../resources/authgear")
		// __generated_asset.html is produced by the frontend build; an empty one is enough here.
		overlay := afero.NewMemMapFs()
		_ = overlay.MkdirAll("templates/en/web/authflowv2", 0777)
		_ = afero.WriteFile(overlay, "templates/en/web/authflowv2/__generated_asset.html", []byte(`{{ define "authflowv2/__generated_asset.html" }}{{ end }}`), 0666)
		manager := resource.NewManager(resource.DefaultRegistry, []resource.Fs{
			resource.LeveledAferoFs{Fs: afero.NewCopyOnWriteFs(builtin, overlay), FsLevel: resource.FsLevelBuiltin},
		})
		engine := &template.Engine{Resolver: &template.Resolver{
			Resources:             manager,
			DefaultLanguageTag:    "en",
			SupportedLanguageTags: []string{"en"},
		}}
		ctx := context.Background()
		tr, err := engine.Translation(ctx, []string{"en"})
		So(err, ShouldBeNil)

		render := func(passkey bool) string {
			out, err := engine.Render(ctx, TemplateWebAuthflowV2SignupHTML, []string{"en"}, map[string]any{
				"Translations":              tr,
				"StaticAssetURL":            func(string) (string, error) { return "", nil },
				"GeneratedStaticAssetURL":   func(string) (string, error) { return "", nil },
				"HasAppSpecificAsset":       func(string) bool { return false },
				"MakeURL":                   func(string, ...string) string { return "" },
				"MakeBackURL":               func(string, ...string) string { return "" },
				"MakeURLWithBackURL":        func(string, ...string) string { return "" },
				"SliceContains":             func([]string, string) bool { return false },
				"LoginIDKey":                "email",
				"FormsByIdentification":     map[string]any{},
				"LoginIDInputType":          "email",
				"EmailLoginIDEnabled":       true,
				"PasskeyEnabled":            passkey,
				"PasskeyRequestOptionsJSON": "{}",
				"IdentityCandidates": []map[string]any{
					{"type": "login_id", "login_id_key": "email", "login_id_type": "email"},
				},
			})
			So(err, ShouldBeNil)
			return out.String
		}

		So(strings.Contains(render(true), `authgear.button.use_passkey`), ShouldBeTrue)
		So(strings.Contains(render(false), `authgear.button.use_passkey`), ShouldBeFalse)
	})
}
