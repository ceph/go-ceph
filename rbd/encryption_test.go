//go:build !octopus && !nautilus
// +build !octopus,!nautilus

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptionFormat(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	testImageSize := uint64(1 << 23) // format requires more than 4194304 bytes
	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	workingImage, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	var opts EncryptionOptionsLUKS1
	opts.Alg = EncryptionAlgorithmAES256
	opts.Passphrase = ([]byte)("test-password")
	err = workingImage.EncryptionFormat(opts)
	assert.NoError(t, err)

	err = workingImage.Close()
	assert.NoError(t, err)
	err = workingImage.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestEncryptionLoad(t *testing.T) {
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
	testImageSize := uint64(1 << 23) // format requires more than 4194304 bytes
	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	var opts EncryptionOptionsLUKS1
	opts.Alg = EncryptionAlgorithmAES256
	opts.Passphrase = ([]byte)("test-password")
	err = img.EncryptionFormat(opts)
	assert.NoError(t, err)

	// close the image so we can reopen it and load the encryption info
	// then write some encrypted data at the end of the image
	err = img.Close()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Remove())
	}()

	testData := []byte("Hi rbd! Nice to talk through go-ceph :)")
	var offset int64

	t.Run("prepare", func(t *testing.T) {
		img, err = OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionLoad(opts)
		assert.NoError(t, err)

		stats, err := img.Stat()
		require.NoError(t, err)
		offset = int64(stats.Size) - int64(len(testData))

		nOut, err := img.WriteAt(testData, offset)
		assert.Equal(t, len(testData), nOut)
		assert.NoError(t, err)
	})

	t.Run("readEnc", func(t *testing.T) {
		require.NotEqual(t, offset, 0)
		// Re-open the image, load the encryption format, and read the encrypted data
		img, err = OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionLoad(opts)
		assert.NoError(t, err)

		inData := make([]byte, len(testData))
		nIn, err := img.ReadAt(inData, offset)
		assert.Equal(t, nIn, len(inData))
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
		assert.Equal(t, nIn, len(inData))
		assert.NotEqual(t, inData, testData)
		assert.NoError(t, err)
	})
}

func TestEncryptedResize(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)
	defer conn.DeletePool(poolname)

	imageName := "resizeme"
	imageSize := uint64(50) * 1024 * 1024
	encOpts := EncryptionOptionsLUKS2{
		Alg:        EncryptionAlgorithmAES256,
		Passphrase: []byte("test-password"),
	}

	t.Run("create", func(t *testing.T) {
		ioctx, err := conn.OpenIOContext(poolname)
		require.NoError(t, err)
		defer ioctx.Destroy()

		err = CreateImage(ioctx, imageName, imageSize, NewRbdImageOptions())
		require.NoError(t, err)

		image, err := OpenImage(ioctx, imageName, NoSnapshot)
		require.NoError(t, err)
		defer image.Close()

		s, err := image.GetSize()
		require.NoError(t, err)
		t.Logf("image size before encryption: %d", s)

		err = image.EncryptionFormat(encOpts)
		require.NoError(t, err)

		s, err = image.GetSize()
		require.NoError(t, err)
		t.Logf("image size after encryption: %d", s)
	})

	t.Run("resize", func(t *testing.T) {
		ioctx, err := conn.OpenIOContext(poolname)
		require.NoError(t, err)
		defer ioctx.Destroy()

		image, err := OpenImage(ioctx, imageName, NoSnapshot)
		require.NoError(t, err)
		defer image.Close()

		err = image.EncryptionLoad(encOpts)
		assert.NoError(t, err)

		s, err := image.GetSize()
		require.NoError(t, err)
		t.Logf("image size before resize: %d", s)
		assert.NotEqual(t, s, imageSize)

		err = image.Resize(imageSize)
		assert.NoError(t, err)

		s, err = image.GetSize()
		require.NoError(t, err)
		t.Logf("image size after resize of encrypted image: %d", s)
	})
}
