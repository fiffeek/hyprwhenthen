package testutils

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/fiffeek/hyprwhenthen/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FileExists(path string) error {
	_, err := os.Stat(path)
	// nolint:wrapcheck
	return err
}

func AssertFileExists(t *testing.T, path string) {
	assert.NoError(t, FileExists(path), "file should exist")
}

func ContentSameAsFixture(t *testing.T, targetFile, fixtureFile string) error {
	// nolint:gosec
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("target content cant be read: %w", err)
	}

	// nolint:gosec
	fixtureContent, err := os.ReadFile(fixtureFile)
	if err != nil {
		return fmt.Errorf("fixture content cant be read: %w", err)
	}

	if !reflect.DeepEqual(string(targetContent), string(fixtureContent)) {
		return errors.New("contents differ")
	}

	return nil
}

func AssertContentsSameAsFixture(t *testing.T, targetFile, fixtureFile string) {
	// nolint:gosec
	targetContent, err := os.ReadFile(targetFile)
	assert.NoError(t, err, "should be able to read the target file")
	// nolint:gosec
	fixtureContent, err := os.ReadFile(fixtureFile)
	assert.NoError(t, err, "should be able to read the fixture file")
	assert.Equal(t, string(fixtureContent), string(targetContent),
		"target content should be the same as in the figture %s", fixtureContent)
}

func UpdateFixture(t *testing.T, targetFile, fixtureFile string) {
	// nolint:gosec
	targetContent, err := os.ReadFile(targetFile)
	require.NoError(t, err, "should be able to read the target file")
	// nolint:gosec
	_, err = os.ReadFile(fixtureFile)
	require.NoError(t, err, "should be able to read the fixture file")
	require.NoError(t, utils.WriteAtomic(fixtureFile, targetContent), "cant write to file")
}
