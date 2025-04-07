/*
Package smb from common/admin contains a set of APIs used to interact
with and administer SMB support for ceph clusters.

The Ceph smb mgr module is based on the concept of resources. Resource
descriptions are used to create, update, or delete configuration state in
the Ceph cluster and the Ceph cluster will attempt to configure SMB Servers
based on these resources.

Resource types include Cluster, JoinAuth, UsersAndGroups, and Share. To
modify the state on the Ceph cluster use the Apply function. To query
the state of the resources on the Ceph cluster use the Show function.
Resources that are to be deleted should have an Intent value of Removed.
Resources being updated or created must have an Intent value of Present.
*/
package smb
