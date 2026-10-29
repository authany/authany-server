package service

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNormalizeProjectName(t *testing.T) {
	Convey("NormalizeProjectName", t, func() {
		Convey("trims surrounding spaces", func() {
			name, err := NormalizeProjectName("  Modeltaps 登录  ")
			So(err, ShouldBeNil)
			So(name, ShouldEqual, "Modeltaps 登录")
		})

		Convey("accepts an empty name", func() {
			name, err := NormalizeProjectName("   ")
			So(err, ShouldBeNil)
			So(name, ShouldEqual, "")
		})

		Convey("counts characters, not bytes", func() {
			_, err := NormalizeProjectName(strings.Repeat("名", ProjectNameMaxLength))
			So(err, ShouldBeNil)
			_, err = NormalizeProjectName(strings.Repeat("名", ProjectNameMaxLength+1))
			So(err, ShouldBeError, "project name is too long")
		})

		Convey("rejects control characters", func() {
			_, err := NormalizeProjectName("a\nb")
			So(err, ShouldBeError, "project name contains control characters")
			_, err = NormalizeProjectName("abc\u202edef")
			So(err, ShouldBeError, "project name contains control characters")
			_, err = NormalizeProjectName("\u200b")
			So(err, ShouldBeError, "project name contains control characters")
		})

		Convey("keeps emoji joined by U+200D", func() {
			name, err := NormalizeProjectName("Dev \U0001F468\u200d\U0001F4BB")
			So(err, ShouldBeNil)
			So(name, ShouldEqual, "Dev \U0001F468\u200d\U0001F4BB")
		})
	})
}
