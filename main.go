package main

import (
	"os"

	"github.com/namsral/flag"
	"github.com/thapakazi/easyssh-go/library"
)

var username, port, start_msg, end_msg string

func init() {

	// taking inputs

	flag.StringVar(&username, "username", "ubuntu", "ssh username to login as")
	flag.StringVar(&port, "port", "22", "ssh port to knock")

	// custom msg at begin/end of generated config
	flag.StringVar(&start_msg, "start_msg", "##START OF GENERATED SSH CONFIG##", "custom comment msg at start of generated ssh config")
	flag.StringVar(&end_msg, "end_msg", "##END OF GENERATED SSH CONFIG##", "custom comment end msg")
}

func main() {
	flag.Parse()
	var tags = make(map[string][]string)
	// tags["tag:eksctl.cluster.k8s.io/v1alpha1/cluster-name"] = []string{"myeks"}
	tags["instance-state-name"] = []string{"running", "pending"}

	response, err := library.FetchIps(tags)
	library.CheckError(err)
	library.GenerateConfig(username, port, response, os.Stdout)
}
