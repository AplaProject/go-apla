package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type StringExpression struct {
	env map[string]string
}

func NewStringExpression(envs ...string) *StringExpression {
	se := &StringExpression{env: make(map[string]string)}

	for _, env := range os.Environ() {
		t := strings.Split(env, "=")
		se.env["ENV_"+t[0]] = t[1]
	}
	n := len(envs)
	for i := 0; i+1 < n; i += 2 {
		se.env[envs[i]] = envs[i+1]
	}

	hostname, err := os.Hostname()
	if err == nil {
		se.env["host_node_name"] = hostname
	}

	return se

}

func (se *StringExpression) Add(key string, value string) *StringExpression {
	se.env[key] = value
	return se
}

func (se *StringExpression) Eval(s string) (string, error) {
	for {
		//find variable start indicator
		start := strings.Index(s, "%(")

		if start == -1 {
			return s, nil
		}

		end := start + 1
		n := len(s)

		//find variable end indicator
		for end < n && s[end] != ')' {
			end++
		}

		//find the type of the variable
		typ := end + 1
		for typ < n && !((s[typ] >= 'a' && s[typ] <= 'z') || (s[typ] >= 'A' && s[typ] <= 'Z')) {
			typ++
		}

		//evaluate the variable
		if typ < n {
			varName := s[start+2 : end]

			varValue, ok := se.env[varName]

			if !ok {
				return "", fmt.Errorf("fail to find the environment variable %s", varName)
			}
			if s[typ] == 'd' {
				i, err := strconv.Atoi(varValue)
				if err != nil {
					return "", fmt.Errorf("can't convert %s to integer", varValue)
				}
				s = s[0:start] + fmt.Sprintf("%"+s[end+1:typ+1], i) + s[typ+1:]
			} else if s[typ] == 's' {
				s = s[0:start] + varValue + s[typ+1:]
			} else {
				return "", fmt.Errorf("not implement type:%v", s[typ])
			}
		} else {
			return "", fmt.Errorf("invalid string expression format")
		}
	}

}
