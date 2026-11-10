package moonshot

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertOpenAIRequestKimiK26UsesOnlyAllowedTemperature(t *testing.T) {
	request := &dto.GeneralOpenAIRequest{
		Model:       "kimi-k2.6",
		Temperature: common.GetPointer[float64](0.7),
	}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "kimi-k2.6",
		},
	}

	converted, err := (&Adaptor{}).ConvertOpenAIRequest(nil, info, request)

	require.NoError(t, err)
	convertedRequest, ok := converted.(*dto.GeneralOpenAIRequest)
	require.True(t, ok)
	require.NotNil(t, convertedRequest.Temperature)
	require.Equal(t, 1.0, *convertedRequest.Temperature)
}

func TestConvertOpenAIRequestKimiK26KeepsOmittedTemperatureOmitted(t *testing.T) {
	request := &dto.GeneralOpenAIRequest{
		Model: "kimi-k2.6",
	}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "kimi-k2.6",
		},
	}

	converted, err := (&Adaptor{}).ConvertOpenAIRequest(nil, info, request)

	require.NoError(t, err)
	convertedRequest, ok := converted.(*dto.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Nil(t, convertedRequest.Temperature)
}

func TestConvertOpenAIRequestOtherMoonshotModelKeepsTemperature(t *testing.T) {
	request := &dto.GeneralOpenAIRequest{
		Model:       "kimi-k2.5",
		Temperature: common.GetPointer[float64](0.7),
	}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "kimi-k2.5",
		},
	}

	converted, err := (&Adaptor{}).ConvertOpenAIRequest(nil, info, request)

	require.NoError(t, err)
	convertedRequest, ok := converted.(*dto.GeneralOpenAIRequest)
	require.True(t, ok)
	require.NotNil(t, convertedRequest.Temperature)
	require.Equal(t, 0.7, *convertedRequest.Temperature)
}

func TestGetRequestURL(t *testing.T) {
	for _, tc := range []struct {
		name   string
		base   string
		format types.RelayFormat
		mode   int
		want   string
	}{
		{"native responses", "https://api.moonshot.ai", types.RelayFormatOpenAIResponses, constant.RelayModeResponses, "https://api.moonshot.ai/v1/responses"},
		{"coding plan responses", "kimi-coding-plan", types.RelayFormatOpenAIResponses, constant.RelayModeResponses, "https://api.kimi.com/coding/v1/responses"},
		{"coding plan responses mode", "kimi-coding-plan", "", constant.RelayModeResponses, "https://api.kimi.com/coding/v1/responses"},
		{"coding plan chat", "kimi-coding-plan", types.RelayFormatOpenAI, constant.RelayModeChatCompletions, "https://api.kimi.com/coding/v1/chat/completions"},
		{"coding plan claude", "kimi-coding-plan", types.RelayFormatClaude, constant.RelayModeChatCompletions, "https://api.kimi.com/coding/v1/messages"},
		{"native chat", "https://api.moonshot.ai", types.RelayFormatOpenAI, constant.RelayModeChatCompletions, "https://api.moonshot.ai/v1/chat/completions"},
		{"native claude", "https://api.moonshot.ai", types.RelayFormatClaude, constant.RelayModeChatCompletions, "https://api.moonshot.ai/anthropic/v1/messages"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			info := &relaycommon.RelayInfo{
				ChannelMeta: &relaycommon.ChannelMeta{ChannelBaseUrl: tc.base},
				RelayMode:   tc.mode,
				RelayFormat: tc.format,
			}
			url, err := (&Adaptor{}).GetRequestURL(info)
			require.NoError(t, err)
			assert.Equal(t, tc.want, url)
		})
	}
}

func TestConvertOpenAIResponsesRequest(t *testing.T) {
	req := dto.OpenAIResponsesRequest{
		Model: "kimi-k2.7-code",
		Reasoning: &dto.Reasoning{
			Effort: "high",
		},
	}
	info := &relaycommon.RelayInfo{}
	converted, err := (&Adaptor{}).ConvertOpenAIResponsesRequest(nil, info, req)
	require.NoError(t, err)
	convertedReq, ok := converted.(dto.OpenAIResponsesRequest)
	require.True(t, ok)
	require.Equal(t, "kimi-k2.7-code", convertedReq.Model)
	require.Equal(t, "high", info.ReasoningEffort)
}

func TestConvertOpenAIResponsesRequestToolsSanitization(t *testing.T) {
	req := dto.OpenAIResponsesRequest{
		Model: "kimi-k3",
		Tools: []byte(`[
			{"type": "function", "name": "get_weather", "description": "Get weather", "parameters": {"type": "object"}},
			{"type": "custom", "name": "apply_patch", "description": "Apply patch"},
			{"type": "tool_search"},
			{"type": "web_search"},
			{"type": "image_generation"},
			{"type": "code_interpreter"}
		]`),
	}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "kimi-k3",
		},
	}
	converted, err := (&Adaptor{}).ConvertOpenAIResponsesRequest(nil, info, req)
	require.NoError(t, err)
	convertedReq, ok := converted.(dto.OpenAIResponsesRequest)
	require.True(t, ok)

	var tools []map[string]any
	require.NoError(t, common.Unmarshal(convertedReq.Tools, &tools))
	require.Len(t, tools, 3)

	require.Equal(t, "function", tools[0]["type"])
	require.Equal(t, "get_weather", tools[0]["name"])

	require.Equal(t, "function", tools[1]["type"])
	require.Equal(t, "apply_patch", tools[1]["name"])
	require.Contains(t, tools[1]["description"], "Apply patch")
	require.NotNil(t, tools[1]["parameters"])
	require.True(t, info.ResponsesToolState().IsCustomTool("apply_patch"))

	require.Equal(t, "web_search", tools[2]["type"])
}

func TestResponsesNamespaceAndCustomToolHistory(t *testing.T) {
	req := dto.OpenAIResponsesRequest{
		Model: "k3.1-pro",
		Tools: json.RawMessage(`[
			{"type":"namespace","name":"compat","tools":[
				{"type":"function","name":"probe_echo","parameters":{"type":"object","properties":{"marker":{"type":"integer","enum":[9007199254740993]}}}},
				{"type":"custom","name":"patch","description":"Apply patch","format":{"type":"grammar","syntax":"regex","definition":"PATCH.*"}}
			]},
			{"type":"web_search","external_web_access":true,"extension":9007199254740993},
			{"type":"tool_search"}
		]`),
		Input: json.RawMessage(`[
			{"type":"function_call","namespace":"compat","name":"probe_echo","call_id":"c1","arguments":"{}"},
			{"type":"function_call_output","call_id":"c1","output":"OK"},
			{"type":"custom_tool_call","namespace":"compat","name":"patch","call_id":"c2","input":"PATCH\n\"hello\""},
			{"type":"custom_tool_call_output","call_id":"c2","output":"patched"}
		]`),
		ToolChoice: json.RawMessage(`{"type":"custom","namespace":"compat","name":"patch"}`),
	}
	info := &relaycommon.RelayInfo{}
	converted, err := (&Adaptor{}).ConvertOpenAIResponsesRequest(nil, info, req)
	require.NoError(t, err)
	got, ok := converted.(dto.OpenAIResponsesRequest)
	require.True(t, ok)
	var tools []map[string]json.RawMessage
	require.NoError(t, common.Unmarshal(got.Tools, &tools))
	require.Len(t, tools, 3)
	require.Len(t, info.ResponsesNamespaceTools(), 2)
	assert.Equal(t, "probe_echo", info.ResponsesNamespaceTools()["compat__probe_echo"].Name)
	assert.Equal(t, "compat", info.ResponsesNamespaceTools()["compat__probe_echo"].Namespace)
	assert.Equal(t, `"function"`, string(tools[0]["type"]))
	assert.Contains(t, string(tools[0]["parameters"]), "9007199254740993")
	assert.Equal(t, `"function"`, string(tools[1]["type"]))
	assert.Contains(t, string(tools[1]["description"]), "PATCH.*")
	assert.NotContains(t, tools[1], "format")
	assert.True(t, info.ResponsesToolState().IsCustomTool("compat__patch"))
	assert.False(t, info.ResponsesToolState().IsCustomTool("compat__probe_echo"))
	assert.NotContains(t, tools[2], "external_web_access")
	assert.Equal(t, "9007199254740993", string(tools[2]["extension"]))
	assert.JSONEq(t, `{"type":"function","name":"compat__patch"}`, string(got.ToolChoice))
	assert.JSONEq(t, `[
		{"type":"function_call","name":"compat__probe_echo","call_id":"c1","arguments":"{}"},
		{"type":"function_call_output","call_id":"c1","output":"OK"},
		{"type":"function_call","name":"compat__patch","call_id":"c2","arguments":"{\"input\":\"PATCH\\n\\\"hello\\\"\"}"},
		{"type":"function_call_output","call_id":"c2","output":"patched"}
	]`, string(got.Input))
	assert.Contains(t, string(req.Tools), `"type":"namespace"`, "retries must retain the client request")

	_, err = (&Adaptor{}).ConvertOpenAIResponsesRequest(nil, info, dto.OpenAIResponsesRequest{Model: "k3.1-pro"})
	require.NoError(t, err)
	assert.Nil(t, info.ResponsesNamespaceTools())
	assert.Nil(t, info.ResponsesToolState())
}

// Kimi serves OpenAI Responses natively (kimi-k3); the body is forwarded
// unchanged and other models are rejected by the upstream.
func TestResponsesRequestPassesThroughToV1Responses(t *testing.T) {
	const body = `{"model":"kimi-k3","input":"Latest Go release?","tools":[{"type":"web_search"}],"stream":false}`
	info := &relaycommon.RelayInfo{
		RelayMode:   constant.RelayModeResponses,
		RelayFormat: types.RelayFormatOpenAIResponses,
		ChannelMeta: &relaycommon.ChannelMeta{ChannelBaseUrl: "https://api.moonshot.invalid", UpstreamModelName: "kimi-k3"},
	}
	url, err := (&Adaptor{}).GetRequestURL(info)
	require.NoError(t, err)
	assert.Equal(t, "https://api.moonshot.invalid/v1/responses", url)

	var request dto.OpenAIResponsesRequest
	require.NoError(t, common.UnmarshalJsonStr(body, &request))
	converted, err := (&Adaptor{}).ConvertOpenAIResponsesRequest(nil, info, request)
	require.NoError(t, err)
	wire, err := common.Marshal(converted)
	require.NoError(t, err)
	assert.JSONEq(t, body, string(wire))
}
