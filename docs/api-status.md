<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->

# go-ceph API Stability

## Package: cephfs

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
FileBlockDiffInit | v0.38.0 | v0.40.0 | 
FileBlockDiffInfo.Close | v0.38.0 | v0.40.0 | 
FileBlockDiffInfo.More | v0.38.0 | v0.40.0 | 
FileBlockDiffInfo.Read | v0.38.0 | v0.40.0 | 

## Package: cephfs/admin

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rados

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: rbd

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
API.CreateAccount | v0.38.0 | v0.40.0 | 
API.GetAccount | v0.38.0 | v0.40.0 | 
API.DeleteAccount | v0.38.0 | v0.40.0 | 
API.ModifyAccount | v0.38.0 | v0.40.0 | 

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
TopIdentityKind.Identity | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ChildIdentityKind.Identity | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
GuessIdentityKind | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
GenericResource.Type | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
GenericResource.Intent | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
GenericResource.Identity | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
GenericResource.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
GenericResource.UnmarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
GenericResource.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
GenericResource.Convert | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ToGeneric | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ShowOptions.SetGeneric | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
ShowOptions.Generic | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

## Package: common/admin/osd

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/admin/nvmegw

No Preview/Deprecated APIs found. All APIs are considered stable.

## Package: common/commands/builder

### Preview APIs

Name | Added in Version | Expected Stable Version | 
---- | ---------------- | ----------------------- | 
NewBuilder | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Builder.Prepare | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Builder.Arguments | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Builder.ArgumentsMap | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Builder.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Builder.MarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Builder.Apply | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
BindArgumentType | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephChoices.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephChoices.Name | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephChoices.Choices | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephChoices.Convert | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephChoices.Check | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephChoices.Set | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephChoices.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephString.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephString.Name | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephString.Convert | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephString.Check | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephString.Set | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephString.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephInt.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephInt.Name | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephInt.Convert | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephInt.Check | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephInt.Set | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephInt.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephFloat.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephFloat.Name | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephFloat.Convert | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephFloat.Check | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephFloat.Set | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephFloat.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephBool.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephBool.Name | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephBool.Convert | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephBool.Check | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephBool.Set | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephBool.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephPoolName.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephObjectName.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephOSDName.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephPGID.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephUnknownType.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephUnknownType.Name | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephUnknownType.Set | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephUnknownType.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephRepeatedArg.TypeName | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephRepeatedArg.Name | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephRepeatedArg.Set | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephRepeatedArg.Append | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CephRepeatedArg.Validate | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
SignatureVar.Required | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
SignatureElement.UnmarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Description.Prefix | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Description.PrefixString | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
Description.Variables | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CommandDescriptions.UnmarshalJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CommandDescriptions.Match | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
CommandDescriptions.Find | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
QueryMgrJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
QueryMonJSON | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
QueryMgrDescriptions | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 
QueryMonDescriptions | $NEXT_RELEASE | $NEXT_RELEASE_STABLE | 

