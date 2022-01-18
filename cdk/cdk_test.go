package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	assertions "github.com/aws/aws-cdk-go/awscdk/v2/assertions"
)

func TestCdkStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCFDistributionStack(app, "MyStack", nil)

	// THEN
	assertions.Template_FromStack(stack)

	// template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
	// 	"BucketName": "ajshearn.com",
	// })
}
