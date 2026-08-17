package execrecovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const FixtureVersion = 1

// Kind labels how a fixture relates to production.
//
// observed: a response the current handler can emit.
// defensive: a client-side invariant against a body KEEP-966 makes unreachable.
// classifier: exercises a classifier option the shipped CLI does not set.
type Kind string

const (
	KindObserved   Kind = "observed"
	KindDefensive  Kind = "defensive"
	KindClassifier Kind = "classifier"
)

// Fixture is one conformance case loaded from testdata/execution_recovery_v1.
type Fixture struct {
	Name                 string          `json:"-"`
	Version              int             `json:"version"`
	Kind                 Kind            `json:"kind"`
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
	Name    string         `json:"name"`
	Version int            `json:"version"`
	Kind    Kind           `json:"kind"`
	Rule    string         `json:"rule"`
	Steps   []SequenceStep `json:"steps"`
}

func validateMeta(name string, version int, kind Kind) error {
	if version != FixtureVersion {
		return fmt.Errorf("%s: unsupported fixture version %d (want %d)", name, version, FixtureVersion)
	}
	switch kind {
	case KindObserved, KindDefensive, KindClassifier:
		return nil
	case "":
		return fmt.Errorf("%s: missing kind (observed|defensive|classifier)", name)
	default:
		return fmt.Errorf("%s: unknown kind %q", name, kind)
	}
}

func isSequenceName(name string) bool {
	return strings.HasSuffix(name, ".sequence.json")
}

// LoadFixtureDir loads every versioned *.json fixture that is not a sequence.
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
		if filepath.Ext(name) != ".json" || isSequenceName(name) {
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
		if err := validateMeta(name, f.Version, f.Kind); err != nil {
			return nil, err
		}
		if f.Rule == "" {
			return nil, fmt.Errorf("%s: missing rule", name)
		}
		f.Name = name
		out = append(out, f)
	}
	return out, nil
}

// LoadSequenceDir loads every *.sequence.json fixture in dir.
func LoadSequenceDir(dir string) ([]SequenceFixture, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []SequenceFixture
	for _, e := range entries {
		if e.IsDir() || !isSequenceName(e.Name()) {
			continue
		}
		seq, err := LoadSequence(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, seq)
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
	name := filepath.Base(path)
	if seq.Name == "" {
		seq.Name = name
	}
	if err := validateMeta(name, seq.Version, seq.Kind); err != nil {
		return SequenceFixture{}, err
	}
	if seq.Rule == "" {
		return SequenceFixture{}, fmt.Errorf("%s: missing rule", name)
	}
	if len(seq.Steps) == 0 {
		return SequenceFixture{}, fmt.Errorf("%s: sequence has no steps", name)
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
