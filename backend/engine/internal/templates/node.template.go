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
    
    console.log("Running test cases...\n");
    
    testCases.forEach((test, index) => {
        try {
            console.log(` + "`" + `Test Case ${index + 1}:` + "`" + `);
            console.log(` + "`" + `Input: [${test.inputs.join(", ")}]` + "`" + `);
            
            const result = solve(...test.inputs);
            console.log(` + "`" + `Output: ${result}\n` + "`" + `);
        } catch (error) {
            console.error(` + "`" + `Error in test case ${index + 1}: ${error.message}\n` + "`" + `);
        }
    });
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
