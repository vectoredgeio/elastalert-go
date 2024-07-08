package queries

import (
	"context"
	"crypto/tls"
	"elastalert-go/rules"

	// "encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	// "github.com/opensearch-project/opensearch-go/opensearchutil"
)

func NewClient(host string, port int, username, password string) (*opensearch.Client, error) {
	address := fmt.Sprintf("https://%s:%d", host, port)

	// Create OpenSearch client with configured transport and credentials
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			TLSHandshakeTimeout: 30 * time.Second,
		},
		Addresses: []string{address},
		Username:  username,
		Password:  password,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating OpenSearch client: %v", err)
	}

	return client, nil
}

func Query(client *opensearch.Client, index string, query *opensearchapi.SearchRequest, size int,rule rules.Rule) (*opensearchapi.Response, error) {

	
	searchResponse, err := query.Do(context.Background(), client)
	if err != nil {
		fmt.Println("failed to search document ", err)
		os.Exit(1)
	}
	fmt.Println("Searching for a document for rule",rule.GetName())
	fmt.Println(searchResponse)
	defer searchResponse.Body.Close()

	return searchResponse, nil
}
