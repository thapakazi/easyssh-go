package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/namsral/flag"

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

func check_err(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func main() {

	// taking inputs
	var username, port string
	flag.StringVar(&username, "username", "ubuntu", "ssh username to login as")
	flag.StringVar(&port, "port", "22", "ssh port to knock")

	var custom_vpc_string string
	flag.StringVar(&custom_vpc_string, "custom_vpc_string", "_vpc", "custom vpc string/tag to append")

	// custom msg at begin/end of generated config
	var start_msg, end_msg string
	flag.StringVar(&start_msg, "start_msg", "##START OF GENERATED SSH CONFIG##", "custom comment msg at start of generated ssh config")
	flag.StringVar(&end_msg, "end_msg", "##END OF GENERATED SSH CONFIG##", "custom comment end msg")

	flag.Parse()

	// assuming the keys and env exported via env
	sess, err := session.NewSession()
	check_err(err)
	ec2obj := ec2.New(sess)

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{ // Required
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"), // Required
					aws.String("pending"), // Required
					// More values...
				},
			},
		},
	}
	resp, err := ec2obj.DescribeInstances(params)
	check_err(err)
	// fmt.Println(resp)
	all_hosts := []HostDefinition{}
	for idx := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			single_host := HostDefinition{UserName: username, PortNumber: port}
			for _, tag := range inst.Tags {
				if *tag.Key == "Name" {
					value := *tag.Value
					if value == "" {
						value = "NameIsBlank_" + *inst.PrivateIpAddress
					}
					single_host.Host = value
				}
			}

			// if vpc use private ip
			if inst.VpcId != nil {
				single_host.IpAddr = *inst.PrivateIpAddress

				// don't know the easy alternative 😢
				var vpc_buffer bytes.Buffer
				vpc_buffer.WriteString(single_host.Host)
				vpc_buffer.WriteString(custom_vpc_string)
				single_host.Host = vpc_buffer.String()

				if inst.PublicIpAddress != nil { // if its a publicly facing vpc instance
					single_host.IpAddr = *inst.PublicIpAddress
				}
			} else {
				single_host.IpAddr = *inst.PublicIpAddress
			}

			// fmt.Println(single_host)
			// append to array
			all_hosts = append(all_hosts, single_host)
		}
	}

	template_sample := template.Must(template.New("sample").Parse(ssh_config_sample))
	fmt.Println(start_msg)
	check_err(template_sample.Execute(os.Stdout, all_hosts))
	fmt.Println(end_msg)
}
