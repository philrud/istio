package model

import (
	"fmt"
	"testing"
)

// Regression test: https://github.com/istio/istio/issues/12066
func TestPushContext_DestinationRule_collidingLocalAndTargetNamespaceRules(t *testing.T) {
	configLocal := Config{
		ConfigMeta: ConfigMeta{Name: "local"},
	}
	configTarget := Config{
		ConfigMeta: ConfigMeta{Name: "target"},
	}
	configAll := Config{
		ConfigMeta: ConfigMeta{Name: "all"},
	}

	tests := []struct {
		name        string
		localRules  processedDestRules
		targetRules processedDestRules
		result      *Config
	}{
		{
			"has-local-no-target",
			processedDestRules{
				hosts: []Hostname{
					"*",
				},
				destRule: map[Hostname]*combinedDestinationRule{
					"*": {config: &configLocal},
				},
			},
			processedDestRules{
				hosts:    []Hostname{},
				destRule: map[Hostname]*combinedDestinationRule{},
			},
			&configLocal,
		},
		{
			"no-local-has-target",
			processedDestRules{
				hosts:    []Hostname{},
				destRule: map[Hostname]*combinedDestinationRule{},
			},
			processedDestRules{
				hosts: []Hostname{
					"*.foo.bar",
				},
				destRule: map[Hostname]*combinedDestinationRule{
					"*.foo.bar": {config: &configTarget},
				},
			},
			&configTarget,
		},
		{
			"no-local-no-target",
			processedDestRules{
				hosts:    []Hostname{},
				destRule: map[Hostname]*combinedDestinationRule{},
			},
			processedDestRules{
				hosts:    []Hostname{},
				destRule: map[Hostname]*combinedDestinationRule{},
			},
			&configAll,
		},
		{
			"local-is-more-specific",
			processedDestRules{
				hosts: []Hostname{
					"*.foo.bar",
				},
				destRule: map[Hostname]*combinedDestinationRule{
					"*.foo.bar": {config: &configLocal},
				},
			},
			processedDestRules{
				hosts: []Hostname{
					"*",
				},
				destRule: map[Hostname]*combinedDestinationRule{
					"*": {config: &configTarget},
				},
			},
			&configLocal,
		},
		{
			"target-is-more-specific",
			processedDestRules{
				hosts: []Hostname{
					"*",
				},
				destRule: map[Hostname]*combinedDestinationRule{
					"*": {config: &configLocal},
				},
			},
			processedDestRules{
				hosts: []Hostname{
					"*.foo.bar",
				},
				destRule: map[Hostname]*combinedDestinationRule{
					"*.foo.bar": {config: &configTarget},
				},
			},
			&configTarget,
		},
		{
			"same-local-and-target",
			processedDestRules{
				hosts: []Hostname{
					"*.foo.bar",
				},
				destRule: map[Hostname]*combinedDestinationRule{
					"*.foo.bar": {config: &configLocal},
				},
			},
			processedDestRules{
				hosts: []Hostname{
					"*.foo.bar",
				},
				destRule: map[Hostname]*combinedDestinationRule{
					"*.foo.bar": {config: &configTarget},
				},
			},
			&configLocal,
		},
	}

	for idx, tc := range tests {
		t.Run(fmt.Sprintf("[%d] %s", idx, tc.name), func(t *testing.T) {
			proxy := Proxy{
				SidecarScope:    nil,
				Type:            Router,
				ConfigNamespace: "istio-system",
			}
			service := Service{
				Attributes: ServiceAttributes{
					Namespace: "default",
				},
				Hostname: "baz.foo.bar",
			}
			pushContext := PushContext{
				namespaceLocalDestRules: map[string]*processedDestRules{
					"istio-system": &tc.localRules,
				},
				namespaceExportedDestRules: map[string]*processedDestRules{
					"default": &tc.targetRules,
				},
				allExportedDestRules: &processedDestRules{
					hosts: []Hostname{"baz.foo.bar"},
					destRule: map[Hostname]*combinedDestinationRule{
						"baz.foo.bar": {
							config: &configAll,
						},
					},
				},
			}

			result := pushContext.DestinationRule(&proxy, &service)
			if tc.result != result {
				t.Errorf("Expected result: %v. Actual result: %v.", tc.result, result)
			}
		})
	}
}
