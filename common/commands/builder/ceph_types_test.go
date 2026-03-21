//go:build !(pacific || quincy) && ceph_preview

package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type St struct {
	v string
}

func (s St) String() string {
	return s.v
}

type XX struct {
	junk string
}

func TestCephChoices(t *testing.T) {
	c := &CephChoices{
		&SignatureVar{
			Name:    "flavor",
			Type:    CephTypeChoices,
			Choices: "chocolate|vanilla|cherry",
		},
	}

	t.Run("Choices", func(t *testing.T) {
		assert.EqualValues(t,
			map[string]bool{"chocolate": true, "vanilla": true, "cherry": true},
			c.Choices())
	})

	t.Run("Convert", func(t *testing.T) {
		var err error
		_, err = c.Convert("cherry")
		assert.NoError(t, err)
		_, err = c.Convert("chocolate")
		assert.NoError(t, err)
		_, err = c.Convert(St{"vanilla"})
		assert.NoError(t, err)

		_, err = c.Convert("dandelion")
		assert.Error(t, err)
		_, err = c.Convert(nil)
		assert.Error(t, err)
	})

	t.Run("Check", func(t *testing.T) {
		assert.NoError(t, c.Check("cherry"))
		assert.Error(t, c.Check(nil))
		assert.Error(t, c.Check("pineapple"))
	})

	t.Run("Set", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Set(m, "pineapple"))
		assert.Len(t, m, 0)
		assert.NoError(t, c.Set(m, "cherry"))
		assert.Len(t, m, 1)
		assert.Contains(t, m, "flavor")
	})

	t.Run("Validate", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Validate(m))
		m["flavor"] = "watermelon"
		assert.Error(t, c.Validate(m))
		m["flavor"] = "vanilla"
		assert.NoError(t, c.Validate(m))
	})
}

func TestCephString(t *testing.T) {
	c := &CephString{
		&SignatureVar{
			Name: "city",
			Type: CephTypeString,
		},
	}

	t.Run("Convert", func(t *testing.T) {
		var err error
		_, err = c.Convert("london")
		assert.NoError(t, err)
		_, err = c.Convert("paris")
		assert.NoError(t, err)
		_, err = c.Convert(St{"paradise"})
		assert.NoError(t, err)

		_, err = c.Convert(nil)
		assert.Error(t, err)
	})

	t.Run("Check", func(t *testing.T) {
		assert.NoError(t, c.Check("emerald"))
		assert.NoError(t, c.Check("")) // empty string still a string
		assert.Error(t, c.Check(nil))
	})

	t.Run("Set", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Set(m, XX{}))
		assert.Len(t, m, 0)
		assert.NoError(t, c.Set(m, "fitchburg"))
		assert.Len(t, m, 1)
		assert.Contains(t, m, "city")
	})

	t.Run("Validate", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Validate(m))
		m["city"] = "metropolis"
		assert.NoError(t, c.Validate(m))
	})
}

func TestCephInt(t *testing.T) {
	c := &CephInt{
		&SignatureVar{
			Name: "weight",
			Type: CephTypeInt,
		},
	}

	t.Run("Convert", func(t *testing.T) {
		var err error
		_, err = c.Convert(5)
		assert.NoError(t, err)
		_, err = c.Convert(int64(1001))
		assert.NoError(t, err)
		_, err = c.Convert(uint8(12))
		assert.NoError(t, err)
		_, err = c.Convert("2112")
		assert.NoError(t, err)

		_, err = c.Convert(nil)
		assert.Error(t, err)
	})

	t.Run("Check", func(t *testing.T) {
		assert.NoError(t, c.Check(55))
		assert.NoError(t, c.Check(uint16(202)))
		assert.Error(t, c.Check("frodo"))
	})

	t.Run("Set", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Set(m, XX{}))
		assert.Len(t, m, 0)
		assert.NoError(t, c.Set(m, "876"))
		assert.Len(t, m, 1)
		assert.Contains(t, m, "weight")
	})

	t.Run("Validate", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Validate(m))
		m["weight"] = 102
		assert.NoError(t, c.Validate(m))
	})
}

func TestCephFloat(t *testing.T) {
	c := &CephFloat{
		&SignatureVar{
			Name: "width",
			Type: CephTypeFloat,
		},
	}

	t.Run("Convert", func(t *testing.T) {
		var err error
		_, err = c.Convert(5.5)
		assert.NoError(t, err)
		_, err = c.Convert(float64(10.77))
		assert.NoError(t, err)
		_, err = c.Convert(float32(77.01))
		assert.NoError(t, err)
		_, err = c.Convert("21.12")
		assert.NoError(t, err)

		_, err = c.Convert(nil)
		assert.Error(t, err)
	})

	t.Run("Check", func(t *testing.T) {
		assert.NoError(t, c.Check(5.5))
		assert.NoError(t, c.Check(float32(2.0002)))
		assert.Error(t, c.Check("frodo"))
	})

	t.Run("Set", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Set(m, XX{}))
		assert.Len(t, m, 0)
		assert.NoError(t, c.Set(m, "87.061"))
		assert.Len(t, m, 1)
		assert.Contains(t, m, "width")
	})

	t.Run("Validate", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Validate(m))
		m["width"] = 10.2
		assert.NoError(t, c.Validate(m))
	})
}

func TestCephUnknownType(t *testing.T) {
	c := &CephUnknownType{
		&SignatureVar{
			Name: "fooblat",
			Type: "CephFooBlat",
		},
	}

	assert.Contains(t, c.TypeName(), "Unknown")
	assert.EqualValues(t, c.Name(), "fooblat")

	m := map[string]any{}
	assert.Error(t, c.Set(m, "foo"))
	assert.NoError(t, c.Validate(m))
}

func TestCephRepeatedArg(t *testing.T) {
	sv := &SignatureVar{
		Name:    "flavor",
		Type:    CephTypeChoices,
		Choices: "chocolate|vanilla|cherry",
	}
	c := &CephRepeatedArg{&CephChoices{sv}, sv}

	assert.Contains(t, c.TypeName(), "Repeat")
	assert.EqualValues(t, c.Name(), "flavor")

	t.Run("SetOne", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Set(m, "melon"))
		assert.Len(t, m, 0)
		assert.NoError(t, c.Set(m, "vanilla"))
		assert.Len(t, m, 1)
		assert.Contains(t, m, "flavor")
	})

	t.Run("Append", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Append(m, "melon"))
		assert.Len(t, m, 0)
		assert.NoError(t, c.Append(m, "vanilla"))
		assert.NoError(t, c.Append(m, "chocolate"))
		assert.Len(t, m, 1)
		assert.Contains(t, m, "flavor")
		assert.Len(t, m["flavor"], 2)
	})

	t.Run("SetSlice", func(t *testing.T) {
		m := map[string]any{}
		assert.Error(t, c.Set(m, []any{"bob", "lucy"}))
		assert.Len(t, m, 0)
		assert.Error(t, c.Set(m, []any{"vanilla", "lucy"}))
		assert.Len(t, m, 0)
		assert.NoError(t, c.Set(m, []any{"vanilla", "cherry"}))
		assert.Len(t, m, 1)
		assert.Contains(t, m, "flavor")
		assert.Len(t, m["flavor"], 2)
	})

	t.Run("Validate", func(t *testing.T) {
		m := map[string]any{}
		assert.ErrorContains(t, c.Validate(m), "missing")
		m["flavor"] = 32
		assert.ErrorContains(t, c.Validate(m), "slice")
		m["flavor"] = []any{"vanilla", XX{}}
		assert.ErrorContains(t, c.Validate(m), "string")
		m["flavor"] = []any{"vanilla", "cherry"}
		assert.NoError(t, c.Validate(m))
	})

	t.Run("ValidateNotReq", func(t *testing.T) {
		req := false
		sv.Req = &req
		m := map[string]any{}
		assert.NoError(t, c.Validate(m))
		m["flavor"] = 32
		assert.ErrorContains(t, c.Validate(m), "slice")
		m["flavor"] = []any{"vanilla", "cherry"}
		assert.NoError(t, c.Validate(m))
	})
}

func TestValidateNotReq(t *testing.T) {
	req := false
	c := &CephInt{
		&SignatureVar{
			Name: "weight",
			Type: CephTypeInt,
			Req:  &req,
		},
	}

	m := map[string]any{}
	assert.NoError(t, c.Validate(m))
}
