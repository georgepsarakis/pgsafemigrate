package rules

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

type NestedTransaction struct{}

func (r NestedTransaction) Alias() string {
	return TransactionRule("no-nested-transactions")
}

func (r NestedTransaction) Documentation() string {
	return "Nested transactions are not supported in PostgreSQL."
}

func (r NestedTransaction) Process(node *pg_query.Node, _ []*pg_query.Node, inTransaction bool) bool {
	if !inTransaction {
		return false
	}
	transactionStmt := node.GetTransactionStmt()
	if transactionStmt == nil {
		return false
	}
	switch transactionStmt.GetKind() {
	case pg_query.TransactionStmtKind_TRANS_STMT_BEGIN,
		pg_query.TransactionStmtKind_TRANS_STMT_COMMIT,
		pg_query.TransactionStmtKind_TRANS_STMT_ROLLBACK:
		return true
	}
	return false
}

type TransactionNotSupportedInConcurrentIndexOperations struct{}

func (t TransactionNotSupportedInConcurrentIndexOperations) Documentation() string {
	return "Concurrent index operations cannot be executed inside a transaction."
}

func (t TransactionNotSupportedInConcurrentIndexOperations) Alias() string {
	return TransactionRule("concurrent-index-operation-cannot-be-executed-in-transaction")
}

func (t TransactionNotSupportedInConcurrentIndexOperations) Process(node *pg_query.Node, _ []*pg_query.Node, inTransaction bool) bool {
	if !inTransaction {
		return false
	}
	indexStmt := node.GetIndexStmt()
	if indexStmt != nil {
		return indexStmt.Concurrent
	}

	dropStmt := node.GetDropStmt()
	if dropStmt == nil {
		return false
	}
	if dropStmt.RemoveType != pg_query.ObjectType_OBJECT_INDEX {
		return false
	}
	return dropStmt.GetConcurrent()
}
