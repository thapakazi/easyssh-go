package library

import (
	"html/template"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type HostDefinition struct {
	Host, IpAddr, UserName, PortNumber string
}

const ssh_config_sample = `
{{range .}}
Host {{.Host}}
	hostname {{.IpAddr}}
	User {{.UserName}}
	Port {{.PortNumber}}
{{end}}
`

// FetchIps fetch ips based on tag
// eg tags: map[instance-state-name:[running pending] tag:eksctl.cluster.k8s.io/v1alpha1/cluster-name:[myeks]]
func FetchIps(tags map[string][]string) (*ec2.DescribeInstancesOutput, error) {
	// assuming the keys and env exported via env
	sess, err := session.NewSession()
	CheckError(err)
	ec2obj := ec2.New(sess)

	params := &ec2.DescribeInstancesInput{
		Filters: func() []*ec2.Filter {
			filters := []*ec2.Filter{}
			for k, v := range tags {
				filter := &ec2.Filter{
					Name: aws.String(k),
					Values: func() []*string {
						awsStr := []*string{}
						for _, _v := range v {
							awsStr = append(awsStr, aws.String(_v))
						}
						return awsStr
					}(),
				}
				filters = append(filters, filter)
			}
			return filters
		}(),
	}
	return ec2obj.DescribeInstances(params)
}

func GenerateConfig(username string, port string, response *ec2.DescribeInstancesOutput) {
	// fmt.Println(resp)
	all_hosts := []HostDefinition{}
	for _, reservation := range response.Reservations {
		for _, instance := range reservation.Instances {
			single_host := HostDefinition{UserName: username, PortNumber: port}

			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					value := *tag.Value
					if value == "" {
						value = "NameIsBlank_" + *instance.PrivateIpAddress
					}
					single_host.Host = value
				}
			}

			// if vpc use private ip
			if instance.VpcId != nil {
				single_host.IpAddr = *instance.PrivateIpAddress

				if instance.PublicIpAddress != nil { // if its a publicly facing vpc instance
					single_host.IpAddr = *instance.PublicIpAddress
				}
			} else {
				single_host.IpAddr = *instance.PublicIpAddress
			}

			// append to array
			all_hosts = append(all_hosts, single_host)
		}
	}
	template_sample := template.Must(template.New("sample").Parse(ssh_config_sample))
	CheckError(template_sample.Execute(os.Stdout, all_hosts))
}
