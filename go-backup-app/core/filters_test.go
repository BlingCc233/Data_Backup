package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterConfig_ExcludePathsBoundary(t *testing.T) {
	root := t.TempDir()
	dirFoo := filepath.Join(root, "foo")
	dirFooBar := filepath.Join(root, "foobar")
	require.NoError(t, os.MkdirAll(dirFoo, 0755))
	require.NoError(t, os.MkdirAll(dirFooBar, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dirFoo, "a.txt"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dirFooBar, "b.txt"), []byte("b"), 0644))

	fc := FilterConfig{
		ExcludePaths: []string{dirFoo},
		MaxSize:      -1,
	}

	infoFoo, err := os.Lstat(filepath.Join(dirFoo, "a.txt"))
	require.NoError(t, err)
	require.False(t, fc.ShouldInclude(filepath.Join(dirFoo, "a.txt"), infoFoo))

	infoFooBar, err := os.Lstat(filepath.Join(dirFooBar, "b.txt"))
	require.NoError(t, err)
	require.True(t, fc.ShouldInclude(filepath.Join(dirFooBar, "b.txt"), infoFooBar))
}

func TestFilterConfig_IncludePathsBoundary(t *testing.T) {
	root := t.TempDir()
	dirFoo := filepath.Join(root, "foo")
	dirFooBar := filepath.Join(root, "foobar")
	require.NoError(t, os.MkdirAll(dirFoo, 0755))
	require.NoError(t, os.MkdirAll(dirFooBar, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dirFoo, "a.txt"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dirFooBar, "b.txt"), []byte("b"), 0644))

	fc := FilterConfig{
		IncludePaths: []string{dirFoo},
		MaxSize:      -1,
	}

	infoFoo, err := os.Lstat(filepath.Join(dirFoo, "a.txt"))
	require.NoError(t, err)
	require.True(t, fc.ShouldInclude(filepath.Join(dirFoo, "a.txt"), infoFoo))

	infoFooBar, err := os.Lstat(filepath.Join(dirFooBar, "b.txt"))
	require.NoError(t, err)
	require.False(t, fc.ShouldInclude(filepath.Join(dirFooBar, "b.txt"), infoFooBar))

	infoDirFooBar, err := os.Lstat(dirFooBar)
	require.NoError(t, err)
	require.False(t, fc.ShouldInclude(dirFooBar, infoDirFooBar))
}

func TestFilterConfig_IncludePathsAncestorDirAllowed(t *testing.T) {
	root := t.TempDir()
	dirFoo := filepath.Join(root, "foo")
	dirFooBar := filepath.Join(dirFoo, "bar")
	require.NoError(t, os.MkdirAll(dirFooBar, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dirFooBar, "a.txt"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dirFoo, "b.txt"), []byte("b"), 0644))

	fc := FilterConfig{
		IncludePaths: []string{dirFooBar},
		MaxSize:      -1,
	}

	infoDirFoo, err := os.Lstat(dirFoo)
	require.NoError(t, err)
	require.True(t, fc.ShouldInclude(dirFoo, infoDirFoo), "ancestor dir should be included to allow traversal")

	infoFooBarFile, err := os.Lstat(filepath.Join(dirFooBar, "a.txt"))
	require.NoError(t, err)
	require.True(t, fc.ShouldInclude(filepath.Join(dirFooBar, "a.txt"), infoFooBarFile))

	infoFooFile, err := os.Lstat(filepath.Join(dirFoo, "b.txt"))
	require.NoError(t, err)
	require.False(t, fc.ShouldInclude(filepath.Join(dirFoo, "b.txt"), infoFooFile))
}

