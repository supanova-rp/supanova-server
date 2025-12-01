package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const firebaseAuthURL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword"

type signInRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

type signInResponse struct {
	IDToken      string `json:"idToken"`
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
}

type errorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
			Domain  string `json:"domain"`
			Reason  string `json:"reason"`
		} `json:"errors"`
	} `json:"error"`
}

func main() {
	var (
		apiKey   string
		email    string
		password string
		verbose  bool
	)

	flag.StringVar(&apiKey, "api-key", "", "Firebase Web API Key (required)")
	flag.StringVar(&email, "email", "", "User email address (required)")
	flag.StringVar(&password, "password", "", "User password (required)")
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output including token details")
	flag.Parse()

	if apiKey == "" || email == "" || password == "" {
		fmt.Fprintln(os.Stderr, "Error: api-key, email, and password are required")
		flag.Usage()
		os.Exit(1)
	}

	token, err := getAccessToken(apiKey, email, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("ID Token: %s\n", token.IDToken)
		fmt.Printf("Email: %s\n", token.Email)
		fmt.Printf("User ID: %s\n", token.LocalID)
		fmt.Printf("Expires In: %s seconds\n", token.ExpiresIn)
		fmt.Printf("\nRefresh Token: %s\n", token.RefreshToken)
	} else {
		fmt.Println(token.IDToken)
	}
}

func getAccessToken(apiKey, email, password string) (*signInResponse, error) {
	reqBody := signInRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	parsedURL, err := url.Parse(fmt.Sprintf("%s?key=%s", firebaseAuthURL, apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsedURL.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("authentication failed: %s", errResp.Error.Message)
	}

	var signInResp signInResponse
	if err := json.Unmarshal(body, &signInResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &signInResp, nil
}
