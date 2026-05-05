//go:build ceph_preview

package rados

import (
	"encoding/binary"
	"hash/crc32"

	xxhash32 "github.com/pierrec/xxHash/xxHash32"
	xxhash64 "github.com/pierrec/xxHash/xxHash64"
	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestChecksum() {
	suite.SetupConnection()

	var (
		oid      = suite.GenObjectName()
		contents = suite.RandomBytes(3 * 1024)
	)

	// write some random data to the object
	assert.NoError(suite.T(), suite.ioctx.Write(oid, contents, 0))

	// test each checksum type
	suite.Run("CRC32", func() {
		init := []byte{0xff, 0xff, 0xff, 0xff}

		dst := make([]byte, 4+4)
		assert.NoError(suite.T(), suite.ioctx.Checksum(oid, ChecksumTypeCRC32C, dst, &ChecksumOptions{InitValue: init}))

		chunks := binary.LittleEndian.Uint32(dst[:4])
		assert.Equal(suite.T(), uint32(1), chunks)

		// Note: the CRC32 Go standard library produces checksum results with the final XOR already applied,
		// while rados_checksum returns raw results; so here we do additional processing to assert correctness.
		sum := binary.LittleEndian.Uint32(dst[4:]) ^ 0xffffffff
		want := crc32.Checksum(contents, crc32.MakeTable(crc32.Castagnoli))
		assert.Equal(suite.T(), want, sum)
	})

	suite.Run("XXHash32", func() {
		init := []byte{0x0, 0x0, 0x0, 0x0}
		seed := binary.LittleEndian.Uint32(init)

		hash := xxhash32.New(seed)
		_, err := hash.Write(contents)
		assert.NoError(suite.T(), err)
		want := binary.LittleEndian.Uint32(hash.Sum(nil))

		dst := make([]byte, 4+4)
		assert.NoError(suite.T(), suite.ioctx.Checksum(oid, ChecksumTypeXXHash32, dst, nil))

		chunks := binary.LittleEndian.Uint32(dst[:4])
		assert.Equal(suite.T(), uint32(1), chunks)

		sum := binary.LittleEndian.Uint32(dst[4:])
		assert.Equal(suite.T(), want, sum)
	})

	suite.Run("XXHash64", func() {
		init := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
		seed := binary.LittleEndian.Uint64(init)

		xxh := xxhash64.New(seed)
		_, err := xxh.Write(contents)
		assert.NoError(suite.T(), err)
		want := binary.LittleEndian.Uint64(xxh.Sum(nil))

		dst := make([]byte, 4+8)
		assert.NoError(suite.T(), suite.ioctx.Checksum(oid, ChecksumTypeXXHash64, dst, nil))

		chunks := binary.LittleEndian.Uint32(dst[:4])
		assert.Equal(suite.T(), uint32(1), chunks)

		sum := binary.LittleEndian.Uint64(dst[4:])
		assert.Equal(suite.T(), want, sum)
	})
}
func (suite *RadosTestSuite) TestChecksumWithOpts() {
	suite.SetupConnection()

	var (
		off        uint64 = 256
		chunkSize  uint64 = 1024
		chunkCount uint64 = 2

		oid      = suite.GenObjectName()
		contents = suite.RandomBytes(int(off + chunkSize*(chunkCount+1))) // write one extra chunk, to be ignored
	)

	// write some random data to the object
	assert.NoError(suite.T(), suite.ioctx.Write(oid, contents, 0))

	// test each checksum type with optional parameters
	suite.Run("CRC32", func() {
		init := []byte{0xff, 0xff, 0xff, 0xff}

		table := crc32.MakeTable(crc32.Castagnoli)
		want := []uint32{
			crc32.Checksum(contents[off:off+chunkSize], table),
			crc32.Checksum(contents[off+chunkSize:off+chunkSize*2], table),
		}

		dst := make([]byte, 4+4*chunkCount) // allocate space for multiple uint32 checksum values
		assert.NoError(suite.T(), suite.ioctx.Checksum(oid, ChecksumTypeCRC32C, dst, &ChecksumOptions{
			InitValue: init,
			ChunkSize: chunkSize,
			Off:       off,
			Len:       chunkSize * chunkCount,
		}))

		chunks := binary.LittleEndian.Uint32(dst[:4])
		assert.Equal(suite.T(), uint32(chunkCount), chunks)

		got := []uint32{
			binary.LittleEndian.Uint32(dst[4:8]) ^ 0xffffffff,
			binary.LittleEndian.Uint32(dst[8:12]) ^ 0xffffffff,
		}
		for i := range want {
			assert.Equal(suite.T(), want[i], got[i])
		}
	})

	suite.Run("XXHash32", func() {
		init := []byte{0xff, 0xff, 0xff, 0xff}
		seed := binary.LittleEndian.Uint32(init)

		var want [2]uint32
		xxh := xxhash32.New(seed)

		_, err := xxh.Write(contents[off : off+chunkSize])
		assert.NoError(suite.T(), err)
		want[0] = binary.LittleEndian.Uint32(xxh.Sum(nil))
		xxh.Reset()
		_, err = xxh.Write(contents[off+chunkSize : off+chunkSize*2])
		assert.NoError(suite.T(), err)
		want[1] = binary.LittleEndian.Uint32(xxh.Sum(nil))

		dst := make([]byte, 4+4*chunkCount) // allocate space for multiple uint32 checksum values
		assert.NoError(suite.T(), suite.ioctx.Checksum(oid, ChecksumTypeXXHash32, dst, &ChecksumOptions{
			InitValue: init,
			ChunkSize: chunkSize,
			Off:       off,
			Len:       chunkSize * chunkCount,
		}))

		chunks := binary.LittleEndian.Uint32(dst[:4])
		assert.Equal(suite.T(), uint32(chunkCount), chunks)

		got := []uint32{
			binary.LittleEndian.Uint32(dst[4:8]),
			binary.LittleEndian.Uint32(dst[8:12]),
		}
		for i := range want {
			assert.Equal(suite.T(), want[i], got[i])
		}
	})

	suite.Run("XXHash64", func() {
		init := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		seed := binary.LittleEndian.Uint64(init)

		var want [2]uint64
		xxh := xxhash64.New(seed)

		_, err := xxh.Write(contents[off : off+chunkSize])
		assert.NoError(suite.T(), err)
		want[0] = binary.LittleEndian.Uint64(xxh.Sum(nil))
		xxh.Reset()
		_, err = xxh.Write(contents[off+chunkSize : off+chunkSize*2])
		assert.NoError(suite.T(), err)
		want[1] = binary.LittleEndian.Uint64(xxh.Sum(nil))

		dst := make([]byte, 4+8*chunkCount) // allocate space for multiple uint64 checksum values
		assert.NoError(suite.T(), suite.ioctx.Checksum(oid, ChecksumTypeXXHash64, dst, &ChecksumOptions{
			InitValue: init,
			ChunkSize: chunkSize,
			Off:       off,
			Len:       chunkSize * chunkCount,
		}))

		chunks := binary.LittleEndian.Uint32(dst[:4])
		assert.Equal(suite.T(), uint32(chunkCount), chunks)

		got := []uint64{
			binary.LittleEndian.Uint64(dst[4:12]),
			binary.LittleEndian.Uint64(dst[12:20]),
		}
		for i := range want {
			assert.Equal(suite.T(), want[i], got[i])
		}
	})
}
