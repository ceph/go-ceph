//go:build !octopus && !pacific && !quincy

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptionOptionsLUKS(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	name := GetUUID()
	testImageSize := uint64(50) * 1024 * 1024
	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	encOpts := EncryptionOptionsLUKS2{
		Alg:        EncryptionAlgorithmAES256,
		Passphrase: []byte("sesame-123-luks-ury"),
	}
	err = img.EncryptionFormat(encOpts)
	assert.NoError(t, err)

	// close the image so we can reopen it and load the encryption info
	// then write some encrypted data at the end of the image
	err = img.Close()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Remove())
	}()

	testData := []byte("Another day another image to unlock")
	var offset int64

	t.Run("prepare", func(t *testing.T) {
		img, err = OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionLoad2([]EncryptionOptions{encOpts})
		assert.NoError(t, err)

		stats, err := img.Stat()
		require.NoError(t, err)
		offset = int64(stats.Size) - int64(len(testData))

		nOut, err := img.WriteAt(testData, offset)
		assert.Equal(t, len(testData), nOut)
		assert.NoError(t, err)
	})

	unlock := []EncryptionOptions{
		EncryptionOptionsLUKS{Passphrase: []byte("sesame-123-luks-ury")},
	}

	t.Run("readEncLoad", func(t *testing.T) {
		require.NotEqual(t, offset, 0)
		// Re-open the image, using the generic luks encryption options
		img, err = OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionLoad(unlock[0])
		assert.NoError(t, err)

		inData := make([]byte, len(testData))
		nIn, err := img.ReadAt(inData, offset)
		assert.Equal(t, nIn, len(testData))
		assert.Equal(t, inData, testData)
		assert.NoError(t, err)
	})

	t.Run("readEncLoad2", func(t *testing.T) {
		require.NotEqual(t, offset, 0)
		// Re-open the image, using the generic luks encryption options
		img, err = OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionLoad2(unlock)
		assert.NoError(t, err)

		inData := make([]byte, len(testData))
		nIn, err := img.ReadAt(inData, offset)
		assert.Equal(t, nIn, len(testData))
		assert.Equal(t, inData, testData)
		assert.NoError(t, err)
	})

	t.Run("noEnc", func(t *testing.T) {
		require.NotEqual(t, offset, 0)
		// Re-open the image and attempt to read the encrypted data without loading the encryption
		img, err = OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()

		inData := make([]byte, len(testData))
		nIn, err := img.ReadAt(inData, offset)
		assert.Equal(t, nIn, len(testData))
		assert.NotEqual(t, inData, testData)
		assert.NoError(t, err)
	})
}
