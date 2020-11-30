// +build !luminous,!mimic

package admin

type commonRmFlags struct {
	force bool
}

func (f commonRmFlags) Update(m map[string]string) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		o[k] = v
	}
	if f.force {
		o["force"] = true
	}
	return o
}
