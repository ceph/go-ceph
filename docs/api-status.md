<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->

# go-ceph API Stability

## Package: cephfs

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
MountInfo.SelectFilesystem | v0.20.0 | v0.22.0 | 
MountInfo.MakeDirs | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

## Package: cephfs/admin

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
FSAdmin.GetMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.SetMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.RemoveMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.ForceRemoveMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.ListMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.GetSnapshotMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.SetSnapshotMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.RemoveSnapshotMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.ForceRemoveSnapshotMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.ListSnapshotMetadata | v0.20.0 | v0.22.0 | 
FSAdmin.PinSubVolume | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
FSAdmin.PinSubVolumeGroup | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
FSAdmin.FetchVolumeInfo | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

## Package: rados

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rbd

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
MigrationPrepare | v0.20.0 | v0.22.0 | 
MigrationPrepareImport | v0.20.0 | v0.22.0 | 
MigrationExecute | v0.20.0 | v0.22.0 | 
MigrationCommit | v0.20.0 | v0.22.0 | 
MigrationAbort | v0.20.0 | v0.22.0 | 
MigrationStatus | v0.20.0 | v0.22.0 | 
SiteMirrorImageStatus.UnmarshalDescriptionJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
SiteMirrorImageStatus.DescriptionReplayStatus | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

### Deprecated APIs

Name | Deprecated in Version | Expected Removal Version | 
---- | --------------------- | ------------------------ | 
MirrorImageGlobalStatusIter.Close | v0.11.0 |  | 
Image.Open | v0.2.0 |  | 
Snapshot.Set | v0.10.0 |  | 

## Package: rbd/admin

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rgw/admin

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
API.ListBucketsWithStat | v0.20.0 | v0.22.0 | 

## Package: common/admin/manager

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/log

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/admin/nfs

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/admin/telemetry

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
NewFromConn | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.Status | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.Data | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

