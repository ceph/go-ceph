package rados

import (
    "github.com/noahdesu/go-rados/rados"
    "github.com/noahdesu/go-rados/rbd"
    "testing"
)

func TestImports(t *testing.T) {
    if assert.Equal(t, 1, 1) != true {
        t.Error("Something is wrong.")
    }
}
