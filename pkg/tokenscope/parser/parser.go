package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

// Part types
const (
	PartTypeSystem  = "system"
	PartTypeHistory = "history"
	PartTypeTool    = "tool"
	PartTypeFile    = "file"
	PartTypeMemory  = "memory"
	PartTypeUser    = "user"
)

// ParseContextParts parses a raw request body and returns context parts.
// It supports OpenAI, Claude, and Gemini formats based on relayFormat.
// Token estimation uses a simple 4-chars-per-token approximation.
func ParseContextParts(requestBody []byte, relayFormat string) []model.RequestContextPart {
	var body map[string]interface{}
	if err := json.Unmarshal(requestBody, &body); err != nil {
		return nil
	}

	var parts []model.RequestContextPart
	position := 0

	switch relayFormat {
	case "claude", "Claude Messages":
		parts = parseClaudeFormat(body, &position)
	case "gemini", "Google Gemini":
		parts = parseGeminiFormat(body, &position)
	default: // OpenAI compatible
		parts = parseOpenAIFormat(body, &position)
	}

	// Post-process: mark prefix, repeated, stable, cache-friendly
	analyzeParts(parts)

	return parts
}

// parseOpenAIFormat parses OpenAI chat.completions format.
func parseOpenAIFormat(body map[string]interface{}, position *int) []model.RequestContextPart {
	var parts []model.RequestContextPart

	// System messages
	messages, _ := body["messages"].([]interface{})
	var systemTexts []string
	for _, msgAny := range messages {
		msg, ok := msgAny.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		if role == "system" {
			content := extractMessageText(msg)
			if content != "" {
				systemTexts = append(systemTexts, content)
			}
		}
	}

	if len(systemTexts) > 0 {
		text := strings.Join(systemTexts, "\n")
		parts = append(parts, makePart(PartTypeSystem, "system_prompt", text, *position, true))
		*position++
	}

	// Messages (history)
	msgPosition := 0
	for _, msgAny := range messages {
		msg, ok := msgAny.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		if role == "system" {
			continue
		}

		content := extractMessageText(msg)
		if content == "" {
			continue
		}

		partType := PartTypeHistory
		switch role {
		case "assistant":
			// Check for tool calls
			if tc, ok := msg["tool_calls"]; ok && tc != nil {
				parts = append(parts, makePart(PartTypeTool, "tool_calls", formatToolCalls(tc), *position, false))
				*position++
				continue
			}
		case "tool":
			partType = PartTypeTool
			toolName := extractToolName(msg)
			if toolName != "" {
				parts = append(parts, makePart(partType, toolName, content, *position, false))
			} else {
				parts = append(parts, makePart(partType, "tool_result", content, *position, false))
			}
			*position++
			continue
		}

		parts = append(parts, makePart(partType, role+"_message", content, *position, false))
		msgPosition++
		*position++
	}

	// Tools (tool definitions)
	tools, _ := body["tools"].([]interface{})
	if len(tools) > 0 {
		toolJSON, _ := json.Marshal(tools)
		parts = append(parts, makePart(PartTypeTool, "tool_definitions", string(toolJSON), *position, true))
		*position++
	}

	// Check for system prompt in other fields
	if sysPrompt, ok := body["system_prompt"].(string); ok && sysPrompt != "" {
		parts = append(parts, makePart(PartTypeSystem, "system_prompt_override", sysPrompt, *position, true))
		*position++
	}

	return parts
}

// parseClaudeFormat parses Claude messages format.
func parseClaudeFormat(body map[string]interface{}, position *int) []model.RequestContextPart {
	var parts []model.RequestContextPart

	// System prompt (Claude has a separate "system" field)
	if systemTexts, ok := body["system"].([]interface{}); ok {
		var texts []string
		for _, item := range systemTexts {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, ok := itemMap["text"].(string); ok {
					texts = append(texts, text)
				}
			}
		}
		if len(texts) > 0 {
			text := strings.Join(texts, "\n")
			parts = append(parts, makePart(PartTypeSystem, "system_prompt", text, *position, true))
			*position++
		}
	} else if systemStr, ok := body["system"].(string); ok && systemStr != "" {
		parts = append(parts, makePart(PartTypeSystem, "system_prompt", systemStr, *position, true))
		*position++
	}

	// Messages
	messages, _ := body["messages"].([]interface{})
	for _, msgAny := range messages {
		msg, ok := msgAny.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		content := extractMessageText(msg)
		if content == "" {
			continue
		}

		partType := PartTypeHistory
		if role == "user" && len(parts) == 0 {
			// First user message is part of the prefix
		}

		// Tool use and tool results
		if contentObj, ok := msg["content"].([]interface{}); ok {
			for _, item := range contentObj {
				if itemMap, ok := item.(map[string]interface{}); ok {
					typeStr, _ := itemMap["type"].(string)
					if typeStr == "tool_use" {
						name, _ := itemMap["name"].(string)
						inputJSON, _ := json.Marshal(itemMap["input"])
						parts = append(parts, makePart(PartTypeTool, name, string(inputJSON), *position, false))
						*position++
						continue
					}
					if typeStr == "tool_result" {
						content, _ := itemMap["content"].(string)
						parts = append(parts, makePart(PartTypeTool, "tool_result", content, *position, false))
						*position++
						continue
					}
				}
			}
		}

		parts = append(parts, makePart(partType, role+"_message", content, *position, false))
		*position++
	}

	// Tools
	tools, _ := body["tools"].([]interface{})
	if len(tools) > 0 {
		toolJSON, _ := json.Marshal(tools)
		parts = append(parts, makePart(PartTypeTool, "tool_definitions", string(toolJSON), *position, true))
		*position++
	}

	return parts
}

// parseGeminiFormat parses Gemini generateContent format.
func parseGeminiFormat(body map[string]interface{}, position *int) []model.RequestContextPart {
	var parts []model.RequestContextPart

	// System instruction
	if sysInstr, ok := body["system_instruction"].(map[string]interface{}); ok {
		if partsArr, ok := sysInstr["parts"].([]interface{}); ok {
			var texts []string
			for _, p := range partsArr {
				if pMap, ok := p.(map[string]interface{}); ok {
					if text, ok := pMap["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
			if len(texts) > 0 {
				parts = append(parts, makePart(PartTypeSystem, "system_instruction", strings.Join(texts, "\n"), *position, true))
				*position++
			}
		}
	}

	// Contents
	contents, _ := body["contents"].([]interface{})
	for _, cAny := range contents {
		content, ok := cAny.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := content["role"].(string)
		partsArr, _ := content["parts"].([]interface{})
		for _, p := range partsArr {
			pMap, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			text, _ := pMap["text"].(string)
			if text == "" {
				continue
			}
			partType := PartTypeHistory
			parts = append(parts, makePart(partType, role+"_content", text, *position, false))
			*position++
		}
	}

	// Tools
	tools, _ := body["tools"].([]interface{})
	if len(tools) > 0 {
		toolJSON, _ := json.Marshal(tools)
		parts = append(parts, makePart(PartTypeTool, "tool_definitions", string(toolJSON), *position, true))
		*position++
	}

	return parts
}

// analyzeParts post-processes parts to mark prefix, repeated, stable, cache-friendly.
func analyzeParts(parts []model.RequestContextPart) {
	if len(parts) == 0 {
		return
	}

	// Mark first N parts as prefix (system + tool definitions are always prefix)
	for i := range parts {
		if parts[i].PartType == PartTypeSystem || parts[i].PartName == "tool_definitions" {
			parts[i].IsPrefix = true
		}
	}

	// Detect repeated parts using content_hash
	hashCount := make(map[string]int)
	for _, p := range parts {
		if p.ContentHash != "" {
			hashCount[p.ContentHash]++
		}
	}
	for i := range parts {
		if parts[i].IsPrefix && hashCount[parts[i].ContentHash] > 0 {
			parts[i].IsRepeated = true
			parts[i].IsStable = true
		}
	}

	// Mark cache-friendly: prefix + stable
	for i := range parts {
		if parts[i].IsPrefix && parts[i].IsStable {
			parts[i].IsCacheFriendly = true
		}
	}
}

// makePart creates a RequestContextPart from text content.
func makePart(partType, partName, text string, position int, isPrefix bool) model.RequestContextPart {
	return model.RequestContextPart{
		PartType:    partType,
		PartName:    partName,
		ContentHash: contentHash(text),
		TokenCount:  estimateTokens(text),
		Position:    position,
		IsPrefix:    isPrefix,
	}
}

// contentHash computes a SHA-256 prefix hash of text.
func contentHash(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:16]) // 32 hex chars
}

// estimateTokens uses a simple 4-chars-per-token approximation.
func estimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return len(text) / 4
}

// extractMessageText extracts text content from a message map.
func extractMessageText(msg map[string]interface{}) string {
	content := msg["content"]
	if content == nil {
		return ""
	}
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var texts []string
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["type"] == "text" {
					if text, ok := itemMap["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
		}
		return strings.Join(texts, "\n")
	}
	return ""
}

// extractToolName extracts the tool name from a tool message.
func extractToolName(msg map[string]interface{}) string {
	if name, ok := msg["tool_call_id"].(string); ok {
		return name
	}
	return ""
}

// formatToolCalls formats tool_calls field into a readable string.
func formatToolCalls(tc interface{}) string {
	b, _ := json.Marshal(tc)
	return string(b)
}

// ParseAndStore parses the request body and stores context parts in the database.
// It marks the debug payload as parsed.
func ParseAndStore(requestId string, requestBody []byte, relayInfo *relaycommon.RelayInfo) error {
	relayFormat := string(relayInfo.GetFinalRequestRelayFormat())
	parts := ParseContextParts(requestBody, relayFormat)
	if len(parts) == 0 {
		// Mark as parsed even with no parts to avoid re-parsing
		return model.MarkDebugPayloadParsed(requestId)
	}

	if err := model.CreateRequestContextParts(requestId, parts); err != nil {
		return err
	}
	return model.MarkDebugPayloadParsed(requestId)
}
