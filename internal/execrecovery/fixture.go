package execrecovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Fixture is one conformance case loaded from testdata/execution_recovery_v1.
type Fixture struct {
	Name                 string          `json:"-"`
	Rule                 string          `json:"rule"`
	HTTPStatus           int             `json:"httpStatus"`
	RequireChainEvidence bool            `json:"requireChainEvidence"`
	Expect               Outcome         `json:"expect"`
	Response             json.RawMessage `json:"response"`
	ResponseRaw          string          `json:"responseRaw,omitempty"`
	Note                 string          `json:"note,omitempty"`
}

// SequenceStep is one observation in a multi-response cold-start sequence.
type SequenceStep struct {
	HTTPStatus           int             `json:"httpStatus"`
	RequireChainEvidence bool            `json:"requireChainEvidence"`
	Expect               Outcome         `json:"expect"`
	Response             json.RawMessage `json:"response"`
	ResponseRaw          string          `json:"responseRaw,omitempty"`
}

// SequenceFixture exercises multi-poll recovery (R6).
type SequenceFixture struct {
	Name  string         `json:"name"`
	Rule  string         `json:"rule"`
	Steps []SequenceStep `json:"steps"`
}

// LoadFixtureDir loads every *.json fixture (not *.sequence.json) from dir.
func LoadFixtureDir(dir string) ([]Fixture, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []Fixture
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}
		if len(name) >= len(".sequence.json") && name[len(name)-len(".sequence.json"):] == ".sequence.json" {
			continue
		}
		path := filepath.Join(dir, name)
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var f Fixture
		if err := json.Unmarshal(raw, &f); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		f.Name = name
		out = append(out, f)
	}
	return out, nil
}

// LoadSequence loads a multi-step sequence fixture.
func LoadSequence(path string) (SequenceFixture, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return SequenceFixture{}, err
	}
	var seq SequenceFixture
	if err := json.Unmarshal(raw, &seq); err != nil {
		return SequenceFixture{}, err
	}
	return seq, nil
}

// Sample converts a fixture into a Classify input.
func (f Fixture) Sample() Sample {
	body := []byte(f.ResponseRaw)
	if len(body) == 0 {
		body = f.Response
	}
	return Sample{HTTPStatus: f.HTTPStatus, Body: body}
}

// Sample converts a sequence step into a Classify input.
func (s SequenceStep) Sample() Sample {
	body := []byte(s.ResponseRaw)
	if len(body) == 0 {
		body = s.Response
	}
	return Sample{HTTPStatus: s.HTTPStatus, Body: body}
}

// DecodeResponse unmarshals the flat DirectStatus wire body.
func (f Fixture) DecodeResponse() (DirectStatus, error) {
	if f.ResponseRaw != "" {
		return DirectStatus{}, fmt.Errorf("raw body is not DirectStatus JSON")
	}
	var st DirectStatus
	if err := json.Unmarshal(f.Response, &st); err != nil {
		return DirectStatus{}, err
	}
	return st, nil
}
