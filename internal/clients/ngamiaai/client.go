// Package ngamiaai is intentionally named after the upstream gateway, not the CV product.
package ngamiaai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)
type Client struct { BaseURL, APIKey string; HTTP *http.Client }
type Message struct { Role string `json:"role"`; Content string `json:"content"` }
type CompletionRequest struct { Model string `json:"model"`; Messages []Message `json:"messages"`; Stream bool `json:"stream"` }
type CompletionResponse struct { Choices []struct { Message Message `json:"message"` } `json:"choices"`; Usage struct { PromptTokens int `json:"prompt_tokens"`; CompletionTokens int `json:"completion_tokens"`; TotalTokens int `json:"total_tokens"` } `json:"usage"` }
func New(baseURL,apiKey string)*Client{return &Client{BaseURL:strings.TrimSuffix(baseURL,"/"),APIKey:apiKey,HTTP:&http.Client{Timeout:45*time.Second}}}
func (c *Client) Chat(ctx context.Context,request CompletionRequest)(CompletionResponse,error){if c.APIKey==""{return CompletionResponse{},fmt.Errorf("NGAMIA_AI_API_KEY is not configured")};raw,err:=json.Marshal(request);if err!=nil{return CompletionResponse{},err};req,err:=http.NewRequestWithContext(ctx,http.MethodPost,c.BaseURL+"/chat/completions",bytes.NewReader(raw));if err!=nil{return CompletionResponse{},err};req.Header.Set("Authorization","Bearer "+c.APIKey);req.Header.Set("Content-Type","application/json");res,err:=c.HTTP.Do(req);if err!=nil{return CompletionResponse{},err};defer res.Body.Close();if res.StatusCode<200||res.StatusCode>299{return CompletionResponse{},fmt.Errorf("Ngamia AI request failed with status %d",res.StatusCode)};var parsed CompletionResponse;if err:=json.NewDecoder(res.Body).Decode(&parsed);err!=nil{return CompletionResponse{},err};return parsed,nil}
