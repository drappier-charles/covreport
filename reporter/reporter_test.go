package reporter_test

import (
	"testing"

	"github.com/cancue/covreport/reporter"
	"github.com/stretchr/testify/assert"
)

func TestParseCutlines(t *testing.T) {
	t.Run("should return error when cannot parse safe cutlines", func(t *testing.T) {
		var err error
		_, err = reporter.ParseCutlines("")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("not-a-number,5")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("3,not-a-number")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("3, 5")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("3 ,5")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines(" 3,5")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("3,5 ")
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("should return last number as warning cut", func(t *testing.T) {
		cutlines, err := reporter.ParseCutlines("3")
		assert.NoError(t, err)
		assert.Equal(t, 3.0, cutlines.Safe)
		assert.Equal(t, 3.0, cutlines.Warning)

		cutlines, err = reporter.ParseCutlines("3,5")
		assert.NoError(t, err)
		assert.Equal(t, 3.0, cutlines.Safe)
		assert.Equal(t, 5.0, cutlines.Warning)

		cutlines, err = reporter.ParseCutlines("3,5,7")
		assert.NoError(t, err)
		assert.Equal(t, 3.0, cutlines.Safe)
		assert.Equal(t, 7.0, cutlines.Warning)
	})
}

func TestParseIgnores(t *testing.T) {
	t.Run("should return nil with empty string", func(t *testing.T) {
		ignores := reporter.ParseIgnores("")
		assert.Nil(t, ignores)
	})

	t.Run("should work", func(t *testing.T) {
		ignores := reporter.ParseIgnores("example.txt,foobar,/.tmp")
		assert.Equal(t, 3, len(ignores))
		assert.Equal(t, "example.txt", ignores[0])
		assert.Equal(t, "foobar", ignores[1])
		assert.Equal(t, "/.tmp", ignores[2])
	})
}

func TestNewCLIConfig(t *testing.T) {
	t.Run("should have valid default values", func(t *testing.T) {
		cfg, err := reporter.NewCLIConfig()
		assert.NoError(t, err)

		assert.Equal(t, "cover.prof", cfg.Input)
		assert.Equal(t, "cover.html", cfg.Output)
		assert.Equal(t, 70.0, cfg.Cutlines.Safe)
		assert.Equal(t, 40.0, cfg.Cutlines.Warning)
		assert.Equal(t, ".", cfg.Root)
	})
}
