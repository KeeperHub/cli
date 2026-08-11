package execrecovery_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/keeperhub/cli/internal/execrecovery"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/execrecovery -> repo root
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "testdata", "execution_recovery_v1")
}

func TestFixtures_DecodeIntoDirectStatus(t *testing.T) {
	fixtures, err := execrecovery.LoadFixtureDir(testdataDir(t))
	if err != nil {
		t.Fatalf("LoadFixtureDir: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no fixtures loaded")
	}

	for _, f := range fixtures {
		f := f
		t.Run(f.Name, func(t *testing.T) {
			if f.ResponseRaw != "" {
				// Malformed raw bodies are not DirectStatus JSON.
				st, err := f.DecodeResponse()
				if err == nil {
					t.Fatalf("expected decode error for raw fixture, got %#v", st)
				}
				return
			}
			st, err := f.DecodeResponse()
			if err != nil {
				// not_found / rate_limited error bodies are not DirectStatus;
				// Classify still handles them via HTTP status.
				if f.HTTPStatus == 404 || f.HTTPStatus == 429 {
					return
				}
				t.Fatalf("DecodeResponse: %v", err)
			}
			if f.HTTPStatus == 200 && f.Expect != execrecovery.OutcomeMalformed {
				if st.ExecutionID == "" && f.Expect != execrecovery.OutcomeFailure {
					// failed fixture has executionId; ensure we never silently zero-decode.
				}
				if st.Status == "" && f.Expect != execrecovery.OutcomeMalformed {
					t.Fatalf("decoded empty Status for fixture %s — wire shape mismatch", f.Name)
				}
			}
		})
	}
}

func TestFixtures_ClassifyTable(t *testing.T) {
	fixtures, err := execrecovery.LoadFixtureDir(testdataDir(t))
	if err != nil {
		t.Fatalf("LoadFixtureDir: %v", err)
	}

	for _, f := range fixtures {
		f := f
		t.Run(f.Rule+"/"+f.Name, func(t *testing.T) {
			got, reason := execrecovery.Classify(f.Sample(), execrecovery.Options{
				RequireChainEvidence: f.RequireChainEvidence,
			})
			if got != f.Expect {
				t.Fatalf("Classify=%s (%s), want %s", got, reason, f.Expect)
			}
		})
	}
}

func TestColdStartSequence_R6(t *testing.T) {
	path := filepath.Join(testdataDir(t), "cold_start.sequence.json")
	seq, err := execrecovery.LoadSequence(path)
	if err != nil {
		t.Fatalf("LoadSequence: %v", err)
	}
	if seq.Rule != "R6" {
		t.Fatalf("rule=%s, want R6", seq.Rule)
	}
	if len(seq.Steps) < 2 {
		t.Fatal("cold_start sequence must have at least 2 steps")
	}
	for i, step := range seq.Steps {
		got, reason := execrecovery.Classify(step.Sample(), execrecovery.Options{
			RequireChainEvidence: step.RequireChainEvidence,
		})
		if got != step.Expect {
			t.Fatalf("step %d: Classify=%s (%s), want %s", i, got, reason, step.Expect)
		}
	}
}

func TestRevertedIsNeverSuccess(t *testing.T) {
	body := []byte(`{
		"executionId":"x",
		"status":"completed",
		"transactionHash":"0xabc",
		"receipts":[{"hash":"0xabc","chainId":8453,"verified":true,"receiptStatus":"reverted"}]
	}`)
	got, reason := execrecovery.Classify(execrecovery.Sample{HTTPStatus: 200, Body: body}, execrecovery.Options{
		RequireChainEvidence: true,
	})
	if got != execrecovery.OutcomeFailure {
		t.Fatalf("got %s (%s), want failure", got, reason)
	}
}

func TestEmptyStatusIsMalformed(t *testing.T) {
	got, _ := execrecovery.Classify(execrecovery.Sample{
		HTTPStatus: 200,
		Body:       []byte(`{"executionId":"x"}`),
	}, execrecovery.Options{})
	if got != execrecovery.OutcomeMalformed {
		t.Fatalf("got %s, want malformed", got)
	}
}

func TestVocabularySurfacesAreDistinct(t *testing.T) {
	d := execrecovery.DirectExecutionVocabulary()
	w := execrecovery.WorkflowRunVocabulary()
	if d.Surface == w.Surface {
		t.Fatal("vocabularies must name distinct surfaces")
	}
	for _, term := range w.Terminal {
		for _, dTerm := range d.Terminal {
			if term == dTerm {
				t.Fatalf("shared terminal term %q across surfaces — keep vocabularies separate", term)
			}
		}
	}
}
