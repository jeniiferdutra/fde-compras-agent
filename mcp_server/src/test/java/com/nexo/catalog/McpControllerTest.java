package com.nexo.catalog;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.http.ResponseEntity;
import org.springframework.web.client.RestClient;

import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.atomic.AtomicBoolean;

import static org.junit.jupiter.api.Assertions.*;

class McpControllerTest {
    private final ObjectMapper mapper = new ObjectMapper();
    private HttpServer dummyJson;
    private McpController controller;
    private final AtomicBoolean failUpstream = new AtomicBoolean();

    @BeforeEach
    void setUp() throws Exception {
        dummyJson = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        dummyJson.createContext("/", exchange -> {
            if (failUpstream.get()) {
                byte[] bytes = "{}".getBytes(StandardCharsets.UTF_8);
                exchange.sendResponseHeaders(503, bytes.length);
                exchange.getResponseBody().write(bytes);
                exchange.close();
                return;
            }
            String path = exchange.getRequestURI().getPath();
            String response;
            if (path.equals("/products/search")) {
                response = """
                        {"total":1,"products":[{"id":42,"title":"Laptop de teste","category":"laptops","price":999.99,"description":"Produto de teste"}]}
                        """;
            } else if (path.equals("/products/42")) {
                response = """
                        {"id":42,"title":"Laptop de teste","category":"laptops","price":999.99,"description":"Produto de teste","stock":7}
                        """;
            } else {
                response = "{}";
            }
            byte[] bytes = response.getBytes(StandardCharsets.UTF_8);
            exchange.getResponseHeaders().set("Content-Type", "application/json");
            exchange.sendResponseHeaders(200, bytes.length);
            exchange.getResponseBody().write(bytes);
            exchange.close();
        });
        dummyJson.start();
        controller = new McpController(RestClient.builder(), mapper,
                "http://127.0.0.1:" + dummyJson.getAddress().getPort());
    }

    @AfterEach
    void tearDown() {
        if (dummyJson != null) dummyJson.stop(0);
    }

    @Test
    void buscarProdutosReturnsCatalogMatches() throws Exception {
        JsonNode response = invokeTool("buscar_produtos", "{\"termo\":\"laptop\"}");
        JsonNode result = response.path("result").path("structuredContent");

        assertFalse(response.has("error"));
        assertEquals(1, result.path("total").asInt());
        assertEquals(42, result.path("produtos").get(0).path("id").asInt());
        assertEquals("Laptop de teste", result.path("produtos").get(0).path("titulo").asText());
    }

    @Test
    void detalharProdutoReturnsCurrentProductDetails() throws Exception {
        JsonNode response = invokeTool("detalhar_produto", "{\"id\":42}");
        JsonNode result = response.path("result").path("structuredContent");

        assertFalse(response.has("error"));
        assertEquals(42, result.path("id").asInt());
        assertEquals(999.99, result.path("preco").asDouble(), 0.001);
        assertEquals(7, result.path("estoque").asInt());
    }

    @Test
    void dummyJsonFailureReturnsActionableMcpError() throws Exception {
        failUpstream.set(true);
        JsonNode response = invokeTool("buscar_produtos", "{\"termo\":\"laptop\"}");

        assertTrue(response.path("error").path("message").asText().contains("DummyJSON está temporariamente indisponível"));
    }

    private JsonNode invokeTool(String name, String arguments) throws Exception {
        ObjectNode request = mapper.createObjectNode();
        request.put("jsonrpc", "2.0");
        request.put("id", 1);
        request.put("method", "tools/call");
        ObjectNode params = request.putObject("params");
        params.put("name", name);
        params.set("arguments", mapper.readTree(arguments));
        @SuppressWarnings("unchecked")
        ResponseEntity<java.util.Map<String, Object>> response = (ResponseEntity<java.util.Map<String, Object>>) controller.handle(request);
        assertEquals(200, response.getStatusCode().value());
        return mapper.valueToTree(response.getBody());
    }
}
