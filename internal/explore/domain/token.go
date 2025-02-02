package domain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type PaginationToken struct {
	Timestamp uint64 `json:"t"`
}

func EncodePaginationToken(timestamp uint64) string {
	token := PaginationToken{
		Timestamp: timestamp,
	}

	data, _ := json.Marshal(token)
	return base64.StdEncoding.EncodeToString(data)
}

func DecodePaginationToken(tokenStr string) (*uint64, error) {
	if tokenStr == "" {
		return nil, nil
	}

	data, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 token: %w", err)
	}

	var token PaginationToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	return &token.Timestamp, nil
}
