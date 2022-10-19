package nescript

import "os"

type dynamicData struct {
	data map[string]any
	env  []string
}

// Data returns the map of template data to be used when compiling the
// script/cmd.
func (dd dynamicData) Data() map[string]any {
	return dd.data
}

// Env returns the env vars in KEY=VALUE format that will be used when executing
// the script/cmd.
func (dd dynamicData) Env() []string {
	return dd.env
}

func (dd *dynamicData) addField(key string, value any) {
	if dd.data == nil {
		dd.data = make(map[string]any)
	}
	dd.data[key] = value
}

func (dd *dynamicData) addFields(fields map[string]any, overwrite bool) {
	if dd.data == nil {
		dd.data = make(map[string]any)
	}
	for k, v := range fields {
		if _, ok := dd.data[k]; !ok || overwrite {
			dd.data[k] = v
		}
	}
}

func (dd *dynamicData) addEnv(env ...string) {
	if dd.env == nil {
		dd.env = make([]string, 0)
	}
	dd.env = append(dd.env, env...)
}

func (dd *dynamicData) addLocalOSEnv() {
	dd.addEnv(os.Environ()...)
}
