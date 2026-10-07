package snipe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wzantopulos/snipe-po/config"
)

type Asset struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Serial     string  `json:"serial"`
	Model      Model   `json:"model"`
	MACAddress string  `json:"mac_address"`
	Status     Status  `json:"status"`
	Location   Location `json:"location"`
}

type Model struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Category  Category `json:"category"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Status struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	StatusType string `json:"status_type"`
}

type Location struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type APIResponse struct {
	Total int64   `json:"total"`
	Rows  []Asset `json:"rows"`
}

type AssetResponse struct {
	Asset Asset `json:"asset"`
}

func GetClient() *Client {
	cfg := config.Get()
	return &Client{
		BaseURL: cfg.SnipeIT.URL,
		APIKey:  cfg.SnipeIT.APIKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func (c *Client) doRequest(method, endpoint string) ([]byte, error) {
	url := c.BaseURL + "/api/v1" + endpoint

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) GetAsset(id string) (*Asset, error) {
	data, err := c.doRequest("GET", "/hardware/"+id)
	if err != nil {
		return nil, err
	}

	var resp AssetResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Asset, nil
}

func (c *Client) GetAssets(ids []string) ([]Asset, error) {
	var assets []Asset

	for _, id := range ids {
		asset, err := c.GetAsset(id)
		if err != nil {
			continue
		}
		assets = append(assets, *asset)
	}

	return assets, nil
}

func (c *Client) SearchAssets(query string) ([]Asset, error) {
	data, err := c.doRequest("GET", "/hardware?search="+query)
	if err != nil {
		return nil, err
	}

	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Rows, nil
}

func (c *Client) CreateAsset(name, serial, modelID, statusID string) (*Asset, error) {
	url := c.BaseURL + "/api/v1/hardware"

	body := fmt.Sprintf(`{"name":"%s","serial":"%s","model_id":%s,"status_id":%s}`, name, serial, modelID, statusID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var respData AssetResponse
	if err := json.Unmarshal([]byte(body), &respData); err != nil {
		respBody, _ := io.ReadAll(resp.Body)
		var newResp AssetResponse
		if err := json.Unmarshal(respBody, &newResp); err != nil {
			return nil, err
		}
		return &newResp.Asset, nil
	}

	return &respData.Asset, nil
}
