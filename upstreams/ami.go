package upstreams

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

type AMI struct {
	UpstreamBase `mapstructure:",squash"`
	Match        string
	Owner        string
}

func (upstream AMI) LatestVersion() (string, error) {
	log.Debugf("Using AMI upstream")
	// Create a new session based on shared / env credentials
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := ec2.New(s)

	// Generate filters based on configuration
	var filters []*ec2.Filter
	filters = append(filters, &ec2.Filter{
		Name:   aws.String("name"),
		Values: []*string{aws.String(upstream.Match)},
	})

	input := &ec2.DescribeImagesInput{
		Owners:  []*string{aws.String(upstream.Owner)},
		Filters: filters,
	}

	// Do the actual API call
	result, err := svc.DescribeImages(input)
	if err != nil {
		return "", err
	}
	images := result.Images

	// Sort images by creation time, so we can return the latest
	sort.Slice(images, func(i, j int) bool { return *images[i].CreationDate > *images[j].CreationDate })
	log.Debugf("Matched AMIs:\n%s", images)

	if len(images) < 1 {
		return "", fmt.Errorf("No AMI found for upstream %s", upstream)
	}

	latestImage := images[0]
	log.Debugf("Latest AMI:\n%s", latestImage)

	return *latestImage.ImageId, nil
}
