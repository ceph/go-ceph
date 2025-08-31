<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->

# go-ceph API Stability

## Package: cephfs

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
File.Fd | v0.35.0 | v0.37.0 | 
File.Futime | v0.35.0 | v0.37.0 | 
File.Futimens | v0.35.0 | v0.37.0 | 
File.Futimes | v0.35.0 | v0.37.0 | 

## Package: cephfs/admin

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
FSAdmin.SubVolumeSnapshotPath | v0.34.0 | v0.36.0 | 

## Package: rados

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rbd

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
Image.GetDataPoolID | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

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
NewFromConn | v0.34.0 | v0.36.0 | 
SimplePlacement | v0.34.0 | v0.36.0 | 
Cluster.Type | v0.34.0 | v0.36.0 | 
Cluster.Intent | v0.34.0 | v0.36.0 | 
Cluster.Identity | v0.34.0 | v0.36.0 | 
Cluster.MarshalJSON | v0.34.0 | v0.36.0 | 
Cluster.Validate | v0.34.0 | v0.36.0 | 
Cluster.SetPlacement | v0.34.0 | v0.36.0 | 
NewUserCluster | v0.34.0 | v0.36.0 | 
NewActiveDirectoryCluster | v0.34.0 | v0.36.0 | 
NewClusterToRemove | v0.34.0 | v0.36.0 | 
JoinAuth.Type | v0.34.0 | v0.36.0 | 
JoinAuth.Intent | v0.34.0 | v0.36.0 | 
JoinAuth.Identity | v0.34.0 | v0.36.0 | 
JoinAuth.Validate | v0.34.0 | v0.36.0 | 
JoinAuth.MarshalJSON | v0.34.0 | v0.36.0 | 
JoinAuth.SetAuth | v0.34.0 | v0.36.0 | 
NewJoinAuth | v0.34.0 | v0.36.0 | 
NewLinkedJoinAuth | v0.34.0 | v0.36.0 | 
NewJoinAuthToRemove | v0.34.0 | v0.36.0 | 
ResourceType.Type | v0.34.0 | v0.36.0 | 
ResourceType.String | v0.34.0 | v0.36.0 | 
ResourceID.Type | v0.34.0 | v0.36.0 | 
ResourceID.String | v0.34.0 | v0.36.0 | 
ChildResourceID.Type | v0.34.0 | v0.36.0 | 
ChildResourceID.String | v0.34.0 | v0.36.0 | 
Admin.Show | v0.34.0 | v0.36.0 | 
Admin.Apply | v0.34.0 | v0.36.0 | 
ValidateResources | v0.34.0 | v0.36.0 | 
Admin.RemoveCluster | v0.34.0 | v0.36.0 | 
Admin.RemoveShare | v0.34.0 | v0.36.0 | 
Admin.RemoveJoinAuth | v0.34.0 | v0.36.0 | 
Admin.RemoveUsersAndGroups | v0.34.0 | v0.36.0 | 
Result.UnmarshalJSON | v0.34.0 | v0.36.0 | 
Result.Ok | v0.34.0 | v0.36.0 | 
Result.Resource | v0.34.0 | v0.36.0 | 
Result.Message | v0.34.0 | v0.36.0 | 
Result.Error | v0.34.0 | v0.36.0 | 
Result.State | v0.34.0 | v0.36.0 | 
Result.Dump | v0.34.0 | v0.36.0 | 
ResultGroup.Ok | v0.34.0 | v0.36.0 | 
ResultGroup.ErrorResults | v0.34.0 | v0.36.0 | 
ResultGroup.Error | v0.34.0 | v0.36.0 | 
ShareAccess.Validate | v0.34.0 | v0.36.0 | 
Share.Type | v0.34.0 | v0.36.0 | 
Share.Intent | v0.34.0 | v0.36.0 | 
Share.Identity | v0.34.0 | v0.36.0 | 
Share.MarshalJSON | v0.34.0 | v0.36.0 | 
Share.Validate | v0.34.0 | v0.36.0 | 
Share.SetCephFS | v0.34.0 | v0.36.0 | 
NewShare | v0.34.0 | v0.36.0 | 
NewShareToRemove | v0.34.0 | v0.36.0 | 
UsersAndGroups.Type | v0.34.0 | v0.36.0 | 
UsersAndGroups.Intent | v0.34.0 | v0.36.0 | 
UsersAndGroups.Identity | v0.34.0 | v0.36.0 | 
UsersAndGroups.Validate | v0.34.0 | v0.36.0 | 
UsersAndGroups.MarshalJSON | v0.34.0 | v0.36.0 | 
UsersAndGroups.SetValues | v0.34.0 | v0.36.0 | 
NewUsersAndGroups | v0.34.0 | v0.36.0 | 
NewLinkedUsersAndGroups | v0.34.0 | v0.36.0 | 
NewUsersAndGroupsToRemove | v0.34.0 | v0.36.0 | 

## Package: common/admin/osd

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
NewFromConn | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.OSDBlocklist | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.OSDBlocklistAdd | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Admin.OSDBlocklistRemove | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Float.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

