<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->

# go-ceph API Stability

## Package: cephfs

## Package: cephfs/admin

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
CloneStatus.GetFailure | v0.16.0 | v0.18.0 | 

### Deprecated APIs

Name | Deprecated in Version | Expected Removal Version | 
---- | --------------------- | ------------------------ | 

## Package: rados

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
IOContext.SetLocator | v0.15.0 | v0.17.0 | 
IOContext.SetAllocationHint | v0.17.0 | v0.19.0 | 
WriteOp.SetAllocationHint | v0.17.0 | v0.19.0 | 
IOContext.Alignment | v0.17.0 | v0.19.0 | 
IOContext.RequiresAlignment | v0.17.0 | v0.19.0 | 

## Package: rbd

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
Snapshot.Rename | v0.16.0 | v0.18.0 | 

### Deprecated APIs

Name | Deprecated in Version | Expected Removal Version | 
---- | --------------------- | ------------------------ | 
MirrorImageGlobalStatusIter.Close | v0.11.0 |  | 
Image.Open | v0.2.0 |  | 
Snapshot.Set | v0.10.0 |  | 

## Package: rbd/admin

## Package: rgw/admin

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
API.UnlinkBucket | v0.15.0 | v0.17.0 | 
API.LinkBucket | v0.15.0 | v0.17.0 | 
API.CreateSubuser | v0.15.0 | v0.17.0 | 
API.RemoveSubuser | v0.15.0 | v0.17.0 | 
API.ModifySubuser | v0.15.0 | v0.17.0 | 

## Package: common/admin/manager

## Package: common/log

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
SetWarnf | v0.15.0 | v0.17.0 | 
SetDebugf | v0.15.0 | v0.17.0 | 

## Package: common/admin/nfs

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
NewFromConn | v0.16.0 | v0.18.0 | 
Admin.CreateCephFSExport | v0.16.0 | v0.18.0 | 
Admin.RemoveExport | v0.16.0 | v0.18.0 | 
Admin.ListDetailedExports | v0.16.0 | v0.18.0 | 
Admin.ExportInfo | v0.16.0 | v0.18.0 | 

