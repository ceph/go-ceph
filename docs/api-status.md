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
OpenSnapDiff | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
SnapDiffInfo.Readdir | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
SnapDiffInfo.Close | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

## Package: cephfs/admin

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rados

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rbd

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
Image.GetDataPoolID | v0.36.0 | v0.38.0 | 

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
NewBindAddress | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewNetworkBindAddress | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
BindAddress.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
BindAddress.UnmarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
BindAddress.Address | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
BindAddress.Network | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
BindAddress.IsNetwork | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
BindAddress.String | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Cluster.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
RemoteControl.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
TLSCredential.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
TLSCredential.Intent | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
TLSCredential.SetIntent | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
TLSCredential.Identity | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
TLSCredential.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
TLSCredential.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
TLSCredential.Set | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewTLSCredential | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewLinkedTLSCredential | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
NewTLSCredentialToRemove | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

## Package: common/admin/osd

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
NewFromConn | v0.36.0 | v0.38.0 | 
Admin.OSDBlocklist | v0.36.0 | v0.38.0 | 
Admin.OSDBlocklistAdd | v0.36.0 | v0.38.0 | 
Admin.OSDBlocklistRemove | v0.36.0 | v0.38.0 | 

## Package: common/admin/nvmegw

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
NewFromConn | v0.36.0 | v0.38.0 | 
Admin.CreateGateway | v0.36.0 | v0.38.0 | 
Admin.DeleteGateway | v0.36.0 | v0.38.0 | 
Admin.ShowGateways | v0.36.0 | v0.38.0 | 

