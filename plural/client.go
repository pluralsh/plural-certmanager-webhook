package plural

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

type Config struct {
	Token    string
	Endpoint string
	Cluster  string
	Provider string
}

type Client struct {
	ctx          context.Context
	pluralClient *gqlclient.Client
	config       *Config
}

type DnsRecord struct {
	Type    string
	Name    string
	Records []string
}

func NewConfig(token, endpoint, cluster, provider string) *Config {
	return &Config{
		Token:    token,
		Endpoint: endpoint,
		Cluster:  cluster,
		Provider: provider,
	}
}

func NewClient(conf *Config) *Client {
	base := conf.BaseUrl()
	httpClient := http.Client{
		Transport: &authedTransport{
			key:     conf.Token,
			wrapped: http.DefaultTransport,
		},
	}
	endpoint := base + "/gql"
	return &Client{
		ctx:          context.Background(),
		pluralClient: gqlclient.NewClient(&httpClient, endpoint),
		config:       conf,
	}
}

func (c *Config) BaseUrl() string {
	host := "https://app.plural.sh"
	if c.Endpoint != "" {
		host = fmt.Sprintf("https://%s", c.Endpoint)
	}
	return host
}

func (client *Client) CreateRecord(record *DnsRecord) (*DnsRecord, error) {
	provider := gqlclient.Provider(strings.ToUpper(client.config.Provider))
	cluster := client.config.Cluster
	attr := gqlclient.DNSRecordAttributes{
		Name:    record.Name,
		Type:    gqlclient.DNSRecordType(record.Type),
		Records: []*string{},
	}

	for _, record := range record.Records {
		attr.Records = append(attr.Records, &record)
	}

	resp, err := client.pluralClient.CreateDNSRecord(client.ctx, cluster, provider, attr)
	if err != nil {
		return nil, err
	}

	return &DnsRecord{
		Type:    string(resp.CreateDNSRecord.Type),
		Name:    resp.CreateDNSRecord.Name,
		Records: utils.ConvertStringArrayPointer(resp.CreateDNSRecord.Records),
	}, nil
}

func (client *Client) DeleteRecord(name, ttype string) error {
	if _, err := client.pluralClient.DeleteDNSRecord(client.ctx, name, gqlclient.DNSRecordType(ttype)); err != nil {
		return err
	}

	return nil
}
