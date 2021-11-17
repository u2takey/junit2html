package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/sprig"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	ID          string            `xml:"-"`
	Outcome     string            `xml:"-"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
}

// JUnitSkipMessage contains the reason why a testcase was skipped.
type JUnitSkipMessage struct {
	Message string `xml:"message,attr"`
}

// JUnitProperty represents a key/value pair used to define properties.
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitFailure contains data related to a failed test.
type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type Junit struct {
	Report *JUnitTestSuites
	Title  string
}

func fixSuitesForHtml(report *JUnitTestSuites) {
	for i := range report.Suites {
		for j := range report.Suites[i].TestCases {
			report.Suites[i].TestCases[j].ID = fmt.Sprintf("id_%d_%d", i, j)
			if report.Suites[i].TestCases[j].Failure != nil {
				report.Suites[i].TestCases[j].Outcome = "FAILED"
			} else if report.Suites[i].TestCases[j].SkipMessage != nil {
				report.Suites[i].TestCases[j].Outcome = "SKIPPED"
			} else {
				report.Suites[i].TestCases[j].Outcome = "SUCCESS"
			}
		}
	}
}

func NewJunit(xmlData []byte, title string) (*Junit, error) {
	report := &JUnitTestSuites{}
	singleSuite := &JUnitTestSuite{}
	err := xml.Unmarshal(xmlData, report)
	if err != nil {
		err = xml.Unmarshal(xmlData, singleSuite)
		if err != nil {
			return nil, err
		}
		report.Suites = append(report.Suites, *singleSuite)
	}

	fixSuitesForHtml(report)

	// fmt.Println("loaded junit xml: \n", prettyJson(report))
	return &Junit{Report: report, Title: title}, nil
}

func (j *Junit) Html() ([]byte, error) {
	styleCss, err := templates.ReadFile("templates/style.css")
	if err != nil {
		return nil, err
	}
	html, err := templates.ReadFile("templates/report.html")
	if err != nil {
		return nil, err
	}
	htmlStr := strings.ReplaceAll(string(html), "{{style.css}}", string(styleCss))

	tmpl, err := template.New("html").Funcs(sprig.FuncMap()).Parse(htmlStr)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	err = tmpl.Execute(buf, j)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// debug
func prettyJson(a interface{}) string {
	data, _ := json.MarshalIndent(a, "", " ")
	return string(data)
}

// JUnitReportXML writes a JUnit xml representation of the given report to w
// in the format described at http://windyroad.org/dl/Open%20Source/JUnit.xsd
func JUnitReportXMLModel(report *Report, goVersion, title string) (*Junit, error) {
	var suites = JUnitTestSuites{}
	// convert Report to JUnit test suites
	for _, pkg := range report.Packages {
		pkg.Benchmarks = mergeBenchmarks(pkg.Benchmarks)
		ts := JUnitTestSuite{
			Tests:      len(pkg.Tests) + len(pkg.Benchmarks),
			Failures:   0,
			Time:       formatTime(pkg.Duration),
			Name:       pkg.Name,
			Properties: []JUnitProperty{},
			TestCases:  []JUnitTestCase{},
		}

		classname := pkg.Name
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
			classname = pkg.Name[idx+1:]
		}

		// properties
		if goVersion == "" {
			// if goVersion was not specified as a flag, fall back to version reported by runtime
			goVersion = runtime.Version()
		}
		ts.Properties = append(ts.Properties, JUnitProperty{"go.version", goVersion})
		if pkg.CoveragePct != "" {
			ts.Properties = append(ts.Properties, JUnitProperty{"coverage.statements.pct", pkg.CoveragePct})
		}

		// individual test cases
		for _, test := range pkg.Tests {
			testCase := JUnitTestCase{
				Classname: classname,
				Name:      test.Name,
				Time:      formatTime(test.Duration),
				Failure:   nil,
			}

			if test.Result == FAIL {
				ts.Failures++
				testCase.Failure = &JUnitFailure{
					Message:  "Failed",
					Type:     "",
					Contents: strings.Join(test.Output, "\n"),
				}
			}

			if test.Result == SKIP {
				testCase.SkipMessage = &JUnitSkipMessage{strings.Join(test.Output, "\n")}
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		// individual benchmarks
		for _, benchmark := range pkg.Benchmarks {
			benchmarkCase := JUnitTestCase{
				Classname: classname,
				Name:      benchmark.Name,
				Time:      formatBenchmarkTime(benchmark.Duration),
			}

			ts.TestCases = append(ts.TestCases, benchmarkCase)
		}

		suites.Suites = append(suites.Suites, ts)
	}

	fixSuitesForHtml(&suites)

	return &Junit{
		Report: &suites,
		Title:  title,
	}, nil
}

func mergeBenchmarks(benchmarks []*Benchmark) []*Benchmark {
	var merged []*Benchmark
	benchmap := make(map[string][]*Benchmark)
	for _, bm := range benchmarks {
		if _, ok := benchmap[bm.Name]; !ok {
			merged = append(merged, &Benchmark{Name: bm.Name})
		}
		benchmap[bm.Name] = append(benchmap[bm.Name], bm)
	}

	for _, bm := range merged {
		for _, b := range benchmap[bm.Name] {
			bm.Allocs += b.Allocs
			bm.Bytes += b.Bytes
			bm.Duration += b.Duration
		}
		n := len(benchmap[bm.Name])
		bm.Allocs /= n
		bm.Bytes /= n
		bm.Duration /= time.Duration(n)
	}

	return merged
}

func formatTime(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}

func formatBenchmarkTime(d time.Duration) string {
	return fmt.Sprintf("%.9f", d.Seconds())
}
