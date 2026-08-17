package execrecovery

// Vocabulary documents which status strings belong to which API surface.
// Direct-execution and workflow-run statuses must not be mixed.
type Vocabulary struct {
	Surface  string
	Pending  []string
	Terminal []string
}

// DirectExecutionVocabulary is GET /api/execute/{id}/status
// (app/api/execute/_lib/types.ts ExecutionStatus).
func DirectExecutionVocabulary() Vocabulary {
	return Vocabulary{
		Surface:  "direct-execution",
		Pending:  []string{"pending", "running", "unconfirmed"},
		Terminal: []string{"completed", "failed"},
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
