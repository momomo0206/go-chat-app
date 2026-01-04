package topics

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
	"time"
)

type TopicService struct {
	client *http.Client
}

type Topic struct {
	Title       string
	Description string
	URL         string
	Source      string
}

func NewTopicService() *TopicService {
	return &TopicService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func cleanText(text string) string {
	decoded := html.UnescapeString(text)
	return strings.TrimSpace(decoded)
}

func (s *TopicService) FetchDiscordTopic(ctx context.Context) (*Topic, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	channelID := os.Getenv("DISCORD_CHANNEL_ID")

	if token == "" || channelID == "" {
		return nil, fmt.Errorf("Discord config missing in .env")
	}

	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages?limit=1", channelID)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "Bot "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var messages []struct {
		Content string `json:"content"`
		Author  struct {
			Username string `json:"username"`
		} `json:"author"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil || len(messages) == 0 {
		return nil, fmt.Errorf("no discord messages found")
	}

	msg := messages[0]
	return &Topic{
		Title:       cleanText(msg.Content),
		Description: fmt.Sprintf("Latest from Discord By %s", msg.Author.Username),
		URL:         "https://discord.com/channels/@me/" + channelID,
		Source:      "Discord",
	}, nil
}

func (s *TopicService) FetchHackerNewsTop(ctx context.Context) (*Topic, error) {
	resp, err := s.client.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if err != nil {
		return nil, fmt.Errorf("fetch HN top stories: %w", err)
	}
	defer resp.Body.Close()

	var storyIDs []int
	if err := json.NewDecoder(resp.Body).Decode(&storyIDs); err != nil || len(storyIDs) == 0 {
		return nil, fmt.Errorf("decode HN story IDs: %w", err)
	}

	if len(storyIDs) == 0 {
		return nil, fmt.Errorf("no HN stories found")
	}

	storyURL := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", storyIDs[0])
	resp, err = s.client.Get(storyURL)
	if err != nil {
		return nil, fmt.Errorf("fetch HN story: %w", err)
	}
	defer resp.Body.Close()

	var story struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Score int    `json:"score"`
		By    string `json:"by"`
		ID    int    `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&story); err != nil {
		return nil, fmt.Errorf("decode HN story: %w", err)
	}

	if story.URL == "" {
		story.URL = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", story.ID)
	}

	topic := &Topic{
		Title:       cleanText(story.Title),
		Description: fmt.Sprintf("Top HN story with %d points by %s", story.Score, story.By),
		URL:         story.URL,
		Source:      "HackerNews",
	}

	return topic, nil
}

func (s *TopicService) FetchAllTopics(ctx context.Context) ([]Topic, error) {
	topics := make([]Topic, 0, 2)

	dsTopic, err := s.FetchDiscordTopic(ctx)
	if err != nil {
		fmt.Printf("Error fetching Discord topic: %v\n", err)
		topics = append(topics, Topic{
			Title:       "Discord discussion",
			Description: "To see the latest chat",
			URL:         "https://discord.com",
			Source:      "Discord",
		})
	} else {
		topics = append(topics, *dsTopic)
	}

	hnTopic, err := s.FetchHackerNewsTop(ctx)
	if err != nil {
		fmt.Printf("Error fetching HackerNews topic: %v\n", err)
		topics = append(topics, Topic{
			Title:       "Tech News Discussion",
			Description: "Discuss today's tech news",
			URL:         "https://news.ycombinator.com",
			Source:      "HackerNews",
		})
	} else {
		topics = append(topics, *hnTopic)
	}

	return topics, nil
}
