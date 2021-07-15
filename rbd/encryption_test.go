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
