package configuration

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/go-ini/ini"
)

type AWS struct {
	TagKey   string
	TagValue string
}

func (a *AWS) GetConfig() (string, error) {
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
	for _, secret := range secrets.SecretList {
		var secretInterface map[string]interface{}
		input := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(*secret.Name),
		}
		secretValue, err := secretsClient.GetSecretValue(input)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal([]byte(*secretValue.SecretString), &secretInterface)
		if err != nil {
			return "", err
		}
		section, err := iniFile.NewSection(*secret.Name)
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
