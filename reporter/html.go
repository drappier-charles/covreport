package reporter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func (dirs GoDirs) Report(output string) error {
	tmpl := template.Must(template.New("html").Parse(templateHTML))
	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("can't create %q: %v", output, err)
	}
	defer file.Close()

	var buf strings.Builder
	dirs.root().write(&buf, "", "root", "root")

	return tmpl.Execute(file, buf.String())
}

func (dir *GoDir) write(w io.Writer, links string, id string, basename string) error {
	links += fmt.Sprintf(`<a href="#%s">%s</a>`, id, basename)

	filesHTML := openHeadingHTML(id, links, "files", dir.numStmtCovered, dir.numStmt)
	for _, subDir := range dir.subDirs {
		subDirBasename := filepath.Base(subDir.dirname)
		subDirID := id + "-" + subDirBasename
		subDir.write(w, links, subDirID, subDirBasename)
		filesHTML += fileItemHTML(subDirID, subDirBasename, subDir.numStmtCovered, subDir.numStmt)
	}
	for _, file := range dir.files {
		fileBasename := filepath.Base(file.filename)
		id := id + "-" + fileBasename
		file.write(w, links, id, fileBasename)
		filesHTML += fileItemHTML(id, fileBasename, file.numStmtCovered, file.numStmt)
	}

	filesHTML += "</div></div>"
	w.Write([]byte(filesHTML))
	return nil
}

func (file *GoFile) write(w io.Writer, links, id string, basename string) error {
	src, err := os.ReadFile(file.filename)
	if err != nil {
		return fmt.Errorf("can't read %q: %v", file.filename, err)
	}
	links += fmt.Sprintf(`<span>%s</span>`, basename)
	numProfileBlock := len(file.profile)
	idxProfile := 0
	dst := bufio.NewWriter(w)

	fmt.Fprint(dst, openHeadingHTML(id, links, "codes", file.numStmtCovered, file.numStmt))

	for idx, code := range strings.Split(string(src), "\n") {
		lineNumber := idx + 1
		var count *int

		if idxProfile < numProfileBlock {
			profile := file.profile[idxProfile]
			if profile.EndLine < lineNumber {
				idxProfile++
				if idxProfile < numProfileBlock {
					count = &file.profile[idxProfile].Count
				}
			} else {
				count = &file.profile[idxProfile].Count
			}
		}

		code = strings.ReplaceAll(code, ">", "&gt;")
		code = strings.ReplaceAll(code, "<", "&lt;")
		code = strings.ReplaceAll(code, "&", "&amp;")
		code = strings.ReplaceAll(code, "\t", "    ")

		if count == nil {
			fmt.Fprintf(dst, "<div class=\"line-number\">%d</div><div class=\"covered-count\"></div><pre class=\"line\">%s</pre>\n", lineNumber, code)
		} else if *count == 0 {
			fmt.Fprintf(dst, "<div class=\"line-number\">%d</div><div class=\"covered-count uncovered\"></div><pre class=\"line uncovered\">%s</pre>\n", lineNumber, code)
		} else {
			fmt.Fprintf(dst, "<div class=\"line-number\">%d</div><div class=\"covered-count covered\">%dx</div><pre class=\"line covered\">%s</pre>\n", lineNumber, *count, code)
		}
	}
	fmt.Fprint(dst, "</div></div>")
	dst.Flush()

	return nil
}

func openHeadingHTML(id, links, subclass string, numStmtCovered, numStmt int) string {
	var percent float64
	if numStmt == 0 {
		percent = 0
	} else {
		percent = float64(numStmtCovered) / float64(numStmt) * 100
	}
	return fmt.Sprintf(`
	<div id="%s" class="view file" style="display:none">
		<div class="links">
			%s
		</div>
		<div class="summary">
			<div class="percent">%.1f%%</div>
			<div class="label">Statements</div>
			<div class="stmts">%d/%d</div>
		</div>
		<div class="%s">
	`, id, links, percent, numStmtCovered, numStmt, subclass)
}

func fileItemHTML(id, baseName string, numStmtCovered, numStmt int) string {
	var percent float64
	var class string

	if numStmt == 0 {
		percent = 0
	} else {
		percent = float64(numStmtCovered) / float64(numStmt) * 100
		if percent > 70 {
			class = "safe"
		} else if percent < 40 {
			class = "danger"
		} else {
			class = "warning"
		}
	}

	return fmt.Sprintf(`
		<a class="wrapper %s" href="#%s">
			<div class="subpath">%s</div>
			<div class="progress"><progress value="%.1f" max="100"></progress></div>
			<div class="percent">%.1f%%</div>
			<div class="statements">%d/%d</div>
		</a>
		`,
		class, id, baseName, percent, percent, numStmtCovered, numStmt)
}

const templateHTML = `
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<title>Go Coverage Report</title>
		<style>
			body {
				font-family: Menlo, monospace;
			}
			a {
				text-decoration: none;
				color: blue;
				&:visited {
					color: blue;
				}
			}
			progress {
				-webkit-appearance: none;
				-moz-appearance: none;        
				appearance: none;
			}
			.view .links {
				font-size: 1.2em;
				padding: 1rem;
			}
			.view .links a {
				&::after {
					content: "/";
					color: black;
				}
			}
			.view .links span {
				color: black;
				font-weight: bold;
			}
			.view .summary {
				padding: 0 1rem 2rem 1rem;
			}
			.view .summary {
				display: flex;
				justify-content: flex-start;
				align-items: center;
				gap: 1rem;
			}
			.view .summary .label {
				opacity: 0.8;
			}
			.view .summary .stmts {
				border: 1px solid gray;
				border-radius: 4px;
				background-color: lightgray;
				padding: 2px 4px;
			}
			.codes {
				display: grid;
				grid-template-columns: 3em 3em auto;
				margin-bottom: 3rem;
			}
			.codes .wrapper {
				display: contents;
			}
			.codes .line-number, .codes .covered-count {
				font-size: 0.5em;
				display: flex;
				justify-content: flex-end;
				align-items: center;
				margin-right: 4px;
				padding-right: 4px;
			}
			.codes .line-number {
				opacity: 0.8;
			}
			.codes .covered-count {
				background-color: lightgray;
			}
			.codes pre {
				margin: 0;
				font-size: 1em;
				line-height: 1.5em;
				height: 1.5em;
			}
			.codes .uncovered {
				background-color: rgba(255, 0, 0, 0.2);
			}
			.codes .covered-count.covered {
				background-color: rgba(0, 255, 0, 0.2);
				color: green;
			}
			.files {
				margin: 0 1rem 3rem 1rem;
				display: grid;
				grid-template-columns: auto max-content max-content max-content;
				gap: 1px;
			}
			.files .wrapper > * {
				padding: 8px 1rem;
				&:not(:first-child) {
					color: black;
				}
			}
			.files .wrapper.danger > * {
				background-color: rgba(255, 0, 0, 0.2);
			}
			.files .wrapper.safe > * {
				background-color: rgba(0, 255, 0, 0.2);
			}
			.files .wrapper.warning > * {
				background-color: rgba(255, 255, 0, 0.2);
			}
			progress::-webkit-progress-value {
				background-color: green;
			}
			progress::-moz-progress-value {
				background-color: green;
			}
			progress::-progress-value {
				background-color: green;
			}
			.files .wrapper {
				display: contents;
				text-align: right;
				border: 1px solid lightgray;
			}
			.files .wrapper .subpath {
				text-align: left;
			}
		</style>
	</head>
	<body>
		{{.}}
	</body>
	<script>
	window.renderView = () => {
		for (const view of document.getElementsByClassName('view')) {
			view.style.display = 'none';
		};
		const id = window.location.hash ? window.location.hash.substring(1) : 'root';
		document.getElementById(id).style.display = 'block';
	};
	window.addEventListener('hashchange', () => {
		window.renderView();
	});
	window.renderView();
	</script>
</html>
`
