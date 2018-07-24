package idetcd

import (
	"strings"
	"testing"
	"text/template"

	"github.com/mholt/caddy"
)

func TestParseIdetcd(t *testing.T) {
	tests := []struct {
		input              string
		shouldErr          bool
		expectedEndpoint   []string
		expectedLimit      int
		expectedPattern    *template.Template
		expectedErrContent string
	}{
		{
			`idetcd {
				endpoint http://localhost:2379
				pattern worker{{.ID}}.local.tf
				limit 5
		}`, false, []string{"http://localhost:2379"}, 5, getExpectedPattern(), "",
		},
		{
			`idetcd {
				endpoint
				pattern worker{{.ID}}.local.tf
				limit 5
		}`, true, []string{"http://localhost:2379"}, 5, getExpectedPattern(), "",
		},
		{
			`idetcd {
				endpoint http://localhost:2379
				pattern
				limit 5
		}`, true, []string{"http://localhost:2379"}, 5, nil, "",
		},
		{
			`idetcd {
				endpoint http://localhost:2379
				pattern worker{{.ID}}.local.tf
				limit
		}`, true, []string{"http://localhost:2379"}, 5, getExpectedPattern(), "",
		},
		{
			`idetcd {
				endpoint http://localhost:2379
				pattern worker{{.ID}}.local.tf
				limit hello
		}`, true, []string{"http://localhost:2379"}, 5, getExpectedPattern(), "",
		},
		{
			`idetcd {
				endpoint http://localhost:2379 http://localhost:3379 http://localhost:4379
				pattern worker{{.ID}}.local.tf
				limit 5
		}`, false, []string{"http://localhost:2379", "http://localhost:3379", "http://localhost:4379"}, 5, getExpectedPattern(), "",
		},
	}
	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		idetcd, err := idetcdParse(c)
		if test.shouldErr && err == nil {
			t.Errorf("Test %d: Expected error but found %s for input %s", i, err, test.input)
		}
		if err != nil {
			if !test.shouldErr {
				t.Errorf("Test %d: Expected no error but found one for input %s. Error was: %v", i, test.input, err)
				continue
			}
			if !strings.Contains(err.Error(), test.expectedErrContent) {
				t.Errorf("Test %d: Expected error to contain: %v, found error: %v, input: %s", i, test.expectedErrContent, err, test.input)
				continue
			}

		}
		if !test.shouldErr && idetcd.limit != test.expectedLimit {
			t.Errorf("Idetcd not correctly set for input %s. Expected: %d, actual: %d", test.input, test.expectedLimit, idetcd.limit)
		}
		if !test.shouldErr {
			if len(idetcd.endpoints) != len(test.expectedEndpoint) {
				t.Errorf("Etcd not correctly set for input %s. Expected: '%+v', actual: '%+v'", test.input, test.expectedEndpoint, idetcd.endpoints)
			}
			for i, endpoint := range idetcd.endpoints {
				if endpoint != test.expectedEndpoint[i] {
					t.Errorf("Etcd not correctly set for input %s. Expected: '%+v', actual: '%+v'", test.input, test.expectedEndpoint, idetcd.endpoints)
				}
			}
		}
	}
}

func getExpectedPattern() *template.Template {
	pattern := template.New("idetcd")
	pattern, err := pattern.Parse("worker{{.ID}}.local.tf")
	if err != nil {
		return nil
	}
	return pattern
}

func TestSetup(t *testing.T) {
	config := `idetcd {
			endpoint http://localhost:2379
			pattern worker{{.ID}}.local.tf
			limit 5
		}`
	c := caddy.NewTestController("dns", config)
	err := setup(c)
	if err != nil {
		t.Fatalf("Shouldn't fail")
	}
}
