package elastic

import (
	"github.com/elastic/go-elasticsearch/v9"
)

func NewClient(addresses []string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}
	return elasticsearch.NewClient(cfg)
}
