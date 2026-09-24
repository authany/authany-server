package service

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	runtimeresource "github.com/authgear/authgear-server"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

// Authany: the admin console translations (the en de-brand overrides and the
// zh-CN locale) ship in the image as resources/portal/translations.json.
// go build does not parse embedded files, but a broken file makes
// /api/system-config.json fail, and with it the whole console.
func TestAuthanyBuiltinTranslations(t *testing.T) {
	Convey("built-in translations.json", t, func() {
		resourceManager := resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
			Registry:              portalresource.PortalRegistry,
			BuiltinResourceFS:     runtimeresource.EmbedFS_resources_portal,
			BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_portal,
		})

		result, err := resourceManager.Read(context.Background(), portalresource.TranslationsJSON, resource.EffectiveResource{})
		So(err, ShouldBeNil)
		data, ok := result.([]byte)
		So(ok, ShouldBeTrue)

		// Every locale is a flat map of message ID to string, which is what the
		// console merges over the bundled en.json.
		var translations map[string]map[string]string
		So(json.Unmarshal(data, &translations), ShouldBeNil)

		So(translations["en"]["system.name"], ShouldEqual, "Authany")
		So(len(translations["zh-CN"]), ShouldBeGreaterThanOrEqualTo, 3000)
	})
}
