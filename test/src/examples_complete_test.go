package test

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// Test the Terraform module in examples/complete using Terratest.
func TestExamplesComplete(t *testing.T) {
	t.Parallel()

	rand.Seed(time.Now().UnixNano())

	attributes := []string{strconv.Itoa(rand.Intn(100000))}

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../examples/complete",
		Upgrade:      true,
		// Variables to pass to our Terraform code using -var-file options
		VarFiles: []string{"fixtures.us-east-2.tfvars"},
		Vars: map[string]interface{}{
			"attributes": attributes,
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of an output variable
	cfArn := terraform.Output(t, terraformOptions, "cf_arn")
	// Verify we're getting back the outputs we expect
	assert.Contains(t, cfArn, "arn:aws:cloudfront::")

	// Run `terraform output` to get the value of an output variable
	s3BucketName := terraform.Output(t, terraformOptions, "s3_bucket")
	expectedS3BucketName := "eg-test-cloudfront-s3-cdn-" + attributes[0] + "-origin"
	// Verify we're getting back the outputs we expect
	assert.Equal(t, expectedS3BucketName, s3BucketName)

	policyString := terraform.Output(t, terraformOptions, "s3_bucket_policy")
	assert.NotPanics(t, func() { getTestResource(policyString) }, "Could not parse S3 Bucket Policy")
	defer func() { recover() }()
	assert.Equal(t, `arn:aws:s3:::`+expectedS3BucketName+`/testprefix/*`, getTestResource(policyString),
		"Templating of var.additional_bucket_policy failed")
}

func getTestResource(jsonString string) string {
	var js interface{}

	err := json.Unmarshal([]byte(jsonString), &js)
	if err != nil {
		return ""
	}
	policy := js.(map[string]interface{})
	statements := policy["Statement"].([]interface{})
	for _, statement := range statements {
		s := statement.(map[string]interface{})
		if s["Sid"].(string) != "TemplateTest" {
			continue
		}
		return s["Resource"].(string)
	}

	return ""
}
