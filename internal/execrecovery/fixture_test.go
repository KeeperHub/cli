package execrecovery_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/keeperhub/cli/internal/execrecovery"
)

const (
	wantFixtures  = 11
	wantSequences = 1
)

func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "testdata", "execution_recovery_v1")
}

func TestLoadFixtureDir_EmptyFails(t *testing.T) {
	dir := t.TempDir()
	fixtures, err := execrecovery.LoadFixtureDir(dir)
	if err != nil {
		t.Fatalf("empty dir should load, got err %v", err)
	}
	if len(fixtures) != 0 {
		t.Fatalf("got %d fixtures, want 0", len(fixtures))
	}
}

func TestFixtures_ExpectedCountsAndRules(t *testing.T) {
	dir := testdataDir(t)
	fixtures, err := execrecovery.LoadFixtureDir(dir)
	if err != nil {
		t.Fatalf("LoadFixtureDir: %v", err)
	}
	if len(fixtures) != wantFixtures {
		t.Fatalf("loaded %d fixtures, want %d (renaming a file to *.sequence.json must not silently drop coverage)", len(fixtures), wantFixtures)
	}
	seqs, err := execrecovery.LoadSequenceDir(dir)
	if err != nil {
		t.Fatalf("LoadSequenceDir: %v", err)
	}
	if len(seqs) != wantSequences {
		t.Fatalf("loaded %d sequences, want %d", len(seqs), wantSequences)
	}

	rules := map[string]int{}
	kinds := map[execrecovery.Kind]int{}
	for _, f := range fixtures {
		if f.Version != execrecovery.FixtureVersion {
			t.Fatalf("%s: version %d", f.Name, f.Version)
		}
		if f.Rule == "" {
			t.Fatalf("%s: missing rule", f.Name)
		}
		rules[f.Rule]++
		kinds[f.Kind]++
	}
	for _, need := range []string{"R1", "R2", "R4", "R5", "R6"} {
		if rules[need] == 0 {
			t.Fatalf("no fixture maps to rule %s", need)
		}
	}
	if kinds[execrecovery.KindObserved] == 0 {
		t.Fatal("expected at least one observed fixture")
	}
	if kinds[execrecovery.KindDefensive] == 0 {
		t.Fatal("expected at least one defensive fixture")
	}
}

func TestLoadFixtureDir_RejectsMissingVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no_version.json")
	body := []byte(`{"kind":"observed","rule":"R1","httpStatus":200,"expect":"pending","response":{"status":"pending"}}`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := execrecovery.LoadFixtureDir(dir); err == nil {
		t.Fatal("expected version error")
	}
}

func TestFixtures_DecodeIntoDirectStatus(t *testing.T) {
	fixtures, err := execrecovery.LoadFixtureDir(testdataDir(t))
	if err != nil {
		t.Fatalf("LoadFixtureDir: %v", err)
	}
	if len(fixtures) != wantFixtures {
		t.Fatalf("loaded %d fixtures, want %d", len(fixtures), wantFixtures)
	}

	for _, f := range fixtures {
		f := f
		t.Run(f.Name, func(t *testing.T) {
			if f.ResponseRaw != "" {
				st, err := f.DecodeResponse()
				if err == nil {
					t.Fatalf("expected decode error for raw fixture, got %#v", st)
				}
				return
			}
			st, err := f.DecodeResponse()
			if err != nil {
				if f.HTTPStatus == 404 || f.HTTPStatus == 429 {
					return
				}
				t.Fatalf("DecodeResponse: %v", err)
			}
			if f.HTTPStatus == 200 && f.Expect != execrecovery.OutcomeMalformed {
				if st.Status == "" {
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
	if len(fixtures) != wantFixtures {
		t.Fatalf("loaded %d fixtures, want %d", len(fixtures), wantFixtures)
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
	seqs, err := execrecovery.LoadSequenceDir(testdataDir(t))
	if err != nil {
		t.Fatalf("LoadSequenceDir: %v", err)
	}
	if len(seqs) != wantSequences {
		t.Fatalf("sequences=%d, want %d", len(seqs), wantSequences)
	}
	seq := seqs[0]
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
		"receipts":[{"hash":"0xabc","verified":true,"receiptStatus":"reverted","verifiedAt":"2026-08-11T00:00:00Z"}]
	}`)
	got, reason := execrecovery.Classify(execrecovery.Sample{HTTPStatus: 200, Body: body}, execrecovery.Options{})
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

func TestUnknownStatusIsUnrecognizedNotMalformed(t *testing.T) {
	got, reason := execrecovery.Classify(execrecovery.Sample{
		HTTPStatus: 200,
		Body:       []byte(`{"executionId":"x","status":"settling"}`),
	}, execrecovery.Options{})
	if got != execrecovery.OutcomeUnrecognized {
		t.Fatalf("got %s (%s), want unrecognized", got, reason)
	}
}

func TestWorkflowStatusesAreNotDirectSuccess(t *testing.T) {
	for _, status := range []string{"success", "error", "cancelled", "queued"} {
		got, _ := execrecovery.Classify(execrecovery.Sample{
			HTTPStatus: 200,
			Body:       []byte(`{"executionId":"x","status":"` + status + `"}`),
		}, execrecovery.Options{})
		if got == execrecovery.OutcomeSuccess {
			t.Fatalf("status %s must not classify as success", status)
		}
		if got == execrecovery.OutcomeMalformed {
			t.Fatalf("status %s must not classify as malformed (would break a future enum addition)", status)
		}
	}
}

func TestVocabularySurfacesAreDistinct(t *testing.T) {
	d := execrecovery.DirectExecutionVocabulary()
	w := execrecovery.WorkflowRunVocabulary()
	if d.Surface == w.Surface {
		t.Fatal("vocabularies must name distinct surfaces")
	}
	direct := map[string]struct{}{}
	for _, s := range append(append([]string{}, d.Pending...), d.Terminal...) {
		direct[s] = struct{}{}
	}
	for _, term := range w.Terminal {
		if _, ok := direct[term]; ok {
			t.Fatalf("workflow terminal %q must not appear in direct-execution vocabulary", term)
		}
	}
	for _, s := range d.Pending {
		if s == "queued" || s == "not_found" {
			t.Fatalf("direct pending must not include %q (not in server ExecutionStatus)", s)
		}
	}
}
