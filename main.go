package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

const bearerToken = ""

// Twitter API response format
type TwitterAPIResponse struct {
	Data []struct {
		ID       string `json:"id"`
		Text     string `json:"text"`
		AuthorID string `json:"author_id"`
	} `json:"data"`
	Includes struct {
		Users []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"users"`
	} `json:"includes"`
}

func getTweets(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query is required"})
		return
	}

	// Twitter API URL with author_id and username expansions
	twitterURL := fmt.Sprintf(
		"https://api.twitter.com/2/tweets/search/recent?query=%s&max_results=10&tweet.fields=text,author_id&expansions=author_id&user.fields=username",
		query,
	)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", bearerToken).
		Get(twitterURL)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var twitterResp TwitterAPIResponse
	if err := json.Unmarshal(resp.Body(), &twitterResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Twitter response"})
		return
	}

	// Map author_id to username
	userMap := make(map[string]string)
	for _, user := range twitterResp.Includes.Users {
		userMap[user.ID] = user.Username
	}

	// Construct frontend-friendly tweet list
	var tweets []gin.H
	for _, tweet := range twitterResp.Data {
		username := userMap[tweet.AuthorID]
		tweets = append(tweets, gin.H{
			"text":     tweet.Text,
			"username": username,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": tweets})
}

func main() {
	r := gin.Default()
	r.GET("/tweets", getTweets)
	r.Run(":8080")
}
