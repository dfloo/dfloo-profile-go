package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const systemPrompt = `You are a professional resume editor. Tailor specific sections of a resume to make it more relevant to a job application without fabricating experience.

Rules:
- You MAY rewrite: summary, skills list (reorder and filter for relevance), and bullet points under each experience entry.
- You MUST NOT change: employer names, job titles, dates, locations, education, contact info, or template settings.
- Do not add bullet points not grounded in the original text.
- The "experience" array must contain exactly the same number of entries as the input, in the same order.
- Skills should be a filtered and reordered subset of the originals, with the most relevant ones first.
- Return ONLY valid JSON with no markdown fences:
  {
    "summary": "...",
    "skills": ["..."],
    "experience": [{ "bulletPoints": ["..."] }, ...]
  }`

type TailorRequest struct {
	Summary        string
	Skills         []string
	Experience     []ExperienceInput
	Company        string
	Role           string
	JobDescription string
}

type ExperienceInput struct {
	Employer     string
	Title        string
	StartDate    string
	EndDate      string
	BulletPoints []string
}

type TailorResponse struct {
	Summary    string              `json:"summary"`
	Skills     []string            `json:"skills"`
	Experience []ExperienceOutput  `json:"experience"`
}

type ExperienceOutput struct {
	BulletPoints []string `json:"bulletPoints"`
}

type Client interface {
	TailorResume(ctx context.Context, req TailorRequest) (*TailorResponse, error)
}

type AnthropicClient struct {
	client anthropic.Client
}

func New() *AnthropicClient {
	return &AnthropicClient{
		client: anthropic.NewClient(option.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY"))),
	}
}

func (c *AnthropicClient) TailorResume(ctx context.Context, req TailorRequest) (*TailorResponse, error) {
	userMsg := buildUserMessage(req)

	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{
				Text:         systemPrompt,
				CacheControl: anthropic.NewCacheControlEphemeralParam(),
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic API error: %w", err)
	}

	if len(msg.Content) == 0 {
		return nil, fmt.Errorf("anthropic returned empty response")
	}

	text := strings.TrimSpace(msg.Content[0].AsText().Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result TailorResponse
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse claude response: %w", err)
	}

	return &result, nil
}

func buildUserMessage(req TailorRequest) string {
	var sb strings.Builder

	sb.WriteString("Job Application Details:\n")
	sb.WriteString(fmt.Sprintf("Company: %s\n", req.Company))
	sb.WriteString(fmt.Sprintf("Role: %s\n", req.Role))
	sb.WriteString(fmt.Sprintf("Job Description:\n%s\n\n", req.JobDescription))

	sb.WriteString("Current Resume Fields to Tailor:\n\n")
	sb.WriteString(fmt.Sprintf("Summary:\n%s\n\n", req.Summary))

	sb.WriteString("Skills:\n")
	for _, skill := range req.Skills {
		sb.WriteString(fmt.Sprintf("- %s\n", skill))
	}
	sb.WriteString("\n")

	sb.WriteString("Experience:\n")
	for _, exp := range req.Experience {
		sb.WriteString(fmt.Sprintf("Employer: %s | Title: %s | %s – %s\n", exp.Employer, exp.Title, exp.StartDate, exp.EndDate))
		sb.WriteString("Current bullet points:\n")
		for _, bullet := range exp.BulletPoints {
			sb.WriteString(fmt.Sprintf("- %s\n", bullet))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Return the tailored JSON now.")
	return sb.String()
}
