// +build !luminous,!mimic

package admin

type rmFlags struct {
	force bool
}

func (f rmFlags) Update(m map[string]string) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		o[k] = v
	}
	if f.force {
		o["force"] = true
	}
	return o
}
