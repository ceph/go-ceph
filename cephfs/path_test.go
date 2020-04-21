package cephfs

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeDir(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	dir1 := mount.CurrentDir()
	assert.NotNil(t, dir1)

	err := mount.MakeDir("/asdf", 0755)
	assert.NoError(t, err)

	err = mount.ChangeDir("/asdf")
	assert.NoError(t, err)

	dir2 := mount.CurrentDir()
	assert.NotNil(t, dir2)

	assert.NotEqual(t, dir1, dir2)
	assert.Equal(t, dir1, "/")
	assert.Equal(t, dir2, "/asdf")

	err = mount.ChangeDir("/")
	assert.NoError(t, err)
	err = mount.RemoveDir("/asdf")
	assert.NoError(t, err)
}

func TestRemoveDir(t *testing.T) {
	useMount(t)

	dirname := "one"
	localPath := path.Join(CephMountDir, dirname)
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	_, err = os.Stat(localPath)
	assert.NoError(t, err)

	err = mount.RemoveDir(dirname)
	assert.NoError(t, err)

	_, err = os.Stat(localPath)
	assert.EqualError(t, err,
		fmt.Sprintf("stat %s: no such file or directory", localPath))
}

func TestLink(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	t.Run("directory operations", func(t *testing.T) {
		err := mount.Link("/", "/")
		// Error directory operations not allowed
		assert.Error(t, err)

		dirname := "/a"
		err = mount.MakeDir(dirname, 0755)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, mount.RemoveDir(dirname)) }()

		err = mount.Link(dirname, "/")
		// Error, directory operations not allowed
		assert.Error(t, err)

		filename := "/tmp/file"
		err = mount.Link(dirname, filename)
		// Error, directory operations not allowed
		assert.Error(t, err)

		err = mount.Link(filename, dirname)
		// Error, file does not exist
		assert.Error(t, err)

		/*
			file, _ := os.Create(filename)
			defer func() {
				assert.NoError(t, file.Close())
				assert.NoError(t, mount.Unlink(filename))
			}()
			err = mount.Link(filename, dirname)
			// No Error, file can link to directory
			assert.NoError(t, err)

			err = mount.Link(filename, dirname)
			// Error, link already exist
			assert.Error(t, err)
		*/
	})

	t.Run("file operations", func(t *testing.T) {
		filename1 := "/tmp/file1"
		err := mount.Link(filename1, "/tmp/hardlnk")
		// Error, file does not exist
		assert.Error(t, err)
		/*
			file1, _ := os.Create(filename1)
			defer func() {
				assert.NoError(t, file1.Close())
				assert.NoError(t, mount.Unlink(filename1))
			}()
			err = mount.Link(filename1, "/tmp/hardlnk")
			defer assert.NoError(t, mount.Unlink("/tmp/hardlnk"))
			// No error, normal link operation
			assert.NoError(t, err)
			// Verify that link got created
			_, err = os.Stat("/tmp/hardlnk")
			assert.NoError(t, err)

			filename2 := "/tmp/file2"
			file2, _ := os.Create(filename2)
			defer func() {
				assert.NoError(t, file2.Close())
				assert.NoError(t, mount.Unlink(filename2))
			}()
			err = mount.Link(filename1, filename2)
			// Error, destination already exists.
			assert.Error(t, err)
		*/
	})
}

func TestUnlink(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	filename := "/tmp/file"
	err := mount.Unlink(filename)
	// Error, file does not exist
	assert.Error(t, err)
	/*
		file, _ := os.Create(filename)
		file.Close()
		assert.NoError(t, mount.Link(filename, "/tmp/hardlnk"))
		err = mount.Unlink(filename)
		// No Error, link will be removed
		assert.NoError(t, err)
		err = mount.Unlink("/tmp/hardlnk")
		// No Error, link will be removed
		assert.NoError(t, err)
	*/
	dirname := "/a"
	err = mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir(dirname)) }()

	err = mount.Unlink(dirname)
	// Error, not permitted on directory
	assert.Error(t, err)
}

func TestSymlink(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	filename1 := "file1"
	filename2 := "file2"

	err := mount.Symlink(filename1, filename2)
	defer func() {
		assert.NoError(t, mount.Unlink(filename2))
		//assert.NoError(t, mount.Unlink(filename1))
	}()
	// Error, file doesn't exist
	assert.NoError(t, err)
	/*
		existingFile := "tmp/existing"
		existing, _ := os.Create(existingFile)
		err = mount.Symlink(filename1, existingFile)
		defer func() {
			existing.Close()
			assert.NoError(t, mount.Unlink(existingFile))
		}()
		// Error, file already exists
		assert.Error(t, err)
	*/
	// 1. Create a directory
	// 2. Create a symlink to that directory
	// 3. Create a file inside symlink.
	// 4. Ensure that it is a file not a directory
	dirname := "/a"
	err = mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir(dirname)) }()

	err = mount.Symlink(dirname, "symlink")
	assert.NoError(t, err)
	/*
			filename := "tmp/symlink/file"
			file, _ := os.Create(filename)
			defer file.Close()
			var fileInfo os.FileInfo
			fileInfo, err = os.Stat("/tmp/symlink")
			assert.EqualError(t, err,
				fmt.Sprintf("Is Regular: %v", !fileInfo.IsDir()))

		defer func() { assert.NoError(t, mount.Unlink("/tmp/symlink")) }()
	*/
}

func TestReadlink(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	path1 := "path1"
	_, err := mount.Readlink(path1)
	// Error, given path is does not exist.
	assert.Error(t, err)

	/*
		filename := "file1.txt"
		file, _ := os.Create(filename)
		defer func() {
			assert.NoError(t, mount.Unlink(filename))
			assert.NoError(t, file.Close())
		}()
		_, err = mount.Readlink(filename)
		// Error, given path is not symbolic link
		assert.Error(t, err)
	*/
	// Symbolic link
	path2 := "path2"
	assert.NoError(t, mount.Symlink(path1, path2))
	defer func() { assert.NoError(t, mount.Unlink(path2)) }()
	_, err = mount.Readlink(path2)
	// No Error, path2 is a symbolic link
	assert.NoError(t, err)

	// Hard link
	/*
		path3 := "path3"
		path4 := "path4"
		p, _ := os.Create(path3)
		defer func() {
			assert.NoError(t, mount.Unlink(path3))
			assert.NoError(t, p.Close())
		}()
		assert.NoError(t, mount.Link(path3, path4))
		defer func() { assert.NoError(t, mount.Unlink(path4)) }()
		_, err = mount.Readlink(path4)
		// Error, path4 is not symbolic link
		assert.Error(t, err)
	*/
}
