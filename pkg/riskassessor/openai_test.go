package riskassessor

import (
	"context"
	"testing"

	"github.com/httpoz/watchtower/pkg/packagemanager"
	"github.com/openai/openai-go/responses"
)

// TestScorePackageRisks tests the ScorePackageRisks function with a mock OpenAI client
func TestScorePackageRisks(t *testing.T) {
	// Create a mock response for package risk scoring
	mockRiskResponse := `{
		"results": [
			{
				"name": "apt",
				"risk_level": "low",
				"risk_reason": "Security update with no behavior changes"
			},
			{
				"name": "bash",
				"risk_level": "medium",
				"risk_reason": "Configuration changes mentioned in changelog"
			}
		]
	}`

	// Create mock chat completions function
	mockResponsesCreate := func(ctx context.Context, params responses.ResponseNewParams) (*responses.Response, error) {
		return &responses.Response{
			Output: []responses.ResponseOutputItemUnion{
				{
					Content: []responses.ResponseOutputMessageContentUnion{
						{
							Type: "output_text",
							Text: mockRiskResponse,
						},
					},
				},
			},
		}, nil
	}

	// Create mock risk score request function
	mockRiskScoreRequest := func(ctx context.Context, packages []packagemanager.Info) ([]CompatibilityScore, error) {
		return []CompatibilityScore{
			{
				Name:       "apt",
				RiskLevel:  "low",
				RiskReason: "Security update with no behavior changes",
			},
			{
				Name:       "bash",
				RiskLevel:  "medium",
				RiskReason: "Configuration changes mentioned in changelog",
			},
		}, nil
	}

	// Create the assessor with our mock
	assessor := NewOpenAIAssessor().
		WithResponsesNewFn(mockResponsesCreate).
		WithCompatibilityScoreRequestFn(mockRiskScoreRequest)

	// Test data
	testPackages := []packagemanager.Info{
		{
			Name:         "apt",
			Version:      "2.4.8",
			Architecture: "amd64",
			Upgrade: &packagemanager.UpgradeInfo{
				NewVersion: "2.4.9",
				HasUpgrade: true,
			},
		},
		{
			Name:         "bash",
			Version:      "5.1-6ubuntu1",
			Architecture: "amd64",
			Upgrade: &packagemanager.UpgradeInfo{
				NewVersion: "5.1-6ubuntu1.1",
				HasUpgrade: true,
			},
		},
	}

	// Call the function with our test data
	result, err := assessor.ScorePackageRisks(context.Background(), testPackages)

	// Verify the results
	if err != nil {
		t.Errorf("ScorePackageRisks returned error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(result))
	}

	// Check that the risk levels were properly assigned
	if result[0].Upgrade.RiskLevel != "low" {
		t.Errorf("Expected apt risk level to be low, got %s", result[0].Upgrade.RiskLevel)
	}

	if result[1].Upgrade.RiskLevel != "medium" {
		t.Errorf("Expected bash risk level to be medium, got %s", result[1].Upgrade.RiskLevel)
	}
}
