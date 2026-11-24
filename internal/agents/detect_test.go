package agents

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testcases.json
var testcasesJSON []byte

type testcase struct {
	Name     string            `json:"name"`
	Env      map[string]string `json:"env"`
	Expected *string           `json:"expected"`
}

func fakeEnv(vars map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := vars[key]
		return v, ok
	}
}

func TestEmbeddedRules(t *testing.T) {
	r, err := parse(agentsJSON)
	require.NoError(t, err)
	assert.Equal(t, "AI_AGENT", r.AIAgentVar)
	require.NotEmpty(t, r.Agents)

	supported := map[string]bool{"anyOf": true, "allOf": true, "env_set": true, "env_value": true}
	var check func(t *testing.T, c condition)
	check = func(t *testing.T, c condition) {
		assert.True(t, supported[c.Type], "unsupported condition type %q", c.Type)
		for _, sub := range c.Conditions {
			check(t, sub)
		}
	}

	seen := map[string]bool{}
	for _, agent := range r.Agents {
		assert.Regexp(t, `^[a-z0-9]+(-[a-z0-9]+)*$`, agent.Name)
		assert.False(t, seen[agent.Name], "duplicate agent name %q", agent.Name)
		seen[agent.Name] = true
		assert.Contains(t, []string{"anyOf", "allOf"}, agent.Match.Type, agent.Name)
		check(t, agent.Match)
	}
}

func TestSharedTestcases(t *testing.T) {
	r, err := parse(agentsJSON)
	require.NoError(t, err)

	var cases []testcase
	require.NoError(t, json.Unmarshal(testcasesJSON, &cases))
	require.NotEmpty(t, cases)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			want := ""
			if tc.Expected != nil {
				want = *tc.Expected
			}
			assert.Equal(t, want, r.detect(fakeEnv(tc.Env)))
		})
	}
}

func TestDetect_ReadsEnvOnEachCall(t *testing.T) {
	t.Setenv("AI_AGENT", "first-agent")
	assert.Equal(t, "first-agent", Detect())

	t.Setenv("AI_AGENT", "second-agent")
	assert.Equal(t, "second-agent", Detect())
}

func TestConditionCombinators(t *testing.T) {
	env := fakeEnv(map[string]string{"A": "1", "B": "x"})
	set := func(name string) condition { return condition{Type: "env_set", Name: name} }

	assert.True(t, condition{Type: "allOf", Conditions: []condition{set("A"), {Type: "env_value", Name: "B", Value: "x"}}}.matches(env))
	assert.False(t, condition{Type: "allOf", Conditions: []condition{set("A"), set("C")}}.matches(env))
	assert.True(t, condition{Type: "anyOf", Conditions: []condition{set("C"), set("A")}}.matches(env))
	assert.False(t, condition{Type: "anyOf"}.matches(env))
	assert.False(t, condition{Type: "file_exists", Name: "/"}.matches(env))
}

func TestParse_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "not json", data: `not json`},
		{name: "missing version", data: `{"namePattern": ".*"}`},
		{name: "unsupported version", data: `{"version": 2, "namePattern": ".*"}`},
		{name: "invalid namePattern", data: `{"version": 1, "namePattern": "("}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parse([]byte(tt.data))
			assert.Error(t, err)
		})
	}
}
