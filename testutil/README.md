# How to use Fixtures

This is example text we need to make generic:

> Look at ./langutil/golang/golang_test.go. I have used this test to write golang.DocExceptions() to the point I believe it is ready for testing. We need to set up a test fixture containing multiple files
in multiple packages where are can test each of the issues delineated in ./langutil/golang/doc_exception.go lines 8-48. We have a TestFixture type in ./testutil/fixtures.go and you can see how it is
used in ./mcptools/create_file_tool_test.go lines 71-77. Please review those files and explain to me how you would add an exhaustive set of tests to ./langutil/golang/golang_test.go to test all
use-cases identified in ./langutil/golang/doc_exception.go.