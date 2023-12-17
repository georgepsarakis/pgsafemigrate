package main_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os/exec"
	"path/filepath"
	_ "pgsafemigrate"
	"strings"
	"testing"
)

func TestExecutable_ListRulesCommand(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("go", "run", "./main.go", "list-rules")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, string(output), "Nested transactions are not supported")
}

const expectedFailureOutput = `
Rule high-availability-avoid-non-concurrent-index-creation violation found for statement:
	  CREATE INDEX ON films (created_at);
	Explanation: Non-concurrent index creation will not allow writes while the index is being built.

	Rule maintainability-indexes-name-is-required violation found for statement:
	  CREATE INDEX ON films (created_at);
	Explanation: Indexes should be explicitly named.

	Rule high-availability-avoid-non-concurrent-index-creation violation found for statement:
	  CREATE UNIQUE INDEX title_idx ON films (title) INCLUDE (director, rating);
	Explanation: Non-concurrent index creation will not allow writes while the index is being built.

	Rule transactions-concurrent-index-operation-cannot-be-executed-in-transaction violation found for statement:
	  CREATE INDEX CONCURRENTLY "email_idx" ON "companies" ("email");
	Explanation: Concurrent index operations cannot be executed inside a transaction.

	Rule transactions-no-nested-transactions violation found for statement:
	  BEGIN;
	Explanation: Nested transactions are not supported in PostgreSQL.

	Rule transactions-no-nested-transactions violation found for statement:
	  COMMIT;
	Explanation: Nested transactions are not supported in PostgreSQL.

	Rule high-availability-avoid-table-rename violation found for statement:
	  ALTER TABLE "movies" RENAME TO "movies_old";
	Explanation: Renaming a table can cause errors in previous application versions.

	Rule high-availability-avoid-non-concurrent-index-drop violation found for statement:
	  DROP INDEX IF EXISTS title_idx;
	Explanation: Non-concurrent index drop will not allow writes while the index is being built.
❌ Problems found.
exit status 1`

func TestExecutable_CheckCommand_ProblemsFound(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob("./testdata/sql/20230930091220-add-index.sql")
	require.NoError(t, err)

	cmd := exec.Command("go", "run", "./main.go", "check", strings.Join(matches, " "))

	output, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	assert.ErrorAs(t, err, &exitErr)
	assert.Error(t, &exec.ExitError{Stderr: []byte("exit status 1")})
	assert.Equal(t, 1, exitErr.ExitCode())
	require.NotEmpty(t, output)

	lines := bytes.Split(bytes.TrimSpace(output), []byte("\n"))
	assert.Regexp(t, "File .*/testdata/sql/20230930091220-add-index.sql Results:", string(lines[0]))

	expectedOutputLines := strings.Split(expectedFailureOutput, "\n")
	var expectedOutput []string
	for _, outputLine := range expectedOutputLines {
		if s := strings.TrimSpace(outputLine); s != "" {
			expectedOutput = append(expectedOutput, s)
		}
	}
	var actualOutput []string
	for _, actualLine := range lines[1:] {
		if s := strings.TrimSpace(string(actualLine)); s != "" {
			actualOutput = append(actualOutput, s)
		}
	}
	require.Greater(t, len(expectedOutput), 0)
	require.Greater(t, len(actualOutput), 0)
	require.Equal(t, expectedOutput, actualOutput)
}

func TestExecutable_CheckCommand_NoProblemsFound(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob("./testdata/sql/20231013091220-add-index-success.sql")
	require.NoError(t, err)
	cmd := exec.Command("go", "run", "./main.go", "check", strings.Join(matches, " "))

	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Equal(t, "\n✓ No problems found!\n", string(output))
}
