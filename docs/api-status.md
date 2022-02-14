<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->

# go-ceph API Stability

## Package: cephfs

## Package: cephfs/admin

### Deprecated APIs

Name | Deprecated in Version | Expected Removal Version | 
---- | --------------------- | ------------------------ | 
FSAdmin.EnableModule | v0.14.0 | v0.16.0 | 
FSAdmin.DisableModule | v0.14.0 | v0.16.0 | 

## Package: rados

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
ReadOp.Read | v0.14.0 | v0.16.0 | 
WriteOp.Remove | v0.14.0 | v0.16.0 | 
ReadOp.AssertVersion | v0.14.0 | v0.16.0 | 
WriteOp.AssertVersion | v0.14.0 | v0.16.0 | 
WriteOp.SetXattr | v0.14.0 | v0.16.0 | 
ReadOpOmapGetValsByKeysStep.Next | v0.14.0 | v0.16.0 | 
ReadOp.GetOmapValuesByKeys | v0.14.0 | v0.16.0 | 
IOContext.Watch | v0.14.0 | v0.16.0 | 
IOContext.WatchWithTimeout | v0.14.0 | v0.16.0 | 
Watcher.ID | v0.14.0 | v0.16.0 | 
Watcher.Events | v0.14.0 | v0.16.0 | 
Watcher.Errors | v0.14.0 | v0.16.0 | 
Watcher.Check | v0.14.0 | v0.16.0 | 
Watcher.Delete | v0.14.0 | v0.16.0 | 
IOContext.Notify | v0.14.0 | v0.16.0 | 
IOContext.NotifyWithTimeout | v0.14.0 | v0.16.0 | 
NotifyEvent.Ack | v0.14.0 | v0.16.0 | 
Conn.WatcherFlush | v0.14.0 | v0.16.0 | 

## Package: rbd

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

## Package: common/admin/manager

