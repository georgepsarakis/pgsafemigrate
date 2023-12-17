package rules

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"strings"
)

type CreateIndexNonConcurrently struct{}

func (r CreateIndexNonConcurrently) Alias() string {
	return HighAvailabilityRule("avoid-non-concurrent-index-creation")
}

func (r CreateIndexNonConcurrently) Documentation() string {
	return "Non-concurrent index creation will not allow writes while the index is being built."
}

func (r CreateIndexNonConcurrently) Process(node *pg_query.Node, _ []*pg_query.Node, _ bool) bool {
	indexStmt := node.GetIndexStmt()
	if indexStmt == nil {
		return false
	}
	return !indexStmt.Concurrent
}

type DropIndexNonConcurrently struct{}

func (r DropIndexNonConcurrently) Alias() string {
	return HighAvailabilityRule("avoid-non-concurrent-index-drop")
}

func (r DropIndexNonConcurrently) Documentation() string {
	return "Non-concurrent index drop will not allow writes while the index is being built."
}

func (r DropIndexNonConcurrently) Process(node *pg_query.Node, _ []*pg_query.Node, _ bool) bool {
	dropStmt := node.GetDropStmt()
	if dropStmt == nil {
		return false
	}
	if dropStmt.RemoveType != pg_query.ObjectType_OBJECT_INDEX {
		return false
	}
	return !dropStmt.GetConcurrent()
}

type IndexOperationNotIdempotent struct{}

func (r IndexOperationNotIdempotent) Alias() string {
	return TransactionRule("index-if-not-exists-missing")
}

func (r IndexOperationNotIdempotent) Documentation() string {
	return "Creating/removing an index outside of a transaction without an IF (NOT) EXISTS option can cause a migration to not be idempotent."
}

func (r IndexOperationNotIdempotent) Process(node *pg_query.Node, _ []*pg_query.Node, inTransaction bool) bool {
	if inTransaction {
		return false
	}

	createStmt := node.GetIndexStmt()
	dropStmt := node.GetDropStmt()
	if (dropStmt != nil && dropStmt.RemoveType != pg_query.ObjectType_OBJECT_INDEX) && createStmt == nil {
		return false
	}
	if dropStmt != nil && !dropStmt.MissingOk {
		return true
	}
	if createStmt != nil && !createStmt.IfNotExists {
		return true
	}
	return false
}

type IndexMustBeNamed struct{}

func (r IndexMustBeNamed) Alias() string {
	return MaintainabilityRule("indexes-name-is-required")
}

func (r IndexMustBeNamed) Documentation() string {
	return "Indexes should be explicitly named."
}

func (r IndexMustBeNamed) Process(node *pg_query.Node, _ []*pg_query.Node, _ bool) bool {
	createStmt := node.GetIndexStmt()
	if createStmt == nil {
		return false
	}
	return strings.TrimSpace(createStmt.GetIdxname()) == ""
}
