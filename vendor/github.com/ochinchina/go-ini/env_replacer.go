package ini

import (
	"bytes"
	"os"
	"strings"
)

func get_env_value(env string) (string, bool) {
	pos := strings.Index(env, ":")
	if pos == -1 {
		return os.LookupEnv(env)
	}

	real_env := env[0:pos]
	def_value := env[pos+1:]
	if len(def_value) > 0 && def_value[0] == '-' {
		def_value = def_value[1:]
	}
	if value, ok := os.LookupEnv(real_env); ok {
		return value, ok
	} else {
		return def_value, true
	}
}

func replace_env(s string) string {
	n := len(s)
	env_start_pos := -1
	result := bytes.NewBuffer(make([]byte, 0))

	for i := 0; i < n; i++ {
		//if env start flag "${" is found but env end flag "}" is not found
		if env_start_pos >= 0 && s[i] != '}' {
			continue
		}
		switch s[i] {
		case '\\':
			result.WriteByte(s[i])
			if i+1 < n {
				i++
				result.WriteByte(s[i])
			}
		case '$':
			if i+1 < n && s[i+1] == '{' {
				env_start_pos = i
				i++
			} else {
				result.WriteByte(s[i])
			}
		case '}':
			if env_start_pos >= 0 {
				if env_value, ok := get_env_value(s[env_start_pos+2 : i]); ok {
					result.WriteString(env_value)
				}
				env_start_pos = -1
			} else {
				result.WriteByte(s[i])
			}
		default:
			result.WriteByte(s[i])
		}
	}
	return result.String()
}
