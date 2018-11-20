package tfit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hcl/hcl/printer"
)

type chanItem struct {
	obj interface{}
	err error
}

func handleError(err error) error {
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		case "NoSuchBucketPolicy":
			return nil
		case "NoSuchWebsiteConfiguration":
			return nil
		case "NoSuchLifecycleConfiguration":
			return nil
		case "ReplicationConfigurationNotFoundError":
			return nil
		case "ServerSideEncryptionConfigurationNotFoundError":
			return nil
		case "NoSuchCORSConfiguration":
			return nil
		default:
			return err
		}
	}

	return err
}

func GetCredentials(c *Config) *credentials.Credentials {
	providers := []credentials.Provider{
		&credentials.StaticProvider{Value: credentials.Value{
			AccessKeyID:     c.AccessKey,
			SecretAccessKey: c.SecretKey,
			SessionToken:    c.Token,
		}},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{
			Filename: c.CredsFile,
			Profile:  c.Profile,
		},
	}

	return credentials.NewChainCredentials(providers)
}

func getZoneId(src *string) *string {
	if strings.Contains(aws.StringValue(src), "/") {
		tokens := strings.Split(aws.StringValue(src), "/")
		return aws.String(tokens[len(tokens)-1])
	}

	return src
}

func quote(src string) string {

	if strings.HasPrefix(src, "\"") && strings.HasSuffix(src, "\"") {
		return src
	}

	return fmt.Sprintf("\"%s\"", src)
}

func joinStringSlice(sep string, src []string) string {
	for k, v := range src {
		src[k] = quote(v)
	}

	return strings.Join(src, sep)
}

// HCLFmt read HCL formatted text from io.Reader
// and do pretty HCL format then write to io.Writer
func HCLFmt(r io.Reader, w io.Writer) error {
	src := bytes.NewBuffer(nil)
	_, err := src.ReadFrom(r)
	if err != nil {
		return err
	}

	hclFile, err := parser.Parse(src.Bytes())
	if err != nil {
		return err
	}

	return printer.Fprint(w, hclFile.Node)
}

func prettyJSON(src *string) string {
	var data map[string]interface{}
	buf := bytes.NewBufferString(aws.StringValue(src))
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		panic(err)
	}
	b, err := json.MarshalIndent(data, "", " ")

	if err != nil {
		panic(err)
	}

	return string(b)
}

func unEscapeHTML(src *string) (string, error) {
	return url.QueryUnescape(aws.StringValue(src))
}

func renderHCL(w io.Writer, Tmpl string, funcMap template.FuncMap, target interface{}) error {
	t := template.New("").Funcs(funcMap)
	t, err := t.Parse(Tmpl)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, target)
	if err != nil {
		return err
	}

	return HCLFmt(buf, w)
}

func renderTerraformImportCmd(Output io.Writer, Tmpl string, funcMap template.FuncMap, target interface{}) error {
	t := template.New("").Funcs(funcMap)
	t, err := t.Parse(Tmpl)
	if err != nil {
		return err
	}

	return t.Execute(Output, target)
}
