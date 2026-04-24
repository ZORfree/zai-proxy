package model

import "testing"

// ===== ParseModelName =====

func TestParseModelName_Plain(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7")
	if base != "glm-4.7" {
		t.Errorf("base = %q, want %q", base, "glm-4.7")
	}
	if thinking || search || tools {
		t.Errorf("flags = (%v, %v, %v), want all false", thinking, search, tools)
	}
}

func TestParseModelName_Lowercase(t *testing.T) {
	base, thinking, _, _ := ParseModelName("glm-4.7-thinking")
	if base != "glm-4.7" {
		t.Errorf("base = %q, want %q", base, "glm-4.7")
	}
	if !thinking {
		t.Error("thinking should be true")
	}
}

func TestParseModelName_Thinking(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7-thinking")
	if base != "glm-4.7" {
		t.Errorf("base = %q", base)
	}
	if !thinking {
		t.Error("thinking should be true")
	}
	if search || tools {
		t.Error("search and tools should be false")
	}
}

func TestParseModelName_Search(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7-search")
	if base != "glm-4.7" {
		t.Errorf("base = %q", base)
	}
	if !search {
		t.Error("search should be true")
	}
	if thinking || tools {
		t.Error("thinking and tools should be false")
	}
}

func TestParseModelName_Tools(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7-tools")
	if base != "glm-4.7" {
		t.Errorf("base = %q", base)
	}
	if !tools {
		t.Error("tools should be true")
	}
	if thinking || search {
		t.Error("thinking and search should be false")
	}
}

func TestParseModelName_ThinkingSearch(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7-thinking-search")
	if base != "glm-4.7" {
		t.Errorf("base = %q", base)
	}
	if !thinking || !search {
		t.Error("thinking and search should both be true")
	}
	if tools {
		t.Error("tools should be false")
	}
}

func TestParseModelName_ToolsThinking(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7-tools-thinking")
	if base != "glm-4.7" {
		t.Errorf("base = %q", base)
	}
	if !tools || !thinking {
		t.Error("tools and thinking should both be true")
	}
	if search {
		t.Error("search should be false")
	}
}

func TestParseModelName_ToolsSearch(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7-tools-search")
	if base != "glm-4.7" {
		t.Errorf("base = %q", base)
	}
	if !tools || !search {
		t.Error("tools and search should both be true")
	}
	if thinking {
		t.Error("thinking should be false")
	}
}

func TestParseModelName_AllTags(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7-tools-thinking-search")
	if base != "glm-4.7" {
		t.Errorf("base = %q", base)
	}
	if !thinking || !search || !tools {
		t.Errorf("all flags should be true, got (%v, %v, %v)", thinking, search, tools)
	}
}

func TestParseModelName_ReverseOrder(t *testing.T) {
	base, thinking, search, tools := ParseModelName("GLM-4.7-search-thinking-tools")
	if base != "glm-4.7" {
		t.Errorf("base = %q", base)
	}
	if !thinking || !search || !tools {
		t.Errorf("all flags should be true, got (%v, %v, %v)", thinking, search, tools)
	}
}

// ===== GLM-5 =====

func TestParseModelName_GLM5(t *testing.T) {
	base, thinking, _, _ := ParseModelName("glm-5")
	if base != "glm-5" {
		t.Errorf("base = %q, want %q", base, "glm-5")
	}
	if thinking {
		t.Error("thinking should be false")
	}
}

func TestGetTargetModel_GLM5(t *testing.T) {
	target := GetTargetModel("glm-5")
	if target != "glm-5" {
		t.Errorf("GetTargetModel(glm-5) = %q, want %q", target, "glm-5")
	}
}

func TestGetTargetModel_GLM5Thinking(t *testing.T) {
	target := GetTargetModel("glm-5-thinking")
	if target != "glm-5" {
		t.Errorf("GetTargetModel(glm-5-thinking) = %q, want %q", target, "glm-5")
	}
}

// ===== GLM-5.1 =====

func TestParseModelName_GLM51(t *testing.T) {
	base, thinking, _, _ := ParseModelName("glm-5.1")
	if base != "glm-5.1" {
		t.Errorf("base = %q, want %q", base, "glm-5.1")
	}
	if thinking {
		t.Error("thinking should be false")
	}
}

func TestGetTargetModel_GLM51(t *testing.T) {
	target := GetTargetModel("glm-5.1")
	if target != "GLM-5.1" {
		t.Errorf("GetTargetModel(glm-5.1) = %q, want %q", target, "GLM-5.1")
	}
}

func TestGetTargetModel_GLM51Thinking(t *testing.T) {
	target := GetTargetModel("glm-5.1-thinking")
	if target != "GLM-5.1" {
		t.Errorf("GetTargetModel(glm-5.1-thinking) = %q, want %q", target, "GLM-5.1")
	}
}

// ===== IsToolsModel =====

func TestIsToolsModel_True(t *testing.T) {
	tests := []string{
		"glm-4.7-tools",
		"GLM-4.7-tools-thinking",
		"glm-4.7-tools-search",
		"GLM-4.7-thinking-tools",
		"glm-4.5-tools",
		"glm-5-tools",
	}
	for _, m := range tests {
		if !IsToolsModel(m) {
			t.Errorf("IsToolsModel(%q) = false, want true", m)
		}
	}
}

// ===== ResolveClaudeModel =====

func TestResolveClaudeModel_Opus(t *testing.T) {
	resolved, thinking := ResolveClaudeModel("claude-opus-4-6", false)
	if resolved != "glm-5.1-thinking-tools" {
		t.Errorf("ResolveClaudeModel(claude-opus-4-6) = %q, want %q", resolved, "glm-5.1-thinking-tools")
	}
	if !thinking {
		t.Error("thinking should be true for opus")
	}
}

func TestIsToolsModel_False(t *testing.T) {
	tests := []string{
		"glm-4.7",
		"GLM-4.7-thinking",
		"glm-4.7-search",
		"GLM-4.7-thinking-search",
	}
	for _, m := range tests {
		if IsToolsModel(m) {
			t.Errorf("IsToolsModel(%q) = true, want false", m)
		}
	}
}

// ===== IsThinkingModel / IsSearchModel =====

func TestIsThinkingModel_WithTools(t *testing.T) {
	if !IsThinkingModel("glm-4.7-tools-thinking") {
		t.Error("IsThinkingModel should be true for glm-4.7-tools-thinking")
	}
	if IsThinkingModel("glm-4.7-tools") {
		t.Error("IsThinkingModel should be false for glm-4.7-tools")
	}
}

func TestIsSearchModel_WithTools(t *testing.T) {
	if !IsSearchModel("glm-4.7-tools-search") {
		t.Error("IsSearchModel should be true for glm-4.7-tools-search")
	}
	if IsSearchModel("glm-4.7-tools") {
		t.Error("IsSearchModel should be false for glm-4.7-tools")
	}
}

// ===== GetTargetModel =====

func TestGetTargetModel_WithTools(t *testing.T) {
	target := GetTargetModel("glm-4.7-tools")
	if target != "glm-4.7" {
		t.Errorf("GetTargetModel(glm-4.7-tools) = %q, want %q", target, "glm-4.7")
	}
}

func TestGetTargetModel_WithToolsThinking(t *testing.T) {
	target := GetTargetModel("GLM-4.7-tools-thinking")
	if target != "glm-4.7" {
		t.Errorf("GetTargetModel(GLM-4.7-tools-thinking) = %q, want %q", target, "glm-4.7")
	}
}

// ===== ModelList =====

func TestModelList_ContainsToolsVariants(t *testing.T) {
	expected := map[string]bool{
		"glm-4.7-tools":          false,
		"glm-4.7-tools-thinking": false,
		"glm-5":                  false,
		"glm-5-tools":            false,
		"glm-5.1":                false,
		"glm-5.1-tools":          false,
	}

	for _, m := range ModelList {
		if _, ok := expected[m]; ok {
			expected[m] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("ModelList missing %q", name)
		}
	}
}
