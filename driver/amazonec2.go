package rancher

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	tag             = "Rancher Cloud"
	vpcCidnBlock    = "10.0.0.0/16"
	subnetCidnBlock = "10.0.0.0/24"
)

func (d *Driver) setupAmazon() error {
	client := d.AmazonEC2Driver.GetClient().(*ec2.EC2)

	vpcID, err := findOrCreateVpc(client)
	if err != nil {
		return err
	}
	d.AmazonEC2Driver.VpcId = vpcID

	subnetID, availabilityZone, err := findOrCreateSubnet(client, vpcID)
	if err != nil {
		return err
	}
	d.AmazonEC2Driver.SubnetId = subnetID
	d.AmazonEC2Driver.Zone = string(availabilityZone[len(availabilityZone)-1])

	if _, err = client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{
			&[]string{vpcID}[0],
			&[]string{subnetID}[0],
		},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   &[]string{"Name"}[0],
				Value: &[]string{tag}[0],
			},
		},
	}); err != nil {
		return err
	}

	securityGroupID, err := findOrCreateSecurityGroup(client, vpcID)
	if err != nil {
		return err
	}
	d.AmazonEC2Driver.SecurityGroupId = securityGroupID

	return nil
}

func findOrCreateVpc(client *ec2.EC2) (string, error) {
	describeVpcsOutput, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: &[]string{"tag:Name"}[0],
				Values: []*string{
					&[]string{tag}[0],
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	vpcs := describeVpcsOutput.Vpcs

	if len(vpcs) == 0 {
		createVpcOutput, err := client.CreateVpc(&ec2.CreateVpcInput{
			CidrBlock: &[]string{vpcCidnBlock}[0],
		})
		if err != nil {
			return "", err
		}
		return *createVpcOutput.Vpc.VpcId, nil
	} else if len(vpcs) == 1 {
		return *vpcs[0].VpcId, nil
	}

	return "", fmt.Errorf("Multiple VPCs named %s found", tag)
}

func findOrCreateSubnet(client *ec2.EC2, vpcID string) (string, string, error) {
	describeSubnetsOutput, err := client.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: &[]string{"tag:Name"}[0],
				Values: []*string{
					&[]string{tag}[0],
				},
			},
		},
	})
	if err != nil {
		return "", "", err
	}

	subnets := describeSubnetsOutput.Subnets

	if len(subnets) == 0 {
		createSubnetOutput, err := client.CreateSubnet(&ec2.CreateSubnetInput{
			CidrBlock: &[]string{subnetCidnBlock}[0],
			VpcId:     &[]string{vpcID}[0],
		})
		if err != nil {
			return "", "", err
		}
		return *createSubnetOutput.Subnet.SubnetId, *createSubnetOutput.Subnet.AvailabilityZone, nil
	} else if len(subnets) == 1 {
		return *subnets[0].SubnetId, *subnets[0].AvailabilityZone, nil
	}

	return "", "", fmt.Errorf("Multiple subnets named %s found", tag)
}

func findOrCreateSecurityGroup(client *ec2.EC2, vpcID string) (string, error) {
	describeSecurityGroupsOutput, err := client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: &[]string{"group-name"}[0],
				Values: []*string{
					&[]string{tag}[0],
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	securityGroups := describeSecurityGroupsOutput.SecurityGroups

	if len(securityGroups) == 0 {
		createSecurityGroupOutput, err := client.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
			Description: &[]string{tag}[0],
			GroupName:   &[]string{tag}[0],
			VpcId:       &[]string{vpcID}[0],
		})
		if err != nil {
			return "", err
		}
		fmt.Println(createSecurityGroupOutput)
		return *createSecurityGroupOutput.GroupId, nil
	} else if len(securityGroups) == 1 {
		return *securityGroups[0].GroupId, nil
	}

	return "", fmt.Errorf("Multiple security groups named %s found", tag)
}
