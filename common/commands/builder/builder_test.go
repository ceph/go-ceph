//go:build !(pacific || quincy) && ceph_preview

package builder

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilderArgs1(t *testing.T) {
	cde := CommandDescriptions{}
	b := []byte(sample1)
	assert.NoError(t, json.Unmarshal(b, &cde))
	assert.Len(t, cde.Entries, 37)

	matches := cde.Find("osd", "reweight-by-utilization")
	assert.Len(t, matches, 1)

	bld := NewBuilder(matches[0])
	argTypes := bld.Arguments()
	assert.Len(t, argTypes, 4)

	t.Run("arg0", func(t *testing.T) {
		cat := argTypes[0]
		assert.EqualValues(t, "CephInt", cat.TypeName())
		assert.NoError(t, cat.Set(bld.Values, 44))
		assert.NoError(t, cat.Set(bld.Values, "11"))
		assert.Error(t, cat.Set(bld.Values, nil))
		assert.NoError(t, cat.Validate(bld.Values))
	})

	t.Run("arg1", func(t *testing.T) {
		cat := argTypes[1]
		assert.EqualValues(t, "CephFloat", cat.TypeName())
		assert.NoError(t, cat.Set(bld.Values, 4.4))
		assert.NoError(t, cat.Set(bld.Values, "11"))
		assert.Error(t, cat.Set(bld.Values, nil))
		assert.NoError(t, cat.Validate(bld.Values))
	})

	t.Run("arg2", func(t *testing.T) {
		cat := argTypes[2]
		assert.EqualValues(t, "CephInt", cat.TypeName())
		assert.NoError(t, cat.Set(bld.Values, 44))
		assert.NoError(t, cat.Set(bld.Values, "11"))
		assert.Error(t, cat.Set(bld.Values, nil))
		assert.NoError(t, cat.Validate(bld.Values))
	})

	t.Run("arg3", func(t *testing.T) {
		cat := argTypes[3]
		assert.EqualValues(t, "CephBool", cat.TypeName())
		assert.NoError(t, cat.Set(bld.Values, true))
		assert.NoError(t, cat.Set(bld.Values, "false"))
		assert.Error(t, cat.Set(bld.Values, nil))
		assert.NoError(t, cat.Validate(bld.Values))

		bld.Values["no_increasing"] = nil
		assert.Error(t, cat.Validate(bld.Values))
	})
}

func TestBuilderCommands1(t *testing.T) {
	cde := CommandDescriptions{}
	b := []byte(sample1)
	assert.NoError(t, json.Unmarshal(b, &cde))
	assert.Len(t, cde.Entries, 37)

	t.Run("osd-pool-scrub", func(t *testing.T) {
		matches := cde.Find("osd", "pool", "scrub")
		assert.Len(t, matches, 1)

		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 1)

		cat := argTypes[0]
		assert.EqualValues(t, "CephPoolname (Repeat: N)", cat.TypeName())
		assert.NoError(t, cat.Set(bld.Values, []string{"foo", "bar"}))
		assert.NoError(t, cat.Set(bld.Values, "bloop"))
		assert.EqualValues(t, []any{"foo", "bar", "bloop"}, bld.Values[cat.Name()])
	})

	t.Run("pg-dump_json", func(t *testing.T) {
		matches := cde.Find("pg", "dump_json")
		assert.Len(t, matches, 1)

		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 1)

		cat := argTypes[0]
		assert.EqualValues(t, "CephChoices (Repeat: N)", cat.TypeName())
		assert.Error(t, cat.Set(bld.Values, "bronco"))
		assert.Error(t, cat.Set(bld.Values, "splatter"))
		assert.NoError(t, cat.Set(bld.Values, "all"))
		assert.NoError(t, cat.Set(bld.Values, "summary"))
		assert.EqualValues(t, []any{"all", "summary"}, bld.Values[cat.Name()])
		assert.NoError(t, cat.Validate(bld.Values))
	})
}

func TestBuilderWrapsAll(t *testing.T) {
	cde := CommandDescriptions{}
	b := []byte(sample1)
	assert.NoError(t, json.Unmarshal(b, &cde))
	assert.Len(t, cde.Entries, 37)

	for _, d := range cde.Entries {
		bld := NewBuilder(d)
		for _, a := range bld.Arguments() {
			assert.NotContains(t, a.TypeName(), "Unknown")
			assert.NotEmpty(t, a.Name())
		}
	}
}

func TestBuilderApply1(t *testing.T) {
	cde := CommandDescriptions{}
	b := []byte(sample1)
	assert.NoError(t, json.Unmarshal(b, &cde))
	assert.Len(t, cde.Entries, 37)

	matches := cde.Find("osd", "df")
	assert.Len(t, matches, 1)

	t.Run("applyArgs", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 3)

		assert.NoError(t, bld.Apply([]string{"tree", "name", "plume"}, nil))
		assert.Len(t, bld.Values, 4)
		assert.EqualValues(t, "osd df", bld.Values["prefix"])
		assert.EqualValues(t, "tree", bld.Values["output_method"])
		assert.EqualValues(t, "name", bld.Values["filter_by"])
		assert.EqualValues(t, "plume", bld.Values["filter"])
		assert.NoError(t, bld.Validate())
	})

	t.Run("applyArgsAndMap", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 3)

		assert.NoError(t, bld.Apply(
			[]string{"tree"},
			map[string]string{"filter_by": "class", "filter": "ghi"}))
		assert.Len(t, bld.Values, 4)
		assert.EqualValues(t, "osd df", bld.Values["prefix"])
		assert.EqualValues(t, "tree", bld.Values["output_method"])
		assert.EqualValues(t, "class", bld.Values["filter_by"])
		assert.EqualValues(t, "ghi", bld.Values["filter"])
		assert.NoError(t, bld.Validate())
	})

	t.Run("applyArgsAndMapError", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 3)

		assert.Error(t, bld.Apply(
			[]string{"tree"},
			map[string]string{"bilter_fy": "class", "filter": "ghi"}))
		assert.Error(t, bld.Apply(
			[]string{"tree"},
			map[string]string{"filter_by": "glass", "filter": "ghi"}))
	})
}

func TestBuilderApply2(t *testing.T) {
	cde := CommandDescriptions{}
	b := []byte(sample1)
	assert.NoError(t, json.Unmarshal(b, &cde))
	assert.Len(t, cde.Entries, 37)

	matches := cde.Find("osd", "pool", "repair")
	assert.Len(t, matches, 1)
	t.Run("missingArgs", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 1)

		// Apply itself doesn't know if things are missing
		assert.NoError(t, bld.Apply([]string{}, nil))
		// Validate will produce errors when required args are not present
		assert.Error(t, bld.Validate())
	})
}

func TestBuilderApply3(t *testing.T) {
	cde := CommandDescriptions{}
	b := []byte(sample1)
	assert.NoError(t, json.Unmarshal(b, &cde))
	assert.Len(t, cde.Entries, 37)

	matches := cde.Find("pg", "dump")
	assert.Len(t, matches, 1)
	t.Run("manyArgs", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 1)

		// all|summary|sum|delta|pools|osds|pgs|pgs_brief
		assert.NoError(t, bld.Apply(
			[]string{"pgs", "all", "pools", "osds", "sum"}, nil))
		assert.NoError(t, bld.Validate())
	})
	t.Run("manyArgsWithError", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 1)

		// all|summary|sum|delta|pools|osds|pgs|pgs_brief
		assert.Error(t, bld.Apply(
			[]string{"pgs", "all", "pools", "osds", "sim"}, nil))
	})
	t.Run("earlyError", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 1)

		// all|summary|sum|delta|pools|osds|pgs|pgs_brief
		assert.Error(t, bld.Apply(
			[]string{"pogs", "delta"}, nil))
	})
}

func TestBuilderMarsalJSON(t *testing.T) {
	cde := CommandDescriptions{}
	b := []byte(sample1)
	assert.NoError(t, json.Unmarshal(b, &cde))
	assert.Len(t, cde.Entries, 37)

	matches := cde.Find("pg", "dump")
	assert.Len(t, matches, 1)

	t.Run("ok", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 1)

		assert.NoError(t, bld.Apply([]string{"pgs"}, nil))
		j, err := bld.MarshalJSON()
		assert.NoError(t, err)
		assert.Contains(t, string(j), "pg dump")
		assert.Contains(t, string(j), "pgs")
	})
	t.Run("error", func(t *testing.T) {
		bld := NewBuilder(matches[0])
		argTypes := bld.Arguments()
		assert.Len(t, argTypes, 1)

		bld.Values["junk"] = 0
		bld.Values["dumpcontents"] = 0
		_, err := bld.MarshalJSON()
		assert.Error(t, err)
	})
}

func TestBindArgumentTypeUnknown(t *testing.T) {
	cat := BindArgumentType(&SignatureVar{
		Type: "FooBar",
		Name: "nopetown",
	})
	_, ok := cat.(*CephUnknownType)
	assert.True(t, ok)
}
