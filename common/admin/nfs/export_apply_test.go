//go:build !(nautilus || octopus || pacific || quincy) && ceph_preview

package nfs

func (suite *NFSAdminSuite) TestApplyExportInfo() {
	require := suite.Require()
	ra := radosConnector.Get(suite.T())
	nfsa := NewFromConn(ra)

	// Create initial export with ClientAddr
	res, err := nfsa.CreateCephFSExport(CephFSExportSpec{
		FileSystemName: suite.fileSystemName,
		ClusterID:      suite.clusterID,
		PseudoPath:     "/applytest",
		Path:           "/january",
		ClientAddr:     []string{"192.168.1.0/24"},
	})
	require.NoError(err)
	require.Equal("/applytest", res.Bind)

	defer func() {
		err = nfsa.RemoveExport(suite.clusterID, "/applytest")
		require.NoError(err)
	}()

	// Get initial export info
	info, err := nfsa.ExportInfo(suite.clusterID, "/applytest")
	require.NoError(err)
	require.Equal("/applytest", info.PseudoPath)
	require.Equal("/january", info.Path)
	require.Len(info.Clients, 1)
	require.Contains(info.Clients[0].Addresses, "192.168.1.0/24")

	// Modify the export info and apply it
	// Several attributes cause an NFS-server restart, this doesn't work in
	// the CI as there is no orchestrator handling it.
	// This test should not modify user_id, fs_name, path or pseudo.
	info.Clients = []ClientInfo{
		{
			Addresses:  []string{"10.0.0.0/8", "172.16.0.0/12"},
			AccessType: "RW",
			Squash:     "none",
		},
	}
	err = nfsa.ApplyExportInfo(suite.clusterID, info)
	require.NoError(err)

	// Verify the update by getting export info
	updatedInfo, err := nfsa.ExportInfo(suite.clusterID, "/applytest")
	require.NoError(err)
	require.Equal("/applytest", updatedInfo.PseudoPath)
	require.Equal("/january", updatedInfo.Path)
	require.Len(updatedInfo.Clients, 1)
	require.Len(updatedInfo.Clients[0].Addresses, 2)
	require.Contains(updatedInfo.Clients[0].Addresses, "10.0.0.0/8")
	require.Contains(updatedInfo.Clients[0].Addresses, "172.16.0.0/12")
}
