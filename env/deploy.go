package env

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
	"github.com/nekrassov01/ignoresync"
	"github.com/nekrassov01/ignoresync/manager"
)

// Deploy creates and waits for the CloudFormation stack deployment.
func (o *Deployer) Deploy(ctx context.Context, state *manager.State) error {
	if err := o.createStack(ctx, state); err != nil {
		return NewEnvError(err)
	}
	err := o.waitStack(ctx, state)
	if err == nil {
		return nil
	}
	if reason := o.getFailReason(ctx, state); reason != "" {
		return NewEnvError(fmt.Errorf("%s: %w", reason, err))
	}
	return err
}

// CheckDeployed checks if the CloudFormation stack exists.
// In situations where you want to check whether the stack has been deployed,
// the state may not yet exist, so you cannot use the state's region information.
func (o *Deployer) CheckDeployed(ctx context.Context, state *manager.State) (bool, error) {
	opt := func(opt *cloudformation.Options) {
		opt.Region = state.Region
	}
	in := &cloudformation.DescribeStacksInput{
		StackName: aws.String(ignoresync.CanonicalName),
	}
	if _, err := o.cfn.DescribeStacks(ctx, in, opt); err != nil {
		if e, ok := errors.AsType[smithy.APIError](err); ok {
			if e.ErrorCode() == "StackNotFoundException" || strings.Contains(e.ErrorMessage(), "does not exist") {
				return false, nil
			}
		}
		return false, NewEnvError(fmt.Errorf("failed to check stack present: %w", err))
	}
	return true, nil
}

// createStack creates the CloudFormation stack.
func (o *Deployer) createStack(ctx context.Context, state *manager.State) error {
	opt := func(opt *cloudformation.Options) {
		opt.Region = state.Region
	}
	in := &cloudformation.CreateStackInput{
		StackName:    aws.String(ignoresync.CanonicalName),
		TemplateBody: aws.String(template),
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam,
		},
		Parameters: []types.Parameter{
			{
				ParameterKey:   aws.String("BucketName"),
				ParameterValue: aws.String(state.Bucket.ARN.Resource),
			},
			{
				ParameterKey:   aws.String("SSEKeyAlias"),
				ParameterValue: aws.String(state.SSEKey.ARN.Resource),
			},
			{
				ParameterKey:   aws.String("CSEKeyAlias"),
				ParameterValue: aws.String(state.CSEKey.ARN.Resource),
			},
		},
	}
	if _, err := o.cfn.CreateStack(ctx, in, opt); err != nil {
		return fmt.Errorf("failed to create stack: %w", err)
	}
	return nil
}

// waitStack waits for the CloudFormation stack creation to complete.
func (o *Deployer) waitStack(ctx context.Context, state *manager.State) error {
	opt := func(o *cloudformation.StackCreateCompleteWaiterOptions) {
		o.LogWaitAttempts = true
		o.ClientOptions = []func(o *cloudformation.Options){
			func(o *cloudformation.Options) {
				o.Region = state.Region
			},
		}
	}
	in := &cloudformation.DescribeStacksInput{
		StackName: aws.String(ignoresync.CanonicalName),
	}
	waiter := o.cfn.NewStackCreateCompleteWaiter(opt)
	if err := waiter.Wait(ctx, in, stackMaxWaitDur); err != nil {
		return fmt.Errorf("failed to wait stack creation: %w", err)
	}
	return nil
}

// getFailReason retrieves the failure reason from the CloudFormation stack events.
func (o *Deployer) getFailReason(ctx context.Context, state *manager.State) string {
	opt := func(opt *cloudformation.Options) {
		opt.Region = state.Region
	}
	in := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(ignoresync.CanonicalName),
	}
	var token *string
	var reason string
	for {
		in.NextToken = token
		out, err := o.cfn.DescribeStackEvents(ctx, in, opt)
		if err != nil {
			return ""
		}
		events := out.StackEvents
		for i := len(events) - 1; i >= 0; i-- {
			event := events[i]
			isFailed := strings.HasSuffix(string(event.ResourceStatus), "_FAILED")
			if event.ResourceStatus != "" && isFailed {
				r := aws.ToString(event.ResourceStatusReason)
				if r != "" {
					reason = r
				} else {
					reason = ""
				}
			}
		}
		token = out.NextToken
		if token == nil {
			return reason
		}
	}
}
