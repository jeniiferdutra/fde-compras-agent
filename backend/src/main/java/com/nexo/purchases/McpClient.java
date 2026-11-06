package com.nexo.purchases;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import java.util.*;
import java.util.concurrent.atomic.AtomicLong;

@Component
public class McpClient {
    private final RestClient client;
    private final ObjectMapper mapper;
    private final AtomicLong ids = new AtomicLong(1);
    private final String url;

    public McpClient(RestClient.Builder builder, ObjectMapper mapper, @Value("${mcp.url}") String url) {
        this.client = builder.build(); this.mapper = mapper; this.url = url;
    }

    private JsonNode post(Map<String, ?> body) {
        return client.post().uri(url).contentType(MediaType.APPLICATION_JSON)
                .accept(MediaType.APPLICATION_JSON, MediaType.TEXT_EVENT_STREAM).header("MCP-Protocol-Version", "2025-03-26").body(body).retrieve().body(JsonNode.class);
    }

    private JsonNode request(String method, Object params) {
        long id = ids.getAndIncrement();
        Map<String,Object> envelope = new LinkedHashMap<>();
        envelope.put("jsonrpc", "2.0"); envelope.put("id", id); envelope.put("method", method);
        if (params != null) envelope.put("params", params);
        JsonNode response = post(envelope);
        if (response == null || response.has("error")) {
            String detail = response == null ? "resposta vazia" : response.path("error").path("message").asText("falha sem detalhes");
            throw new IllegalStateException("Erro MCP em " + method + ": " + detail);
        }
        return response.path("result");
    }

    public List<Map<String,Object>> discoverTools() {
        request("initialize", Map.of("protocolVersion", "2025-03-26", "capabilities", Map.of(),
                "clientInfo", Map.of("name", "nexo-purchase-api", "version", "1.0.0")));
        post(Map.of("jsonrpc", "2.0", "method", "notifications/initialized"));
        JsonNode result = request("tools/list", Map.of());
        List<Map<String,Object>> tools = new ArrayList<>();
        result.path("tools").forEach(tool -> {
            Map<String,Object> function = new LinkedHashMap<>();
            function.put("name", tool.path("name").asText());
            function.put("description", tool.path("description").asText(""));
            function.put("parameters", mapper.convertValue(tool.path("inputSchema"), Map.class));
            tools.add(Map.of("type", "function", "function", function));
        });
        return tools;
    }

    public JsonNode callTool(String name, JsonNode arguments) {
        JsonNode result = request("tools/call", Map.of("name", name, "arguments", arguments));
        if (result.path("isError").asBoolean()) throw new IllegalStateException("A ferramenta MCP " + name + " retornou erro.");
        JsonNode structured = result.path("structuredContent");
        if (!structured.isMissingNode() && !structured.isNull()) return structured;
        JsonNode content = result.path("content");
        if (content.isArray() && !content.isEmpty()) {
            String text = content.get(0).path("text").asText();
            try { return mapper.readTree(text); }
            catch (Exception ignored) { return mapper.createObjectNode().put("resultado", text); }
        }
        return result;
    }
}
