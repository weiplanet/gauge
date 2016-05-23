// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package run_failed

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

type MySuite1 struct{}

var _ = Suite(&MySuite{})
var _ = Suite(&MySuite1{})

func (s *MySuite1) SetUpTest(c *C) {
	failedMeta = newFailedMetaData()
}

func (s *MySuite1) TearDownTest(c *C) {
	p, _ := filepath.Abs("_testdata")
	os.RemoveAll(filepath.Join(p, dotGauge))
}

func (s *MySuite) TestIfFailedFileIsCreated(c *C) {
	p, _ := filepath.Abs("_testdata")
	config.ProjectRoot = p
	failedInfo := "hello world"

	writeFailedMeta(failedInfo)

	file := filepath.Join(config.ProjectRoot, dotGauge, failedFile)
	c.Assert(common.FileExists(file), Equals, true)
	expected := failedInfo

	content, _ := ioutil.ReadFile(file)

	c.Assert(string(content), Equals, expected)
	os.RemoveAll(filepath.Join(config.ProjectRoot, dotGauge))
}

func (s *MySuite1) TestListenToSpecFailure(c *C) {
	p, _ := filepath.Abs("_testdata")
	config.ProjectRoot = p
	event.InitRegistry()
	specRel := filepath.Join("specs", "example.spec")
	specAbs := filepath.Join(p, specRel)
	failedMeta.Flags["-tags"] = "tag1"
	failedMeta.Flags["-verbose"] = "true"

	ListenFailedScenarios()
	sr := &result.SpecResult{IsFailed: true, ProtoSpec: &gauge_messages.ProtoSpec{FileName: &specAbs}, FailedScenarioIndices: []int{2}}
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, &result.SuiteResult{SpecResults: []*result.SpecResult{sr}}, 0))

	contents, _ := common.ReadFileContents(filepath.Join(p, dotGauge, failedFile))
	pathInFile := specRel
	if util.IsWindows() {
		pathInFile = `specs\\example.spec`
	}
	expected := `{
	"Flags": {
		"-tags": "tag1",
		"-verbose": "true"
	},
	"FailedScenarios": [
		` + `"` + pathInFile + `:2"
	]
}`
	c.Assert(contents, Equals, expected)
}

func (s *MySuite1) TestListenToSpecPass(c *C) {
	p, _ := filepath.Abs("_testdata")
	config.ProjectRoot = p
	event.InitRegistry()
	specRel := filepath.Join("specs", "example.spec")
	specAbs := filepath.Join(p, specRel)
	failedMeta.Flags["-tags"] = "tag1 & tag2"
	failedMeta.Flags["-verbose"] = "true"
	failedMeta.Flags["-env"] = "chrome"

	ListenFailedScenarios()
	sr := &result.SpecResult{IsFailed: false, ProtoSpec: &gauge_messages.ProtoSpec{FileName: &specAbs}}
	event.Notify(event.NewExecutionEvent(event.SuiteEnd, nil, &result.SuiteResult{SpecResults: []*result.SpecResult{sr}}, 0))

	contents, _ := common.ReadFileContents(filepath.Join(p, dotGauge, failedFile))
	expected := `{
	"Flags": {
		"-env": "chrome",
		"-tags": "tag1 \u0026 tag2",
		"-verbose": "true"
	},
	"FailedScenarios": []
}`
	c.Assert(contents, Equals, expected)
}

func (s *MySuite) TestGetFailedMetadata(c *C) {
	p, _ := filepath.Abs("_testdata")
	config.ProjectRoot = p
	spec1Rel := filepath.Join("specs", "example1.spec")
	spec1Abs := filepath.Join(p, spec1Rel)
	sr1 := &result.SpecResult{IsFailed: true, ProtoSpec: &gauge_messages.ProtoSpec{FileName: &spec1Abs}, FailedScenarioIndices: []int{2, 6}}

	meta := getFailedMetadata([]*result.SpecResult{sr1})

	c.Assert(len(meta.FailedScenarios), Equals, 2)
	c.Assert(meta.FailedScenarios[0], Equals, spec1Rel+":2")
	c.Assert(meta.FailedScenarios[1], Equals, spec1Rel+":6")
}
