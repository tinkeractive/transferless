package configuration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/go-ini/ini"
)

type AWSSecretsManager struct {
	TagKey   string
	TagValue string
}

func (a *AWSSecretsManager) GetConfig() (string, error) {
	secretsClient := secretsmanager.New(session.New(), &aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	// TODO filter only if tag key specified
	secretsFilter := []*secretsmanager.Filter{
		{
			Key:    aws.String("tag-key"),
			Values: aws.StringSlice([]string{a.TagKey}),
		},
		{
			Key:    aws.String("tag-value"),
			Values: aws.StringSlice([]string{a.TagValue}),
		},
	}
	input := &secretsmanager.ListSecretsInput{
		Filters: secretsFilter,
	}
	secrets, err := secretsClient.ListSecrets(input)
	if err != nil {
		return "", err
	}
	iniFile, err := ini.Load([]byte{})
	if err != nil {
		return "", nil
	}
	for _, secretListEntry := range secrets.SecretList {
		input := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(*secretListEntry.Name),
		}
		secretValue, err := secretsClient.GetSecretValue(input)
		if err != nil {
			return "", err
		}
		var secretInterface map[string]interface{}
		err = json.Unmarshal([]byte(*secretValue.SecretString), &secretInterface)
		if err != nil {
			return "", err
		}
		section, err := iniFile.NewSection(*secretListEntry.Name)
		if err != nil {
			return "", err
		}
		for key, value := range secretInterface {
			_, err := section.NewKey(key, value.(string))
			if err != nil {
				return "", err
			}
		}
	}
	var byt []byte
	buf := bytes.NewBuffer(byt)
	_, err = iniFile.WriteTo(buf)
	if err != nil {
		return "", nil
	}
	iniBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}
	return string(iniBytes), err
}

type AWSSystemsManager struct {
	TagKey   string
	TagValue string
}

func (a *AWSSystemsManager) GetConfig() (string, error) {
	ssmClient := ssm.New(session.New(), &aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	// create DescribeParametersInput with filter and DescribeParameters
	parameterStringFilter := []*ssm.ParameterStringFilter{
		{
			Key:    aws.String("tag:" + a.TagKey),
			Values: aws.StringSlice([]string{a.TagValue}),
		},
	}
	input := &ssm.DescribeParametersInput{
		ParameterFilters: parameterStringFilter,
	}
	parameters, err := ssmClient.DescribeParameters(input)
	if err != nil {
		return "", err
	}
	iniFile, err := ini.Load([]byte{})
	if err != nil {
		return "", nil
	}
	for _, parameterMeta := range parameters.Parameters {
		input := &ssm.GetParameterInput{
			Name: aws.String(*parameterMeta.Name),
		}
		output, err := ssmClient.GetParameter(input)
		if err != nil {
			return "", err
		}
		var parameterInterface map[string]interface{}
		err = json.Unmarshal([]byte(*output.Parameter.Value), &parameterInterface)
		if err != nil {
			return "", err
		}
		section, err := iniFile.NewSection(*output.Parameter.Name)
		if err != nil {
			return "", err
		}
		for key, value := range parameterInterface {
			_, err := section.NewKey(key, value.(string))
			if err != nil {
				return "", err
			}
		}
	}
	var byt []byte
	buf := bytes.NewBuffer(byt)
	_, err = iniFile.WriteTo(buf)
	if err != nil {
		return "", nil
	}
	iniBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}
	return string(iniBytes), err
}
