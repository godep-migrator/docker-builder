package builder_test

import (
	. "github.com/modcloth/docker-builder/builder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"path/filepath"
)

var _ = Describe("Sanitize Builderfile path", func() {

	var (
		dotDotPath       = "Specs/fixtures/../fixtures/repodir/foo/bar/Bobfile"
		symlinkPath      = "Specs/fixtures/repodir/foo/symlink/Bobfile"
		bogusPath        = "foobarbaz"
		validPath        = "Specs/fixtures/repodir/foo/bar/Bobfile"
		absValidPath, _  = filepath.Abs("../" + validPath)
		cleanedValidPath = filepath.Clean(absValidPath)
	)

	Context("when the path is bogus", func() {
		It(`returns an error when the path contains ".."`, func() {
			config, _ := NewBuildConfig(dotDotPath, "..")
			_, err := SanitizeBuilderfilePath(config)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(DotDotSanitizeErrorMessage))
		})

		It("returns an error when the path contains symlinks", func() {
			config, _ := NewBuildConfig(symlinkPath, "..")
			_, err := SanitizeBuilderfilePath(config)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(SymlinkSanitizeErrorMessage))
		})

		It("returns an error when the path is invalid", func() {
			config, _ := NewBuildConfig(bogusPath, "..")
			_, err := SanitizeBuilderfilePath(config)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(InvalidPathSanitizeErrorMessage))
		})
	})

	Context("when the path is valid", func() {
		It("does not return an error", func() {
			config, _ := NewBuildConfig(validPath, "..")
			_, err := SanitizeBuilderfilePath(config)
			Expect(err).To(BeNil())
		})

		It("returns a cleaned version of the path", func() {
			config, _ := NewBuildConfig(validPath, "..")
			path, _ := SanitizeBuilderfilePath(config)
			Expect(path).To(Equal(cleanedValidPath))
		})
	})
})
