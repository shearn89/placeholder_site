package main

import (
	"github.com/aws/jsii-runtime-go"
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	assertions "github.com/aws/aws-cdk-go/awscdk/v2/assertions"
)

func testEnv() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String("000000000000"),
		Region:  jsii.String("eu-west-1"),
	}
}

func TestCdkStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewCFDistributionStack(app, "MyStack", &CdkStackProps{
		awscdk.StackProps{
			Env: testEnv(),
			Tags: &map[string]*string{
				"cdk":       jsii.String("true"),
				"namespace": jsii.String(NAMESPACE),
			},
		},
	})

	// THEN
	assertions.Template_FromStack(stack)

	// TODO: cdk-nag
	// awscdk.Aspects_Of(app).Add( // some aspect here)

	// template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
	// 	"BucketName": "ajshearn.com",
	// })
}
