package api

import (
	"testing"
)

var client = CephClient {
	BaseUrl: "<Api Endpoint>",
}

func TestCephStatus(t *testing.T) {
	_, err := client.GetStatus()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetOsdFlag(t *testing.T) {
	if err := client.SetOsdFlag("noout"); err != nil {
		t.Fatal(err)
	}
}

func TestUnsetOsdFlag(t *testing.T) {
	if err := client.UnsetOsdFlag("noout"); err != nil {
		t.Fatal(err)
	}
}

func TestGetBlacklist(t *testing.T) {
	blacklists, err := client.GetBlacklist()
	if err != nil {
		t.Fatal(err)
	}
	if len(blacklists.Nodes) != 0 {
		t.Fatalf("Expected 0 nodes to be blacklisted, but %d were", len(blacklists.Nodes))
	}
}

func TestAddBlacklist(t *testing.T) {
	blacklists, err := client.GetBlacklist()
	if err != nil {
		t.Fatal(err)
	}
	numBlacklists := len(blacklists.Nodes)

	if err = client.BlacklistOp("0.0.0.0", "add"); err != nil {
		t.Fatal(err)
	}

	blacklists, err = client.GetBlacklist()
	if err != nil {
		t.Fatal(err)
	}

	if len(blacklists.Nodes) != numBlacklists + 1 {
		t.Fatalf("Expected %d blacklists, but got %d instead", numBlacklists+1, len(blacklists.Nodes))
	}
}

func TestRemoveBlacklist(t *testing.T) {
	if err := client.BlacklistOp("0.0.0.0", "add"); err != nil {
		t.Fatal(err)
	}	
	blacklists, err := client.GetBlacklist()
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range blacklists.Nodes {
		if err := client.BlacklistOp(node.Addr, "rm"); err != nil {
			t.Fatal(err)
		}
	}
	blacklists, err = client.GetBlacklist()
	if err != nil {
		t.Fatal(err)
	}
	if len(blacklists.Nodes) != 0 {
		t.Fatalf("Expected 0 blacklists, but got %d instead", len(blacklists.Nodes))
	}
}

func TestGetMdsStatus(t *testing.T) {
	_, err := client.GetMdsStat()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetOsdTree(t *testing.T) {
	_, err := client.GetOsdTree()
	if err != nil {
		t.Fatal(err)
	}
}
