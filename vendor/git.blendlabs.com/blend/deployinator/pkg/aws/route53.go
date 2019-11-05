package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	exception "github.com/blend/go-sdk/exception"
)

var (
	// regexpFQDNOctal matches escape code in the format  `\three-digit octal code`
	regexpFQDNOctal = regexp.MustCompile(`\\[0-7]{3,3}`)
)

// DeleteRoute53Entry deletes the route53 entry for the domain
func (a *AWS) DeleteRoute53Entry(fqdn string) error {
	zone, err := a.GetHostedZoneFromAWS(fqdn)
	if err != nil {
		return exception.New(err)
	}
	set, err := a.getResourceRecordSet(fqdn, zone)
	if err != nil {
		return exception.New(err)
	}
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  a.getChangeBatchForRecordSet(route53.ChangeActionDelete, set),
		HostedZoneId: zone.Id,
	}
	send := func() error {
		_, err := a.ensureRoute53().ChangeResourceRecordSets(input)
		return exception.New(err)
	}
	err = a.SendRequestOrTimeout(send)
	if err != nil && ErrorCode(err) != ErrCodeNoRecordFound {
		return exception.New(err)
	}
	return nil
}

// AddRoute53EntryForFQDN adds an entry for a target
func (a *AWS) AddRoute53EntryForFQDN(fqdn, target string, ttl int64) error {
	zone, err := a.GetHostedZoneFromAWS(fqdn)
	if err != nil {
		return exception.New(err)
	}
	var rs *route53.ResourceRecordSet
	if strings.TrimSuffix(aws.StringValue(zone.Name), ".") == fqdn {
		elb, err := a.GetELBV1ByDNSName(target)
		if err != nil {
			return exception.New(err)
		}
		rs = a.getRecordSetForAlias(fqdn, target, aws.StringValue(elb.CanonicalHostedZoneNameID))
	} else {
		rs = a.getRecordSetForCNAME(fqdn, target, ttl)
	}
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  a.getChangeBatchForRecordSet(route53.ChangeActionUpsert, rs),
		HostedZoneId: zone.Id,
	}
	_, err = a.ensureRoute53().ChangeResourceRecordSets(input)
	return exception.New(err)
}

// GetRoute53Entry checks if an entry exists
func (a *AWS) GetRoute53Entry(fqdn string) (*route53.ResourceRecordSet, error) {
	zone, err := a.GetHostedZoneFromAWS(fqdn)
	if err != nil {
		return nil, err
	}
	return a.getResourceRecordSet(fqdn, zone)
}

// DoesRoute53EntryExist checks if an entry exists
func (a *AWS) DoesRoute53EntryExist(fqdn string) (bool, error) {
	_, err := a.GetRoute53Entry(fqdn)
	if err != nil && ErrorCode(err) != ErrCodeNoRecordFound {
		return false, err
	}
	return err == nil, nil
}

// getChangeBatchForRecordSet creates a batch change request for route53
func (a *AWS) getChangeBatchForRecordSet(action string, rs *route53.ResourceRecordSet) *route53.ChangeBatch {
	return &route53.ChangeBatch{
		Changes: []*route53.Change{
			&route53.Change{
				Action:            &action,
				ResourceRecordSet: rs,
			},
		},
	}
}

// GetHostedZonesFromAWS gets a given list of hosted zones from aws
func (a *AWS) GetHostedZonesFromAWS(fqdns []string) ([]*route53.HostedZone, error) {
	zones, err := a.ListHostedZones()
	if err != nil {
		return nil, exception.New(err)
	}
	hostedZones := make([]*route53.HostedZone, len(fqdns))
	for i, fqdn := range fqdns {
		hostedZones[i], err = a.hostedZoneForFQDN(fqdn, zones)
		if err != nil {
			return nil, exception.New(err)
		}
	}
	return hostedZones, nil
}

// GetHostedZoneFromAWS gets a given hosted zone from aws
func (a *AWS) GetHostedZoneFromAWS(fqdn string) (*route53.HostedZone, error) {
	hostedZones, err := a.GetHostedZonesFromAWS([]string{fqdn})
	if err != nil {
		return nil, exception.New(err)
	}
	hostedZone := hostedZones[0]
	if hostedZone == nil {
		return nil, exception.New(fmt.Sprintf("No hosted zone found for `%s`", fqdn))
	}
	return hostedZone, nil
}

// HostedZoneForFQDN finds the best match hosted zone for a fqdn (returns nil if not found)
func (a *AWS) hostedZoneForFQDN(fqdn string, zones []*route53.HostedZone) (*route53.HostedZone, error) {
	var bestMatch *route53.HostedZone
	for _, zone := range zones {
		if zone != nil && zone.Name != nil {
			name := strings.TrimSuffix(*zone.Name, ".") //remove trailing periods
			// check to see if this zone is a parent domain (suffix of) the fqdn, and then if it is a longer match
			// than the current best matching domain
			if strings.HasSuffix(fqdn, name) && (bestMatch == nil || len(name) > len(*bestMatch.Name)) {
				bestMatch = zone
			}
		}
	}
	if bestMatch != nil && aws.StringValue(bestMatch.Name) == "" {
		bestMatch = nil
	}
	return bestMatch, nil
}

// ListHostedZones lists all the hosted zones
func (a *AWS) ListHostedZones() ([]*route53.HostedZone, error) {
	svc := a.ensureRoute53()
	input := route53.ListHostedZonesInput{}
	output, err := svc.ListHostedZones(&input)
	if err != nil {
		return nil, exception.New(err)
	}
	zones := output.HostedZones
	for output.IsTruncated != nil && *output.IsTruncated {
		input.Marker = output.NextMarker
		output, err = svc.ListHostedZones(&input)
		if err != nil {
			return nil, exception.New(err)
		}
		zones = append(zones, output.HostedZones...)
	}
	return zones, nil
}

// getRecordSetForCNAME gets a recordset for a given CNAME
func (a *AWS) getRecordSetForCNAME(fqdn, elb string, ttl int64) *route53.ResourceRecordSet {
	records := []*route53.ResourceRecord{
		&route53.ResourceRecord{
			Value: aws.String(elb),
		},
	}
	return &route53.ResourceRecordSet{
		Type:            aws.String(route53.RRTypeCname),
		TTL:             aws.Int64(ttl),
		Name:            aws.String(fqdn),
		ResourceRecords: records,
	}
}

// getRecordSetForAlias gets a recordset for a given alias (A) entry
func (a *AWS) getRecordSetForAlias(fqdn, elb, hostedZoneID string) *route53.ResourceRecordSet {
	return &route53.ResourceRecordSet{
		Type: aws.String(route53.RRTypeA),
		Name: aws.String(fqdn),
		AliasTarget: &route53.AliasTarget{
			DNSName:              aws.String(fmt.Sprintf("%s%s", ELBDualStackPrefix, elb)),
			EvaluateTargetHealth: aws.Bool(false),
			HostedZoneId:         aws.String(strings.TrimPrefix(hostedZoneID, "/hostedzone/")),
		},
	}
}

// SendRequestOrTimeout sends a route53 request or times out
func (a *AWS) SendRequestOrTimeout(send Action) error {
	ignore := func(err error) error {
		if err != nil {
			switch ErrorCode(err) {
			case ErrCodeAccessDenied: // We cannot edit this zone so stop trying
				if a.log != nil {
					a.log.Error(exception.New(fmt.Sprintf("Access denied when writing to hosted zone. Deployinator does not have write permissions to this zone, please create the Route 53 entry manually")))
				}
				return err
			case ErrCodeInvalidClientTokenID, // AWS credentials might not fully propagate yet
				route53.ErrCodePriorRequestNotComplete: // We just need to wait until aws is done processing our last request
				return nil
			default:
				return err
			}
		}
		return nil
	}
	return a.DoAction(send, ignore, GetReadyInterval(), GetReadyTimeout())
}

func (a *AWS) getResourceRecordSet(fqdn string, zone *route53.HostedZone) (*route53.ResourceRecordSet, error) {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    zone.Id,
		StartRecordName: &fqdn,
	}
	output, err := a.ensureRoute53().ListResourceRecordSets(input)
	if err != nil {
		return nil, exception.New(err)
	}
	for _, set := range output.ResourceRecordSets {
		// the aws api escapes wildcard in recordset name, among other things.
		if set.Name != nil &&
			unescapeFQDN(strings.TrimSuffix(aws.StringValue(set.Name), ".")) == fqdn &&
			(aws.StringValue(set.Type) == route53.RRTypeCname || aws.StringValue(set.Type) == route53.RRTypeA) {
			return set, nil
		}
	}
	return nil, NewError(ErrCodeNoRecordFound, "No record set matches the fqdn in the hosted zone")
}

// Unescape octal string in fqdn http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html
func unescapeFQDN(fqdn string) string {
	return regexpFQDNOctal.ReplaceAllStringFunc(fqdn, func(match string) string {
		ascii, err := strconv.ParseInt(match[1:], 8, 0)
		if err != nil || ascii > 0x7f { // extended ascii is still not supported
			return match
		}
		return string(ascii)
	})
}

// GetReadyTimeout default timeout on route53 operations
func GetReadyTimeout() time.Duration {
	return 20 * time.Minute
}

// GetReadyInterval default interval to poll for route53 operations
func GetReadyInterval() time.Duration {
	return 30 * time.Second
}
