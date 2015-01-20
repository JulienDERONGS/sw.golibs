// These tests verify the test running logic.

package gocheck_test

import (
	. "launchpad.net/gocheck"
	"time"
)

var benchmarkS = Suite(&BenchmarkS{})

type BenchmarkS struct{}

func (s *BenchmarkS) TestCountSuite(c *C) {
	suitesRun += 1
}

func (s *BenchmarkS) TestBasicTestTiming(c *C) {
	c.Skip("unreliable results with other processes")
	helper := FixtureHelper{sleepOn: "Test1", sleep: 1000000 * time.Nanosecond}
	output := String{}
	runConf := RunConf{Output: &output, Verbose: true}
	Run(&helper, &runConf)

	expected := "PASS: gocheck_test\\.go:[0-9]+: FixtureHelper\\.Test1\t0\\.001s\n" +
		"PASS: gocheck_test\\.go:[0-9]+: FixtureHelper\\.Test2\t0\\.000s\n"
	c.Assert(output.value, Matches, expected)
}

func (s *BenchmarkS) TestStreamTestTiming(c *C) {
	c.Skip("unreliable results with other processes")
	helper := FixtureHelper{sleepOn: "SetUpSuite", sleep: 1000000 * time.Nanosecond}
	output := String{}
	runConf := RunConf{Output: &output, Stream: true}
	Run(&helper, &runConf)

	expected := "(?s).*\nPASS: gocheck_test\\.go:[0-9]+: FixtureHelper\\.SetUpSuite\t *0\\.001s\n.*"
	c.Assert(output.value, Matches, expected)
}

func (s *BenchmarkS) TestBenchmark(c *C) {
	c.Skip("unreliable results with other processes")
	helper := FixtureHelper{sleep: 100000}
	output := String{}
	runConf := RunConf{
		Output:        &output,
		Benchmark:     true,
		BenchmarkTime: 10000000,
		Filter:        "Benchmark1",
	}
	Run(&helper, &runConf)
	c.Check(helper.calls[0], Equals, "SetUpSuite")
	c.Check(helper.calls[1], Equals, "SetUpTest")
	c.Check(helper.calls[2], Equals, "Benchmark1")
	c.Check(helper.calls[3], Equals, "TearDownTest")
	c.Check(helper.calls[4], Equals, "SetUpTest")
	c.Check(helper.calls[5], Equals, "Benchmark1")
	c.Check(helper.calls[6], Equals, "TearDownTest")
	// ... and more.

	expected := "PASS: gocheck_test\\.go:[0-9]+: FixtureHelper\\.Benchmark1\t *(20|100)\t *[12][0-9]{5,7} ns/op\n"
	c.Assert(output.value, Matches, expected)
}

func (s *BenchmarkS) TestBenchmarkBytes(c *C) {
	c.Skip("unreliable results with other processes")
	helper := FixtureHelper{sleep: 100000}
	output := String{}
	runConf := RunConf{
		Output:        &output,
		Benchmark:     true,
		BenchmarkTime: 10000000,
		Filter:        "Benchmark2",
	}
	Run(&helper, &runConf)

	expected := "PASS: gocheck_test\\.go:[0-9]+: FixtureHelper\\.Benchmark2\t *(20|100)\t *[12][0-9]{5,7} ns/op\t *[1-9]\\.[0-9]{2} MB/s\n"
	c.Assert(output.value, Matches, expected)
}
