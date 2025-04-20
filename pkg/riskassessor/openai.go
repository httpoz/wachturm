// Package riskassessor provides functionality for assessing the risk of package updates.
package riskassessor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/httpoz/watchtower/pkg/packagemanager"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

// OpenAIAssessor uses OpenAI's API to assess the compatibility of package updates.
type OpenAIAssessor struct {
	client                      openai.Client
	responsesNewFn              func(ctx context.Context, params responses.ResponseNewParams) (*responses.Response, error)
	compatibilityScoreSchema    map[string]interface{}
	compatibilityScoreRequestFn func(ctx context.Context, packages []packagemanager.Info) ([]CompatibilityScore, error)
}

// CompatibilityScore represents the compatibility assessment for a package update.
type CompatibilityScore struct {
	Name       string `json:"name" jsonschema_description:"Package name"`
	RiskLevel  string `json:"risk_level" jsonschema_description:"Auto-update safety level based on changelog analysis"`
	RiskReason string `json:"risk_reason" jsonschema_description:"Contextual explanation based on release notes and update impact"`
}

// CompatibilityScoreResponse is the response structure for AI compatibility assessment.
type CompatibilityScoreResponse struct {
	Results []CompatibilityScore `json:"results"`
}

// NewOpenAIAssessor creates a new OpenAIAssessor.
func NewOpenAIAssessor() *OpenAIAssessor {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	assessor := &OpenAIAssessor{
		client: client,
	}

	assessor.responsesNewFn = func(ctx context.Context, params responses.ResponseNewParams) (*responses.Response, error) {
		return client.Responses.New(ctx, params)
	}

	// Generate the schema at initialization time
	assessor.compatibilityScoreSchema = schemaToMap(generateSchema[CompatibilityScoreResponse]())

	// Set up the default compatibility score request function
	assessor.compatibilityScoreRequestFn = assessor.defaultCompatibilityScoreRequest

	return assessor
}

// WithResponsesNewFn sets a custom function for responses.new, primarily for testing.
func (a *OpenAIAssessor) WithResponsesNewFn(fn func(ctx context.Context, params responses.ResponseNewParams) (*responses.Response, error)) *OpenAIAssessor {
	a.responsesNewFn = fn
	return a
}

// WithCompatibilityScoreRequestFn sets a custom function for compatibility score requests, primarily for testing.
func (a *OpenAIAssessor) WithCompatibilityScoreRequestFn(fn func(ctx context.Context, packages []packagemanager.Info) ([]CompatibilityScore, error)) *OpenAIAssessor {
	a.compatibilityScoreRequestFn = fn
	return a
}

// generateSchema generates a JSON schema for the given type.
func generateSchema[T any]() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	return reflector.Reflect(v)
}

// schemaToMap converts a jsonschema.Schema to a map[string]interface{}
func schemaToMap(schema *jsonschema.Schema) map[string]interface{} {
	// Marshal the schema to JSON then unmarshal to map
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		// If there's an error, return an empty map
		return make(map[string]interface{})
	}

	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
		// If there's an error, return an empty map
		return make(map[string]interface{})
	}

	return schemaMap
}

// defaultCompatibilityScoreRequest is the default implementation for compatibility score requests.
func (a *OpenAIAssessor) defaultCompatibilityScoreRequest(ctx context.Context, packages []packagemanager.Info) ([]CompatibilityScore, error) {
	if len(packages) == 0 {
		return nil, nil
	}

	prompt := generatePrompt(packages)

	schemaParam := responses.ResponseFormatTextJSONSchemaConfigParam{
		Name:        "update_compatibility_assessment",
		Description: openai.String("Assess the compatibility of package updates"),
		Schema:      a.compatibilityScoreSchema,
		Strict:      openai.Bool(true),
	}

	// Make API call to OpenAI
	resp, err := a.responsesNewFn(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT4oMini2024_07_18,
		Instructions: param.Opt[string]{
			Value: "You are a Linux system administrator specializing in package management. Analyze package changelog data to determine update compatibility levels. Respond with structured data only.",
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &schemaParam,
			}},
	})

	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	// Parse the response
	var compatibilityResponse CompatibilityScoreResponse
	if err := json.Unmarshal([]byte(resp.OutputText()), &compatibilityResponse); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return compatibilityResponse.Results, nil
}

// ScorePackageRisks assesses the compatibility of the provided packages.
func (a *OpenAIAssessor) ScorePackageRisks(ctx context.Context, packages []packagemanager.Info) ([]packagemanager.Info, error) {
	results, err := a.compatibilityScoreRequestFn(ctx, packages)
	if err != nil {
		return nil, err
	}

	// Create a copy of packages to avoid modifying the input
	scoredPackages := make([]packagemanager.Info, len(packages))
	copy(scoredPackages, packages)

	// Apply compatibility scores
	for i, pkg := range scoredPackages {
		if pkg.Upgrade == nil {
			continue
		}

		for _, r := range results {
			if r.Name == pkg.Name {
				scoredPackages[i].Upgrade.RiskLevel = r.RiskLevel
				scoredPackages[i].Upgrade.RiskReason = r.RiskReason
				break
			}
		}
	}

	return scoredPackages, nil
}

func generatePrompt(packages []packagemanager.Info) string {
	var prompt string
	prompt = "Analyze these Ubuntu package updates and provide a compatibility assessment:\n\n"
	for _, pkg := range packages {
		if pkg.Upgrade == nil || pkg.Upgrade.Changelog == nil {
			continue
		}

		prompt += fmt.Sprintf("Package: %s\nUpdate: %s -> %s\nChangelog:\n%s\n\n",
			pkg.Name,
			pkg.Version,
			pkg.Upgrade.NewVersion,
			pkg.Upgrade.Changelog.Raw,
		)
	}
	return prompt
}
