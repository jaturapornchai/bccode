package generalledger

import "context"

// JournalForAuthorization includes tombstones so retries of successful deletes
// can reach the idempotency check. It is never exposed as a read endpoint.
func (s *Store) JournalForAuthorization(ctx context.Context, scope Scope, id string) (Journal, error) {
	var journal Journal
	err := s.load(ctx, scope, "journals", id, &journal)
	if err == nil && scope.Branch != "" && journal.BranchCode != scope.Branch {
		err = ErrNotFound
	}
	return journal, err
}
