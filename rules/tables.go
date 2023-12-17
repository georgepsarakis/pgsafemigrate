package rules

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

// RenameTable - Renaming a table can cause downtime to the previous application service version.
type RenameTable struct{}

func (r RenameTable) Alias() string {
	return HighAvailabilityRule("avoid-table-rename")
}
func (r RenameTable) Documentation() string {
	return "Renaming a table can cause errors in previous application versions."
}
func (r RenameTable) Process(node *pg_query.Node, _ []*pg_query.Node, _ bool) bool {
	renameTableStmt := node.GetRenameStmt()
	if renameTableStmt == nil {
		return false
	}
	return renameTableStmt.RenameType == pg_query.ObjectType_OBJECT_TABLE
}

// RequiredColumn - Adding a non-nullable column without a default value, makes the column required
type RequiredColumn struct{}

func (r RequiredColumn) Alias() string {
	return HighAvailabilityRule("avoid-required-column")
}
func (r RequiredColumn) Documentation() string {
	return "Newly added columns must either define a default value or be nullable."
}

func (r RequiredColumn) Process(node *pg_query.Node, _ []*pg_query.Node, _ bool) bool {
	alterTable := node.GetAlterTableStmt()
	if alterTable == nil {
		return false
	}
	if alterTable.Objtype != pg_query.ObjectType_OBJECT_TABLE {
		return false
	}

	for _, cmd := range alterTable.GetCmds() {
		alterTableCmd := cmd.GetAlterTableCmd()
		if alterTableCmd == nil {
			continue
		}
		if alterTableCmd.Subtype != pg_query.AlterTableType_AT_AddColumn {
			continue
		}
		var isNotNull bool
		var hasDefault bool
		for _, constraint := range alterTableCmd.GetDef().GetColumnDef().GetConstraints() {
			if constraint.GetConstraint().GetContype() == pg_query.ConstrType_CONSTR_NOTNULL {
				isNotNull = true
			}
			if constraint.GetConstraint().GetContype() == pg_query.ConstrType_CONSTR_DEFAULT {
				hasDefault = true
			}
		}

		if isNotNull && !hasDefault {
			return true
		}
	}
	return false
}

type ColumnComment struct{}

func (r ColumnComment) Alias() string {
	return MaintainabilityRule("describe-new-column-with-comment")
}

func (r ColumnComment) Documentation() string {
	return "Newly added columns should also include a COMMENT for documentation purposes."
}

func (r ColumnComment) Process(node *pg_query.Node, allNodes []*pg_query.Node, _ bool) bool {
	alterTable := node.GetAlterTableStmt()
	if alterTable == nil {
		return false
	}
	if alterTable.Objtype != pg_query.ObjectType_OBJECT_TABLE {
		return false
	}

	colNames := mapset.NewSet[string]()
	for _, cmd := range alterTable.GetCmds() {
		if cmd.GetAlterTableCmd().Subtype != pg_query.AlterTableType_AT_AddColumn {
			continue
		}
		colNames.Add(fmt.Sprintf("%s.%s", alterTable.Relation.Relname, cmd.GetAlterTableCmd().GetDef().GetColumnDef().Colname))
	}
	if colNames.Cardinality() == 0 {
		return false
	}
	for _, n := range allNodes {
		comment := n.GetCommentStmt()
		if comment == nil {
			continue
		}
		if comment.GetObjtype() != pg_query.ObjectType_OBJECT_COLUMN {
			continue
		}
		if strings.TrimSpace(comment.GetComment()) == "" {
			continue
		}
		var fullyQualifiedColumnName string
		for _, item := range comment.GetObject().GetList().GetItems() {
			fullyQualifiedColumnName += item.GetString_().GetSval() + "."
		}
		fullyQualifiedColumnName = strings.TrimSuffix(fullyQualifiedColumnName, ".")
		if colNames.Contains(fullyQualifiedColumnName) {
			return false
		}
	}
	return true
}

type ColumnSetNotNull struct{}

func (r ColumnSetNotNull) Alias() string {
	return HighAvailabilityRule("alter-column-not-null-exclusive-lock")
}

// https://dba.stackexchange.com/a/268128
// - Add NOT NULL constraint marked as NOT VALID
// - Run ALTER TABLE ... VALIDATE on the constraint
// - Run ALTER TABLE ... ALTER COLUMN ... SET NOT NULL
// - Drop the constraint
func (r ColumnSetNotNull) Documentation() string {
	return "Setting a column as NOT NULL acquires an exclusive lock on the table until the constraint is validated on all table rows."
}

func (r ColumnSetNotNull) Process(node *pg_query.Node, allNodes []*pg_query.Node, _ bool) bool {
	alterTableStmt := node.GetAlterTableStmt()
	if alterTableStmt == nil {
		return false
	}
	tableName := alterTableStmt.GetRelation().GetRelname()
	var colName string
	for _, cmd := range alterTableStmt.GetCmds() {
		if alterTableCmd := cmd.GetAlterTableCmd(); alterTableCmd != nil {
			if alterTableCmd.Subtype == pg_query.AlterTableType_AT_SetNotNull {
				colName = alterTableCmd.GetName()
				break
			}
		}
	}
	if colName == "" {
		return false
	}
	for _, n := range allNodes {
		alterTable := n.GetAlterTableStmt()
		if alterTable == nil || alterTable.GetRelation().GetRelname() != tableName {
			continue
		}
		for _, c := range alterTable.GetCmds() {
			alterTableCommand := c.GetAlterTableCmd()
			if alterTableCommand == nil || alterTableCommand.Subtype != pg_query.AlterTableType_AT_AddConstraint {
				continue
			}
			if definition := alterTableCommand.GetDef(); definition != nil {
				if constraint := definition.GetConstraint(); constraint.GetContype() == pg_query.ConstrType_CONSTR_CHECK {
					if constraint.SkipValidation && constraint.RawExpr.GetNullTest().GetNulltesttype() == pg_query.NullTestType_IS_NOT_NULL {
						if constraint.RawExpr.GetNullTest().Arg.GetColumnRef().GetFields()[0].GetString_().Sval == colName {
							return false
						}
					}
				}
			}
		}
	}
	return true
}
