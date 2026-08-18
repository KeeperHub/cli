package execrecovery

// Vocabulary documents which status strings belong to which API surface.
// Direct-execution and workflow-run statuses must not be mixed.
//
// Pending and Terminal are client wait semantics: Pending means keep polling,
// Terminal means stop waiting and report. Terminal is not a claim that the
// server will never change the row again.
type Vocabulary struct {
	Surface  string
	Pending  []string
	Terminal []string
}

// DirectExecutionVocabulary is GET /api/execute/{id}/status
// (app/api/execute/_lib/types.ts ExecutionStatus).
//
// `unconfirmed` is listed Terminal in the client sense only. The server
// documents it as non-terminal and a reconciliation sweep settles it to
// completed or failed; clients still stop there so that an unreadable receipt
// never becomes a re-run that broadcasts twice.
func DirectExecutionVocabulary() Vocabulary {
	return Vocabulary{
		Surface:  "direct-execution",
		Pending:  []string{"pending", "running"},
		Terminal: []string{"unconfirmed", "completed", "failed"},
	}
}

// WorkflowRunVocabulary is GET /api/workflows/executions/{id}/status.
func WorkflowRunVocabulary() Vocabulary {
	return Vocabulary{
		Surface:  "workflow-run",
		Pending:  []string{"pending", "running"},
		Terminal: []string{"success", "error", "cancelled"},
	}
}
