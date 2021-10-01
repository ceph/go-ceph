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
	img, err = OpenImage(ioctx, name, NoSnapshot)
	err = img.EncryptionLoad(opts)
	assert.NoError(t, err)

	outData := []byte("Hi rbd! Nice to talk through go-ceph :)")

	stats, err := img.Stat()
	require.NoError(t, err)
	offset := int64(stats.Size) - int64(len(outData))

	nOut, err := img.WriteAt(outData, offset)
	assert.Equal(t, len(outData), nOut)
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	// Re-open the image, load the encryption format, and read the encrypted data
	img, err = OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	err = img.EncryptionLoad(opts)
	assert.NoError(t, err)

	inData := make([]byte, len(outData))
	nIn, err := img.ReadAt(inData, offset)
	assert.Equal(t, nIn, len(inData))
	assert.Equal(t, inData, outData)
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	// Re-open the image and attempt to read the encrypted data without loading the encryption
	img, err = OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	nIn, err = img.ReadAt(inData, offset)
	assert.Equal(t, nIn, len(inData))
	assert.NotEqual(t, inData, outData)
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)
	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}
