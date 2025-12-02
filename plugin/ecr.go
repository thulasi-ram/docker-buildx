// Source: https://github.com/drone-plugins/drone-docker/tree/939591f01828eceae54f5768dc7ce08ad0ad0bba/cmd/drone-ecr
package plugin

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

const DefaultRegion = "us-east-1"

var (
	repo       string
	assumeRole string
	externalID string
	ecr_login  Login
	aws_region string
)

func (p *Plugin) EcrInit() {
	// create a standalone Login object to account for single repo and multi-repo case
	if len(p.Settings.Logins) >= 1 {
		for _, login := range p.Settings.Logins {
			if strings.Contains(login.Registry, "amazonaws.com") {
				ecr_login = login
				aws_region = login.Aws_region

				// filter repo containing ecr registry
				substrings := make([]string, 0)
				for _, repo := range p.Settings.Build.Repo {
					substrings = append(substrings, strings.Split(repo, ",")...)
				}
				filtered := make([]string, 0)
				for _, s := range substrings {
					if strings.Contains(s, "amazonaws.com") {
						filtered = append(filtered, s)
					}
				}

				// Join the filtered substrings into a comma-separated string
				repo = strings.Join(filtered, ",")

				// set the region
				if aws_region == "" {
					aws_region = DefaultRegion
				}

				_ = os.Setenv("AWS_REGION", aws_region)
				_ = os.Setenv("AWS_ACCESS_KEY_ID", ecr_login.Aws_access_key_id)
				_ = os.Setenv("AWS_SECRET_ACCESS_KEY", ecr_login.Aws_secret_access_key)

			}
		}
	} else {
		ecr_login.Aws_access_key_id = p.Settings.AwsAccessKeyId
		ecr_login.Aws_secret_access_key = p.Settings.AwsSecretAccessKey
		ecr_login.Registry = p.Settings.DefaultLogin.Registry
		aws_region = p.Settings.AwsRegion
		repo = p.Settings.Build.Repo[0]

		// set the region
		if aws_region == "" {
			aws_region = DefaultRegion
		}

		_ = os.Setenv("AWS_REGION", p.Settings.AwsRegion)
		_ = os.Setenv("AWS_ACCESS_KEY_ID", p.Settings.AwsAccessKeyId)
		_ = os.Setenv("AWS_SECRET_ACCESS_KEY", p.Settings.AwsSecretAccessKey)
	}
	// here the env vars are used for authentication
	sess, err := session.NewSession(&aws.Config{Region: &aws_region})
	if err != nil {
		log.Fatalf("error creating aws session: %v", err)
	}

	svc := getECRClient(sess, assumeRole, externalID)
	username, password, registry, err := getAuthInfo(svc)
	if err != nil {
		log.Fatalf("error getting ECR auth: %v", err)
	}
	log.Printf("ECR auth info: %s %s %s", username, password, registry)

	if registry != ecr_login.Registry {
		// This is because ecr.GetAuthorizationToken deprecated passing registry id
		log.Printf("ECR registry does not match login registry. Expected %s, got %s. Proceeding with specified registry.", ecr_login.Registry, registry)
		registry = ecr_login.Registry
	}

	if !strings.HasPrefix(repo, registry) {
		repo = fmt.Sprintf("%s/%s", registry, repo)
	}

	if p.Settings.EcrCreateRepository {
		err = ensureRepoExists(svc, trimHostname(repo, registry), p.Settings.EcrScanOnPush)
		if err != nil {
			log.Fatalf("error creating ECR repo: %v", err)
		}
		err = updateImageScannningConfig(svc, trimHostname(repo, registry), p.Settings.EcrScanOnPush)
		if err != nil {
			log.Fatalf("error updating scan on push for ECR repo: %v", err)
		}
	}

	if p.Settings.EcrLifecyclePolicy != "" {
		p, err := os.ReadFile(p.Settings.EcrLifecyclePolicy)
		if err != nil {
			log.Fatal(err)
		}
		if err := uploadLifeCyclePolicy(svc, string(p), trimHostname(repo, registry)); err != nil {
			log.Fatalf("error uploading ECR lifecycle policy: %v", err)
		}
	}

	if p.Settings.EcrRepositoryPolicy != "" {
		p, err := os.ReadFile(p.Settings.EcrRepositoryPolicy)
		if err != nil {
			log.Fatal(err)
		}
		if err := uploadRepositoryPolicy(svc, string(p), trimHostname(repo, registry)); err != nil {
			log.Fatalf("error uploading ECR repository policy. %v", err)
		}
	}

	// set Username and Password for all Login which contain an AWS key
	if len(p.Settings.Logins) >= 1 {
		for i, login := range p.Settings.Logins {
			if login.Aws_secret_access_key != "" && login.Aws_access_key_id != "" {
				p.Settings.Logins[i].Username = username
				p.Settings.Logins[i].Password = password
				p.Settings.Logins[i].Registry = registry
			}
		}
	} else {
		p.Settings.DefaultLogin.Username = username
		p.Settings.DefaultLogin.Password = password
		p.Settings.DefaultLogin.Registry = registry
	}
}

func trimHostname(repo, registry string) string {
	repo = strings.TrimPrefix(repo, registry)
	repo = strings.TrimLeft(repo, "/")
	return repo
}

func ensureRepoExists(svc *ecr.ECR, name string, scanOnPush bool) (err error) {
	input := &ecr.CreateRepositoryInput{}
	input.SetRepositoryName(name)
	input.SetImageScanningConfiguration(&ecr.ImageScanningConfiguration{ScanOnPush: &scanOnPush})
	_, err = svc.CreateRepository(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == ecr.ErrCodeRepositoryAlreadyExistsException {
			// eat it, we skip checking for existing to save two requests
			err = nil
		}
	}

	return err
}

func updateImageScannningConfig(svc *ecr.ECR, name string, scanOnPush bool) (err error) {
	input := &ecr.PutImageScanningConfigurationInput{}
	input.SetRepositoryName(name)
	input.SetImageScanningConfiguration(&ecr.ImageScanningConfiguration{ScanOnPush: &scanOnPush})
	_, err = svc.PutImageScanningConfiguration(input)

	return err
}

func uploadLifeCyclePolicy(svc *ecr.ECR, lifecyclePolicy, name string) (err error) {
	input := &ecr.PutLifecyclePolicyInput{}
	input.SetLifecyclePolicyText(lifecyclePolicy)
	input.SetRepositoryName(name)
	_, err = svc.PutLifecyclePolicy(input)

	return err
}

func uploadRepositoryPolicy(svc *ecr.ECR, repositoryPolicy, name string) (err error) {
	input := &ecr.SetRepositoryPolicyInput{}
	input.SetPolicyText(repositoryPolicy)
	input.SetRepositoryName(name)
	_, err = svc.SetRepositoryPolicy(input)

	return err
}

func getAuthInfo(svc *ecr.ECR) (username, password, registry string, err error) {
	var result *ecr.GetAuthorizationTokenOutput
	var decoded []byte

	result, err = svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return username, password, registry, err
	}

	auth := result.AuthorizationData[0]
	token := *auth.AuthorizationToken
	decoded, err = base64.StdEncoding.DecodeString(token)
	if err != nil {
		return username, password, registry, err
	}

	registry = strings.TrimPrefix(*auth.ProxyEndpoint, "https://")
	creds := strings.Split(string(decoded), ":")
	username = creds[0]
	password = creds[1]
	return username, password, registry, err
}

func getECRClient(sess *session.Session, role, externalId string) *ecr.ECR {
	if role == "" {
		return ecr.New(sess)
	}
	if externalId != "" {
		return ecr.New(sess, &aws.Config{
			Credentials: stscreds.NewCredentials(sess, role, func(p *stscreds.AssumeRoleProvider) {
				p.ExternalID = &externalId
			}),
		})
	} else {
		return ecr.New(sess, &aws.Config{
			Credentials: stscreds.NewCredentials(sess, role),
		})
	}
}
