package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/infinityworks/fk-infra/util"
	"log"
)

func NewSession(region string) *session.Session {
	newSession, e := session.NewSession(&aws.Config{
		Region: util.String(region),
	})
	util.CheckError(e)
	return newSession
}

func CreateBucket(bucketName, region string) string {
	s3api := s3.New(NewSession(region))
	bucketsOutput, err := s3api.ListBuckets(&s3.ListBucketsInput{})
	util.CheckError(err)

	exists := false
	for _, bucket := range bucketsOutput.Buckets {
		if bucketName == *bucket.Name {
			exists = true
			break
		}
	}

	if !exists {
		_, err := s3api.CreateBucket(&s3.CreateBucketInput{
			Bucket:                    &bucketName,
			CreateBucketConfiguration: bucketLocationConfiguration(region),
		})
		util.CheckError(err)
		_, err = s3api.PutBucketVersioning(&s3.PutBucketVersioningInput{
			Bucket: &bucketName,
			VersioningConfiguration: &s3.VersioningConfiguration{
				Status: util.String(s3.BucketVersioningStatusEnabled),
			},
		})
		util.CheckError(err)
	}
	return bucketName
}

func bucketLocationConfiguration(region string) *s3.CreateBucketConfiguration {
	if region == "us-east-1" {
		return nil
	}
	return &s3.CreateBucketConfiguration{
		LocationConstraint: util.String(region),
	}
}

func CreateKmsKey(keyName, region string) string {
	kmsApi := kms.New(NewSession(region))
	keyAlias := util.String("alias/environment-key-" + keyName)

	if key, err := kmsApi.DescribeKey(&kms.DescribeKeyInput{
		KeyId: keyAlias,
	}); err != nil {
		output, err := kmsApi.CreateKey(&kms.CreateKeyInput{
			Description: util.String("Used to encrypt and decrypt infrastructure secrets for safe storage"),
		})
		util.CheckError(err)
		_, err = kmsApi.CreateAlias(&kms.CreateAliasInput{
			AliasName:   keyAlias,
			TargetKeyId: output.KeyMetadata.Arn,
		})
		util.CheckError(err)
		return *keyAlias
	} else if *key.KeyMetadata.KeyState != kms.KeyStateEnabled {
		log.Panic("KMS key already exists but is not enabled")
	}
	return *keyAlias
}
