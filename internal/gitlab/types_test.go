package gitlab

import "testing"

func TestMergeRequest_IsDraft(t *testing.T) {
	tests := []struct {
		name     string
		mr       MergeRequest
		expected bool
	}{
		{
			name: "draft field is true",
			mr: MergeRequest{
				Title: "Regular title",
				Draft: true,
			},
			expected: true,
		},
		{
			name: "title starts with Draft:",
			mr: MergeRequest{
				Title: "Draft: Implement new feature",
				Draft: false,
			},
			expected: true,
		},
		{
			name: "title starts with draft: (lowercase)",
			mr: MergeRequest{
				Title: "draft: Fix bug",
				Draft: false,
			},
			expected: true,
		},
		{
			name: "title starts with WIP:",
			mr: MergeRequest{
				Title: "WIP: Work in progress",
				Draft: false,
			},
			expected: true,
		},
		{
			name: "title starts with wip: (lowercase)",
			mr: MergeRequest{
				Title: "wip: Testing",
				Draft: false,
			},
			expected: true,
		},
		{
			name: "title with extra spaces",
			mr: MergeRequest{
				Title: "  Draft: Feature",
				Draft: false,
			},
			expected: true,
		},
		{
			name: "non-draft title",
			mr: MergeRequest{
				Title: "Merge op-stage into op-rc",
				Draft: false,
			},
			expected: false,
		},
		{
			name: "title contains but does not start with Draft",
			mr: MergeRequest{
				Title: "This is a Draft version",
				Draft: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mr.IsDraft()
			if result != tt.expected {
				t.Errorf("IsDraft() = %v, expected %v for title: %q, draft field: %v",
					result, tt.expected, tt.mr.Title, tt.mr.Draft)
			}
		})
	}
}
