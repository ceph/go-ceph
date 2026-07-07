<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->

# go-ceph API Stability

## Package: cephfs

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: cephfs/admin

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
FSAdmin.SubVolumeGroupInfo | v0.40.0 | v0.42.0 | 

## Package: rados

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
IOContext.Checksum | v0.40.0 | v0.42.0 | 

## Package: rbd

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
Image.FlattenWithProgress | v0.40.0 | v0.42.0 | 

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
API.SetIndividualBucketRateLimit | v0.40.0 | v0.42.0 | 
API.GetIndividualBucketRateLimit | v0.40.0 | v0.42.0 | 
API.SetAccountQuota | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

## Package: common/admin/manager

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/log

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/admin/nfs

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
Admin.ApplyExportInfo | v0.40.0 | v0.42.0 | 

## Package: rados/striper

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/admin/smb

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
TopIdentityKind.Identity | v0.39.0 | v0.41.0 | 
ChildIdentityKind.Identity | v0.39.0 | v0.41.0 | 
GuessIdentityKind | v0.39.0 | v0.41.0 | 
GenericResource.Type | v0.39.0 | v0.41.0 | 
GenericResource.Intent | v0.39.0 | v0.41.0 | 
GenericResource.Identity | v0.39.0 | v0.41.0 | 
GenericResource.MarshalJSON | v0.39.0 | v0.41.0 | 
GenericResource.UnmarshalJSON | v0.39.0 | v0.41.0 | 
GenericResource.Validate | v0.39.0 | v0.41.0 | 
GenericResource.Convert | v0.39.0 | v0.41.0 | 
ToGeneric | v0.39.0 | v0.41.0 | 
ShowOptions.SetGeneric | v0.39.0 | v0.41.0 | 
ShowOptions.Generic | v0.39.0 | v0.41.0 | 
ApplyOptions.SetGeneric | v0.40.0 | v0.42.0 | 
ApplyOptions.Generic | v0.40.0 | v0.42.0 | 

## Package: common/admin/osd

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/admin/nvmegw

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/commands/builder

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
NewBuilder | v0.39.0 | v0.41.0 | 
Builder.Prepare | v0.39.0 | v0.41.0 | 
Builder.Arguments | v0.39.0 | v0.41.0 | 
Builder.ArgumentsMap | v0.39.0 | v0.41.0 | 
Builder.Validate | v0.39.0 | v0.41.0 | 
Builder.MarshalJSON | v0.39.0 | v0.41.0 | 
Builder.Apply | v0.39.0 | v0.41.0 | 
BindArgumentType | v0.39.0 | v0.41.0 | 
CephChoices.TypeName | v0.39.0 | v0.41.0 | 
CephChoices.Name | v0.39.0 | v0.41.0 | 
CephChoices.Choices | v0.39.0 | v0.41.0 | 
CephChoices.Convert | v0.39.0 | v0.41.0 | 
CephChoices.Check | v0.39.0 | v0.41.0 | 
CephChoices.Set | v0.39.0 | v0.41.0 | 
CephChoices.Validate | v0.39.0 | v0.41.0 | 
CephString.TypeName | v0.39.0 | v0.41.0 | 
CephString.Name | v0.39.0 | v0.41.0 | 
CephString.Convert | v0.39.0 | v0.41.0 | 
CephString.Check | v0.39.0 | v0.41.0 | 
CephString.Set | v0.39.0 | v0.41.0 | 
CephString.Validate | v0.39.0 | v0.41.0 | 
CephInt.TypeName | v0.39.0 | v0.41.0 | 
CephInt.Name | v0.39.0 | v0.41.0 | 
CephInt.Convert | v0.39.0 | v0.41.0 | 
CephInt.Check | v0.39.0 | v0.41.0 | 
CephInt.Set | v0.39.0 | v0.41.0 | 
CephInt.Validate | v0.39.0 | v0.41.0 | 
CephFloat.TypeName | v0.39.0 | v0.41.0 | 
CephFloat.Name | v0.39.0 | v0.41.0 | 
CephFloat.Convert | v0.39.0 | v0.41.0 | 
CephFloat.Check | v0.39.0 | v0.41.0 | 
CephFloat.Set | v0.39.0 | v0.41.0 | 
CephFloat.Validate | v0.39.0 | v0.41.0 | 
CephBool.TypeName | v0.39.0 | v0.41.0 | 
CephBool.Name | v0.39.0 | v0.41.0 | 
CephBool.Convert | v0.39.0 | v0.41.0 | 
CephBool.Check | v0.39.0 | v0.41.0 | 
CephBool.Set | v0.39.0 | v0.41.0 | 
CephBool.Validate | v0.39.0 | v0.41.0 | 
CephPoolName.TypeName | v0.39.0 | v0.41.0 | 
CephObjectName.TypeName | v0.39.0 | v0.41.0 | 
CephOSDName.TypeName | v0.39.0 | v0.41.0 | 
CephPGID.TypeName | v0.39.0 | v0.41.0 | 
CephUnknownType.TypeName | v0.39.0 | v0.41.0 | 
CephUnknownType.Name | v0.39.0 | v0.41.0 | 
CephUnknownType.Set | v0.39.0 | v0.41.0 | 
CephUnknownType.Validate | v0.39.0 | v0.41.0 | 
CephRepeatedArg.TypeName | v0.39.0 | v0.41.0 | 
CephRepeatedArg.Name | v0.39.0 | v0.41.0 | 
CephRepeatedArg.Set | v0.39.0 | v0.41.0 | 
CephRepeatedArg.Append | v0.39.0 | v0.41.0 | 
CephRepeatedArg.Validate | v0.39.0 | v0.41.0 | 
SignatureVar.Required | v0.39.0 | v0.41.0 | 
SignatureElement.UnmarshalJSON | v0.39.0 | v0.41.0 | 
Description.Prefix | v0.39.0 | v0.41.0 | 
Description.PrefixString | v0.39.0 | v0.41.0 | 
Description.Variables | v0.39.0 | v0.41.0 | 
CommandDescriptions.UnmarshalJSON | v0.39.0 | v0.41.0 | 
CommandDescriptions.Match | v0.39.0 | v0.41.0 | 
CommandDescriptions.Find | v0.39.0 | v0.41.0 | 
QueryMgrJSON | v0.39.0 | v0.41.0 | 
QueryMonJSON | v0.39.0 | v0.41.0 | 
QueryMgrDescriptions | v0.39.0 | v0.41.0 | 
QueryMonDescriptions | v0.39.0 | v0.41.0 | 

