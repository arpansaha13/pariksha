package templates

import (
	"bytes"
	"text/template"

	"pariksha/common/pkg/constants"
)

const nodeTemplate = `
// User code starts
{{.Code}}
// User code ends

// Test runner
function runTests() {
	const testCases = {{.TestCases}};
	let results = [];
	
	console.log("Running test cases...\n");
	
	for (let i = 0; i < testCases.length; i++) {
		const test = testCases[i];
		const startTime = Date.now();
		
		console.log("` + TEST_CASE_START + `");
		try {
			const result = solve(...test.inputs);
			const endTime = Date.now();
			const stringifiedResult = JSON.stringify(result)
			const expected = test.expectedOutput;
			
			results.push({
				output: stringifiedResult,
				executionTime: endTime - startTime,
				match: stringifiedResult === expected,
				inputs: test.inputs,
				expectedOutput: test.expectedOutput
			});
		} catch (error) {
			const endTime = Date.now();
			results.push({
				error: error.message,
				executionTime: endTime - startTime,
				match: false,
				inputs: test.inputs,
				expectedOutput: test.expectedOutput
			});
		}
		console.log("` + TEST_CASE_END + `");
	}

	console.log("` + RESULTS_START + `");
	console.log(JSON.stringify(results));
	console.log("` + RESULTS_END + `");
}

runTests();
`

type NodeTemplateData struct {
	Code      string
	TestCases string
}

// GenerateNodeScript creates a Node.js script with the user's code and test cases
var GenerateNodeScript TemplateFunc = func(code string, testCases string) (string, error) {
	tmpl, err := template.New(constants.LangNode).Parse(nodeTemplate)
	if err != nil {
		return "", err
	}

	data := NodeTemplateData{
		Code:      code,
		TestCases: testCases,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
