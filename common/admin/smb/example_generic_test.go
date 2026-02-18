//go:build !(pacific || quincy || reef || squid) && ceph_preview

package smb_test

import (
	"fmt"

	"github.com/ceph/go-ceph/common/admin/smb"
	"github.com/ceph/go-ceph/rados"
)

func ExampleGenericResource() {
	conn, err := rados.NewConn()
	if err != nil {
		panic("rados.NewConn failed")
	}
	sadmin := smb.NewFromConn(conn)

	// fetch share resources from the ceph smb mgr module
	res, err := sadmin.Show([]smb.ResourceRef{smb.ShareType}, nil)
	if err != nil {
		panic("sadmin.Show failed")
	}

	// the default show behavior is to fetch resources based on
	// concrete Structs
	if res[0].Type() != smb.ShareType {
		panic("incorrect type")
	}
	share := res[0].(*smb.Share)
	fmt.Printf("Cluster ID: %s\n", share.ClusterID)
	fmt.Printf("Share ID: %s\n", share.ShareID)
	fmt.Printf("Volume: %s\n", share.CephFS.Volume)

	// fetch generic share resources
	opts := &smb.ShowOptions{}
	opts.SetGeneric(true)
	res, err = sadmin.Show([]smb.ResourceRef{smb.ShareType}, opts)
	if err != nil {
		panic("sadmin.Show failed")
	}

	// now the resources fetched with be Generic Resources that
	// can represent fields and values unknown to the structs but
	// also have all the same data
	if res[0].Type() != smb.ShareType {
		panic("incorrect type")
	}
	gr := res[0].(*smb.GenericResource)
	fmt.Printf("Cluster ID: %s\n", gr.Values["cluster_id"].(string))
	fmt.Printf("Share ID: %s\n", gr.Values["share_id"].(string))
	fmt.Printf("Volume: %s\n",
		gr.Values["cephfs"].(map[string]any)["subvolume"].(string))
	// we can look for fields that the Share struct doesn't know about
	if v, ok := gr.Values["cephfs"].(map[string]any)["flux_capacity"]; ok {
		fmt.Printf("Flux Capacity: %v\n", v.(int))
	}

	// for convenience one can convert between structs for Share/Cluster/etc
	// and Generic Resources.
	// The Convert method of a generic resource returns a struct-based concrete
	// resource type if possible.
	r2, err := gr.Convert()
	if res[0].Type() != smb.ShareType {
		panic("invalid resource type")
	}
	share2 := r2.(*smb.Share)
	fmt.Printf("Cluster ID: %s\n", share2.ClusterID)
	fmt.Printf("Share ID: %s\n", share2.ShareID)
	fmt.Printf("Volume: %s\n", share2.CephFS.Volume)
	// N.B. converting from a generic resource can lose information if
	// the resource has fields that are not known to the struct like
	// our hypothetical "flux_capacity" field above.

	// It is possible, and probably more convenient to start with an empty
	// struct based resource and then make a generic resource out of it
	// in order to pass extra fields to the mgr module.
	share3 := smb.NewShare("cluster1", "example1")
	share3.Name = "Example Share 1"
	share3.SetCephFS("cephfs", "smb", "sv1", "/")
	share3.LoginControl = []smb.ShareAccess{
		{
			Name:     "domain1\\bwayne",
			Category: smb.UserAccess,
			Access:   smb.ReadWriteAccess,
		},
		{
			Name:     "domain1\\ckent",
			Category: smb.UserAccess,
			Access:   smb.ReadWriteAccess,
		},
	}
	// Use the ToGeneric method to create a generic resource from a struct
	// based resource of the same type.
	gr2, err := smb.ToGeneric(share3)
	if res[0].Type() != smb.ShareType {
		panic("ToGeneric failed")
	}
	// Internally the values will be the same
	fmt.Printf("Cluster ID: %s\n", gr.Values["cluster_id"].(string))
	fmt.Printf("Share ID: %s\n", gr.Values["share_id"].(string))
	// But now you can extend the resource with fields and data
	// that do not exist in the struct-based forms.
	gr2.Values["cephfs"].(map[string]any)["flux_capacity"] = 88

	// Because both struct-based and generic resources both meet the Resource
	// interface, either type may be used to update the mgr module via the
	// apply method.
	_, err = sadmin.Apply([]smb.Resource{share2, gr2}, nil)
	if err != nil {
		panic("sadmin.Apply failed")
	}
}
