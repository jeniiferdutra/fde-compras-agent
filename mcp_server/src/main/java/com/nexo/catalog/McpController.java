package com.nexo.catalog;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestClient;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.server.ResponseStatusException;

import java.util.*;

@RestController
public class McpController {
    private final RestClient http;
    private final ObjectMapper mapper;
    private final String baseUrl;

    public McpController(RestClient.Builder builder, ObjectMapper mapper,
                         @Value("${dummyjson.base-url}") String baseUrl) {
        this.baseUrl=baseUrl.replaceAll("/$", "");
        SimpleClientHttpRequestFactory requestFactory = new SimpleClientHttpRequestFactory();
        requestFactory.setConnectTimeout(java.time.Duration.ofSeconds(3));
        requestFactory.setReadTimeout(java.time.Duration.ofSeconds(5));
        this.http=builder.baseUrl(this.baseUrl).requestFactory(requestFactory).build(); this.mapper=mapper;
    }

    @GetMapping("/health") public Map<String,String> health() { return Map.of("status", "ok"); }

    @PostMapping(value="/mcp", consumes=MediaType.APPLICATION_JSON_VALUE, produces=MediaType.APPLICATION_JSON_VALUE)
    public ResponseEntity<?> handle(@RequestBody JsonNode message) {
        String method = message.path("method").asText();
        if (method.startsWith("notifications/")) return ResponseEntity.status(HttpStatus.ACCEPTED).build();
        JsonNode id = message.path("id");
        try {
            Object result = switch (method) {
                case "initialize" -> initialize(message.path("params"));
                case "ping" -> Map.of();
                case "tools/list" -> Map.of("tools", toolDefinitions());
                case "tools/call" -> call(message.path("params"));
                default -> throw new McpError(-32601, "Método MCP não encontrado: " + method);
            };
            return ResponseEntity.ok(envelope(id, result));
        } catch (McpError e) {
            return ResponseEntity.ok(errorEnvelope(id, e.code, e.getMessage()));
        } catch (ResponseStatusException e) {
            return ResponseEntity.ok(errorEnvelope(id, -32000, e.getReason() == null ? "Falha ao consultar o catálogo." : e.getReason()));
        } catch (Exception e) {
            return ResponseEntity.ok(errorEnvelope(id, -32603, "Erro interno ao consultar o catálogo. Tente novamente."));
        }
    }

    private Map<String,Object> initialize(JsonNode params) {
        String requested = params.path("protocolVersion").asText("2025-03-26");
        return Map.of("protocolVersion", requested, "capabilities", Map.of("tools", Map.of("listChanged", false)),
                "serverInfo", Map.of("name", "catalogo-dummyjson", "version", "1.0.0"));
    }

    private List<Map<String,Object>> toolDefinitions() {
        Map<String,Object> searchSchema = Map.of("type", "object", "properties", Map.of("termo", Map.of("type", "string", "description", "Palavras-chave para pesquisar no catálogo")), "required", List.of("termo"));
        Map<String,Object> detailSchema = Map.of("type", "object", "properties", Map.of("id", Map.of("type", "integer", "description", "ID do produto no catálogo")), "required", List.of("id"));
        return List.of(
                Map.of("name", "buscar_produtos", "description", "Busca produtos no catálogo DummyJSON por palavras-chave. Retorna somente produtos encontrados.", "inputSchema", searchSchema),
                Map.of("name", "detalhar_produto", "description", "Obtém os dados atuais de um produto específico pelo ID no catálogo DummyJSON.", "inputSchema", detailSchema));
    }

    private Map<String,Object> call(JsonNode params) {
        String name = params.path("name").asText();
        JsonNode args = params.path("arguments");
        Map<String,Object> data = switch (name) {
            case "buscar_produtos" -> search(args.path("termo").asText(""));
            case "detalhar_produto" -> detail(args.path("id").asInt(-1));
            default -> throw new McpError(-32602, "Ferramenta MCP desconhecida: " + name);
        };
        String text;
        try { text = mapper.writeValueAsString(data); }
        catch (Exception e) { throw new McpError(-32603, "Falha ao serializar o resultado."); }
        return Map.of("content", List.of(Map.of("type", "text", "text", text)), "structuredContent", data, "isError", false);
    }

    private Map<String,Object> search(String term) {
        if (term.isBlank()) return Map.of("erro", "Informe um termo de busca.");
        JsonNode payload = get("/products/search", Map.of("q", term.trim(), "limit", "10"));
        List<Map<String,Object>> products = new ArrayList<>();
        payload.path("products").forEach(p -> products.add(product(p, false)));
        return Map.of("total", payload.path("total").asInt(0), "produtos", products);
    }

    private Map<String,Object> detail(int id) {
        if (id < 1) return Map.of("erro", "Informe um ID válido.");
        try { return product(get("/products/" + id, Map.of()), true); }
        catch (ResponseStatusException e) {
            if (e.getStatusCode().value() == 404) return Map.of("erro", "Produto " + id + " não encontrado.");
            throw e;
        }
    }

    private JsonNode get(String path, Map<String,String> query) {
        try {
            RestClient.RequestHeadersUriSpec<?> request = http.get();
            RestClient.RequestHeadersSpec<?> spec = request.uri(uriBuilder -> {
                uriBuilder.path(path); query.forEach(uriBuilder::queryParam); return uriBuilder.build();
            });
            return spec.retrieve().body(JsonNode.class);
        } catch (org.springframework.web.client.HttpClientErrorException e) {
            throw new ResponseStatusException(e.getStatusCode(), "A DummyJSON recusou a consulta (HTTP " + e.getStatusCode().value() + "). Confira o termo ou o ID do produto.");
        } catch (ResourceAccessException e) {
            if (hasTimeoutCause(e))
                throw new ResponseStatusException(HttpStatus.GATEWAY_TIMEOUT, "A consulta à DummyJSON demorou mais de 5 segundos. Tente novamente.");
            throw new ResponseStatusException(HttpStatus.BAD_GATEWAY, "Não foi possível conectar à DummyJSON. Verifique sua conexão e tente novamente.");
        } catch (org.springframework.web.client.HttpServerErrorException e) {
            throw new ResponseStatusException(HttpStatus.BAD_GATEWAY, "A DummyJSON está temporariamente indisponível (HTTP " + e.getStatusCode().value() + "). Tente novamente em alguns instantes.");
        } catch (Exception e) { throw new ResponseStatusException(HttpStatus.BAD_GATEWAY, "Falha inesperada ao consultar a DummyJSON. Tente novamente."); }
    }

    private boolean hasTimeoutCause(Throwable error) {
        for (Throwable cause = error; cause != null; cause = cause.getCause())
            if (cause instanceof java.net.SocketTimeoutException || cause instanceof java.net.http.HttpTimeoutException) return true;
        return false;
    }

    private Map<String,Object> product(JsonNode p, boolean includeStock) {
        Map<String,Object> result = new LinkedHashMap<>();
        result.put("id", p.path("id").asInt()); result.put("titulo", p.path("title").asText());
        result.put("categoria", p.path("category").asText("")); result.put("preco", p.path("price").decimalValue());
        result.put("moeda", "USD"); result.put("descricao", p.path("description").asText(""));
        if (includeStock) result.put("estoque", p.path("stock").asInt());
        return result;
    }

    private Map<String,Object> envelope(JsonNode id, Object result) { return Map.of("jsonrpc", "2.0", "id", mapper.convertValue(id, Object.class), "result", result); }
    private Map<String,Object> errorEnvelope(JsonNode id, int code, String message) {
        return Map.of("jsonrpc", "2.0", "id", mapper.convertValue(id, Object.class), "error", Map.of("code", code, "message", message));
    }
    private static final class McpError extends RuntimeException {
        final int code; McpError(int code, String message) { super(message); this.code=code; }
    }
}
