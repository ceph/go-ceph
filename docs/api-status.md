<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->

# go-ceph API Stability

## Package: cephfs

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
Wrap | v0.33.0 | v0.35.0 | 
MountWrapper.SetTracing | v0.33.0 | v0.35.0 | 
MountWrapper.Open | v0.33.0 | v0.35.0 | 

## Package: cephfs/admin

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
FSAdmin.SubVolumeSnapshotPath | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

## Package: rados

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rbd

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
Image.DiffIterateByID | v0.33.0 | v0.35.0 | 

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
API.CheckBucketIndex | v0.33.0 | v0.35.0 | 

## Package: common/admin/manager

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/log

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/admin/nfs

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rados/striper

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/admin/smb

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
NewFromConn | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
SimplePlacement | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Cluster.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Cluster.Intent | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Cluster.Identity | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Cluster.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Cluster.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Cluster.SetPlacement | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewUserCluster | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewActiveDirectoryCluster | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewClusterToRemove | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
JoinAuth.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
JoinAuth.Intent | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
JoinAuth.Identity | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
JoinAuth.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
JoinAuth.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
JoinAuth.SetAuth | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewJoinAuth | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewLinkedJoinAuth | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewJoinAuthToRemove | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ResourceType.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ResourceType.String | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ResourceID.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ResourceID.String | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ChildResourceID.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ChildResourceID.String | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.Show | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.Apply | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ValidateResources | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.RemoveCluster | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.RemoveShare | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.RemoveJoinAuth | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.RemoveUsersAndGroups | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Result.UnmarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Result.Ok | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Result.Resource | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Result.Message | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Result.Error | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Result.State | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Result.Dump | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ResultGroup.Ok | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ResultGroup.ErrorResults | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ResultGroup.Error | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ShareAccess.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Share.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Share.Intent | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Share.Identity | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Share.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Share.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Share.SetCephFS | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewShare | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewShareToRemove | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
UsersAndGroups.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
UsersAndGroups.Intent | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
UsersAndGroups.Identity | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
UsersAndGroups.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
UsersAndGroups.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
UsersAndGroups.SetValues | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewUsersAndGroups | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewLinkedUsersAndGroups | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewUsersAndGroupsToRemove | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

