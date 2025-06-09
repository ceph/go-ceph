//go:build !octopus && !pacific && !quincy

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptionLoad2(t *testing.T) {
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
		Passphrase: []byte("test-password"),
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

	testData := []byte("Jinxed wizards pluck ivy from the big quilt")
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

	t.Run("readEnc", func(t *testing.T) {
		require.NotEqual(t, offset, 0)
		// Re-open the image, load the encryption format, and read the encrypted data
		img, err = OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionLoad2([]EncryptionOptions{encOpts})
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

func TestEncryptionLoad2WithParents(t *testing.T) {
	dlength := int64(32)
	testData1 := []byte("Very nice object ahead of change")
	testData2 := []byte("A nice object encryption applied")
	testData3 := []byte("A good object encryption abounds")
	testData4 := []byte("Another portion is here and well")
	written := [][]byte{}
	assert.EqualValues(t, len(testData1), dlength)
	assert.EqualValues(t, len(testData2), dlength)
	assert.EqualValues(t, len(testData3), dlength)
	assert.EqualValues(t, len(testData4), dlength)

	encOpts1 := EncryptionOptionsLUKS1{
		Alg:        EncryptionAlgorithmAES128,
		Passphrase: []byte("test-password"),
	}
	encOpts2 := EncryptionOptionsLUKS2{
		Alg:        EncryptionAlgorithmAES128,
		Passphrase: []byte("test-password"),
	}
	encOpts3 := EncryptionOptionsLUKS2{
		Alg:        EncryptionAlgorithmAES256,
		Passphrase: []byte("something-stronger"),
	}

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
	testImageSize := uint64(256) * 1024 * 1024
	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	t.Run("prepare", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()

		_, err = img.WriteAt(testData1, 0)
		assert.NoError(t, err)
		written = append(written, testData1)
	})

	t.Run("createClone1", func(t *testing.T) {
		require.Len(t, written, 1)
		parent, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer parent.Close()
		snap, err := parent.CreateSnapshot("sn1")
		assert.NoError(t, err)
		err = snap.Protect()
		assert.NoError(t, err)

		err = CloneImage(ioctx, name, "sn1", ioctx, name+"clone1", options)
		assert.NoError(t, err)

		img, err := OpenImage(ioctx, name+"clone1", NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionFormat(encOpts1)
		assert.NoError(t, err)

		err = img.EncryptionLoad2([]EncryptionOptions{encOpts1})
		assert.NoError(t, err)
		_, err = img.WriteAt(testData2, dlength)
		assert.NoError(t, err)
		written = append(written, testData2)
	})

	t.Run("createClone2", func(t *testing.T) {
		require.Len(t, written, 2)
		parentName := name + "clone1"
		cloneName := name + "clone2"

		parent, err := OpenImage(ioctx, parentName, NoSnapshot)
		assert.NoError(t, err)
		defer parent.Close()
		snap, err := parent.CreateSnapshot("sn2")
		assert.NoError(t, err)
		err = snap.Protect()
		assert.NoError(t, err)

		err = CloneImage(ioctx, parentName, "sn2", ioctx, cloneName, options)
		assert.NoError(t, err)

		img, err := OpenImage(ioctx, cloneName, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionFormat(encOpts2)
		assert.NoError(t, err)

		err = img.EncryptionLoad2([]EncryptionOptions{encOpts2, encOpts1})
		assert.NoError(t, err)
		_, err = img.WriteAt(testData3, dlength*2)
		assert.NoError(t, err)
		written = append(written, testData3)
	})

	t.Run("createClone3", func(t *testing.T) {
		require.Len(t, written, 3)
		parentName := name + "clone2"
		cloneName := name + "clone3"

		parent, err := OpenImage(ioctx, parentName, NoSnapshot)
		assert.NoError(t, err)
		defer parent.Close()
		snap, err := parent.CreateSnapshot("sn3")
		assert.NoError(t, err)
		err = snap.Protect()
		assert.NoError(t, err)

		err = CloneImage(ioctx, parentName, "sn3", ioctx, cloneName, options)
		assert.NoError(t, err)

		img, err := OpenImage(ioctx, cloneName, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()
		err = img.EncryptionFormat(encOpts3)
		assert.NoError(t, err)

		err = img.EncryptionLoad2([]EncryptionOptions{
			encOpts3, encOpts2, encOpts1,
		})
		assert.NoError(t, err)
		_, err = img.WriteAt(testData4, dlength*3)
		assert.NoError(t, err)
		written = append(written, testData4)
	})

	t.Run("readAll", func(t *testing.T) {
		require.Len(t, written, 4)
		img, err := OpenImage(ioctx, name+"clone3", NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()

		err = img.EncryptionLoad2([]EncryptionOptions{
			encOpts3, encOpts2, encOpts1,
		})
		assert.NoError(t, err)

		inData := make([]byte, int(dlength))
		for idx, td := range written {
			n, err := img.ReadAt(inData, int64(idx)*dlength)
			assert.NoError(t, err)
			assert.EqualValues(t, dlength, n)
			assert.Equal(t, inData, td)
		}
	})
}
