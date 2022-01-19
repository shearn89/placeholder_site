package main

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"os"
)

type CdkStackProps struct {
	awscdk.StackProps
}

const (
	NAMESPACE = "placeholder"
)

var (
	// UUID for certificate, last part of ARN
	certUuid = "5b441288-e615-492b-bd49-ae87981cb4a0"
	// Name for the bucket/zone/apex record
	siteName    = jsii.String("ajshearn.com")
	domainNames = jsii.Strings("ajshearn.com", "www.ajshearn.com")
	// logging bucket
	logBucketName = jsii.String("ajshearn-bucket-access-logs")
)

func wrapName(extension string) *string {
	return jsii.String(fmt.Sprintf("%s-%s", NAMESPACE, extension))
}

func NewCFDistributionStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Cert from ARN as parameter
	createdArn := jsii.String(fmt.Sprintf("arn:aws:acm:us-east-1:%s:certificate/%s", *sprops.Env.Account, certUuid))
	certificate := awscertificatemanager.Certificate_FromCertificateArn(stack, wrapName("certificate"), createdArn)

	// import log bucket
	logBucket := awss3.Bucket_FromBucketName(stack, wrapName("logBucket"), logBucketName)

	// S3 bucket for storing static website
	hostingBucket := awss3.NewBucket(stack, wrapName("bucket"), &awss3.BucketProps{
		BlockPublicAccess:      awss3.BlockPublicAccess_BLOCK_ALL(),
		BucketName:             siteName,
		Encryption:             awss3.BucketEncryption_S3_MANAGED,
		Versioned:              jsii.Bool(true),
		ServerAccessLogsBucket: logBucket,
		ServerAccessLogsPrefix: jsii.String("/s3/placeholder_bucket/"),
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				AbortIncompleteMultipartUploadAfter: awscdk.Duration_Days(jsii.Number(1)),
				Enabled:                             jsii.Bool(true),
				Expiration:                          awscdk.Duration_Days(jsii.Number(30)),
				ExpiredObjectDeleteMarker:           jsii.Bool(true),
				NoncurrentVersionExpiration:         awscdk.Duration_Days(jsii.Number(7)),
			},
		},
	})
	// won't need the bucket without the distribution
	hostingBucket.ApplyRemovalPolicy(awscdk.RemovalPolicy_DESTROY)

	// OAI for cloudfront distribution
	oai := awscloudfront.NewOriginAccessIdentity(stack, wrapName("oai"), &awscloudfront.OriginAccessIdentityProps{
		Comment: jsii.String("used for the placeholder site only"),
	})

	// Cloudfront distribution
	distribution := awscloudfront.NewDistribution(stack, wrapName("distribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.NewS3Origin(hostingBucket, &awscloudfrontorigins.S3OriginProps{
				OriginAccessIdentity: oai,
			}),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
		},
		Certificate:       certificate,
		Comment:           jsii.String("CDN for the placeholder site"),
		DefaultRootObject: jsii.String("index.html"),
		DomainNames:       domainNames,
		LogBucket:         logBucket,
		LogFilePrefix:     jsii.String("/cloudfront/placeholder_site/"),
	})
	// wrap as a target for the r53 record
	target := awsroute53targets.NewCloudFrontTarget(distribution)

	// hosted zone record
	zone := awsroute53.HostedZone_FromLookup(stack, wrapName("hostedZone"), &awsroute53.HostedZoneProviderProps{
		DomainName: siteName,
	})
	for i, domain := range *domainNames {
		awsroute53.NewARecord(stack, wrapName(fmt.Sprintf("%s-%d", *domain, i)), &awsroute53.ARecordProps{
			Zone:       zone,
			RecordName: domain,
			Target:     awsroute53.RecordTarget_FromAlias(target),
		})
	}

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewCFDistributionStack(app, "PlaceholderSiteStack", &CdkStackProps{
		awscdk.StackProps{
			Env: env(),
			Tags: &map[string]*string{
				"cdk":       jsii.String("true"),
				"namespace": jsii.String(NAMESPACE),
			},
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
