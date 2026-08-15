package webp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daanv2/go-webp/pkg/encoding/webp"
	"github.com/stretchr/testify/require"
)

func Test_Reading_Examples(t *testing.T) {
	dir, err := filepath.Abs("../../examples")
	require.NoError(t, err)

	files, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, dirEntry := range files {
		if filepath.Ext(dirEntry.Name()) != ".webp" {
			continue
		}

		fp := filepath.Clean(filepath.Join(dir, dirEntry.Name()))
		t.Run("reading_"+dirEntry.Name(), func(t *testing.T) {
			f, err := os.Open(fp)
			require.NoError(t, err)
			defer func() {
				_ = f.Close()
			}()

			data, err := webp.Read(f)
			require.NoError(t, err)
			require.NotNil(t, data)
		})
	}
}
