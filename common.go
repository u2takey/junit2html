package main

type ReportInterface interface {
	addReport(file string)
	failures() []string
	skips() []string
}

type Report interface {
	suites() []Suite
}

type Suite interface {
	failed() []string
	skipped() []string
}

type ReportContainer struct {
	reports []Report
}

func (r *ReportContainer) failures() (found []string) {
	for _, r := range r.reports {
		for _, s := range r.suites() {
			found = append(found, s.failed()...)
		}
	}
	return
}

func (r *ReportContainer) skips() (found []string) {
	for _, r := range r.reports {
		for _, s := range r.suites() {
			found = append(found, s.skipped()...)
		}
	}
	return
}
