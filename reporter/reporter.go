package reporter

import (
	"fmt"
	"path/filepath"
	"pgsafemigrate/rules"
	"pgsafemigrate/slicesort"
	"sort"
	"strings"
)

type Reporter interface {
	Print(reports []Report) string
}

type ValidationErrors []rules.ReportedError

func (v ValidationErrors) Empty() bool {
	return len(v) == 0
}

func (v ValidationErrors) Sorted() ValidationErrors {
	sorted := make(ValidationErrors, len(v))
	copy(sorted, v)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Alias() < sorted[j].Alias()
	})
	return sorted
}

type Report struct {
	FilePath string
	Errors   ValidationErrors
}

func NewReport(path string, errors []rules.ReportedError) Report {
	return Report{FilePath: path, Errors: errors}
}

type Printer func([]Report) string

func (f Printer) Print(reports []Report) string {
	return f(reports)
}

type PlainText struct{}

func (p PlainText) Print(reports []Report) string {
	outputBuffer := strings.Builder{}
	// Group by file
	reportsPerFile := make(map[string][]Report)
	for _, r := range reports {
		reportsPerFile[r.FilePath] = append(reportsPerFile[r.FilePath], r)
	}
	var filePaths []string
	for p := range reportsPerFile {
		filePaths = append(filePaths, p)
	}
	for _, path := range slicesort.StringsAscendingOrder(filePaths) {
		fileOutput := strings.Builder{}
		absPath, err := filepath.Abs(path)
		if err != nil {
			path = absPath
		}
		fileOutput.WriteString(fmt.Sprintf("File %s Results:\n", path))
		var hasFailure bool
		for _, r := range reportsPerFile[path] {
			if r.Errors.Empty() {
				continue
			}
			hasFailure = true
			for _, e := range r.Errors.Sorted() {
				fileOutput.WriteString(fmt.Sprintf("\tRule %s violation found for statement:\n\t  %s\n\tExplanation: %s\n\n",
					e.Alias(),
					strings.TrimSpace(e.Statement()),
					e.Documentation()))
			}
		}
		if hasFailure {
			outputBuffer.WriteString(fileOutput.String())
		}
	}
	return strings.TrimSpace(outputBuffer.String())
}
