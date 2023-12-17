package rules

import (
	"errors"
	"fmt"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/pganalyze/pg_query_go/v4/parser"
	migrate "github.com/rubenv/sql-migrate"
	"sort"
	"strings"
)

type Rule interface {
	// Documentation returns a string containing more in-depth details and insights on the rule logic,
	// along with any guidelines about how to fix or handle this operation optimally.
	Documentation() string
	// Alias returns the unique identifier of the rule. The alias is prefixed with a shared category prefix,
	// followed by a code that briefly explains the rule scope.
	Alias() string
	// Process receives a parsed SQL statement that is part of the migration,
	// along with the entire set of migration statements and a flag denoting that
	// the statement is executed within an active transaction or not.
	// Returns true if the rule matches and a warning must be produced for this statement.
	Process(node *pg_query.Node, allStatements []*pg_query.Node, inTransaction bool) bool
}

type RuleSet map[string]Rule

func NewRuleSet() RuleSet {
	return make(RuleSet, 0)
}

func (r RuleSet) Add(rule Rule) {
	if _, ok := r[rule.Alias()]; ok {
		panic(fmt.Sprintf("%s alias already defined", rule.Alias()))
	}
	r[rule.Alias()] = rule
}

func (r RuleSet) Except(aliases ...string) RuleSet {
	if len(aliases) == 0 {
		return r
	}
	filtered := NewRuleSet()
	isExcluded := func(alias string) bool {
		for _, a := range aliases {
			if a == alias {
				return true
			}
		}
		return false
	}
	for alias, rule := range r {
		rule := rule
		if !isExcluded(alias) {
			filtered.Add(rule)
		}
	}
	return filtered
}

func (r RuleSet) Contains(alias string) bool {
	_, ok := r[alias]
	return ok
}

func (r RuleSet) SortedSlice() []Rule {
	var rules []Rule
	for _, rule := range r {
		rules = append(rules, rule)
	}
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Alias() < rules[j].Alias()
	})
	return rules
}

type StatementResult struct {
	Passed    bool
	Direction migrate.MigrationDirection
	Errors    []ReportedError
}

type ReportedError interface {
	Alias() string
	Documentation() string
	Statement() string
}

type Violation struct {
	rule      Rule
	statement string
}

func (e Violation) Alias() string {
	return e.rule.Alias()
}

func (e Violation) Documentation() string {
	return e.rule.Documentation()
}

func (e Violation) Statement() string {
	return e.statement
}

func (r RuleSet) ProcessAll(ctx MigrationContext, statements []string) ([]StatementResult, error) {
	var results []StatementResult
	type task struct {
		rawSQL     string
		statements []*pg_query.RawStmt
	}
	var tasks []task
	var allStatements []*pg_query.Node
	for _, sql := range statements {
		stmts, err := parseStatements(sql)
		if err != nil {
			var parserErr *parser.Error
			if errors.As(err, &parserErr) {
				results = append(results, StatementResult{
					Passed:    false,
					Direction: ctx.Direction,
					Errors: []ReportedError{
						ParseError{message: err.Error(), statement: sql},
					},
				})
				continue
			}
			return nil, err
		}
		tasks = append(tasks, task{
			rawSQL:     sql,
			statements: stmts,
		})
		for _, s := range stmts {
			allStatements = append(allStatements, s.Stmt)
		}
	}
	ctx.AllStatements = allStatements
	for _, task := range tasks {
		ctx := ctx
		ctx.RawSQL = task.rawSQL
		for _, stmt := range task.statements {
			result := r.processSingle(ctx, stmt)
			results = append(results, result)
		}
	}
	return results, nil
}

func (r RuleSet) processSingle(ctx MigrationContext, statement *pg_query.RawStmt) StatementResult {
	result := StatementResult{Passed: true, Direction: ctx.Direction}
	for _, rule := range r.SortedSlice() {
		if rule.Process(statement.Stmt, ctx.AllStatements, ctx.InTransaction) {
			result.Passed = false
			result.Errors = append(result.Errors, Violation{rule: rule, statement: ctx.RawSQL})
		}
	}
	return result
}

func parseStatements(sql string) ([]*pg_query.RawStmt, error) {
	tree, err := pg_query.Parse(sql)
	if err != nil {
		return nil, err
	}
	return tree.GetStmts(), nil
}

var availableRules = NewRuleSet()

func All() RuleSet {
	set := make(RuleSet, len(availableRules))
	for _, rule := range availableRules {
		rule := rule
		set.Add(rule)
	}
	return set
}

func init() {
	availableRules.Add(ColumnComment{})
	availableRules.Add(ColumnSetNotNull{})
	availableRules.Add(CreateIndexNonConcurrently{})
	availableRules.Add(DropIndexNonConcurrently{})
	availableRules.Add(IndexMustBeNamed{})
	availableRules.Add(IndexOperationNotIdempotent{})
	availableRules.Add(NestedTransaction{})
	availableRules.Add(RenameTable{})
	availableRules.Add(RequiredColumn{})
	availableRules.Add(TransactionNotSupportedInConcurrentIndexOperations{})
}

type Category string

const (
	CategoryHighAvailability Category = "high-availability"
	CategoryTransactions     Category = "transactions"
	CategoryMaintainability  Category = "maintainability"
)

func Categories() []Category {
	return []Category{
		CategoryHighAvailability,
		CategoryMaintainability,
		CategoryTransactions,
	}
}

func HighAvailabilityRule(code string) string {
	return fmt.Sprintf("%s-%s", CategoryHighAvailability, code)
}

func TransactionRule(code string) string {
	return fmt.Sprintf("%s-%s", CategoryTransactions, code)
}

func MaintainabilityRule(code string) string {
	return fmt.Sprintf("%s-%s", CategoryMaintainability, code)
}

func CategoryFromAlias(alias string) string {
	for _, c := range Categories() {
		if strings.HasPrefix(alias, string(c)+"-") {
			return string(c)
		}
	}
	return ""
}

type ParseError struct {
	message   string
	statement string
}

func (e ParseError) Alias() string {
	return "parse-error"
}

func (e ParseError) Documentation() string {
	return e.message
}

func (e ParseError) Statement() string {
	return e.statement
}
