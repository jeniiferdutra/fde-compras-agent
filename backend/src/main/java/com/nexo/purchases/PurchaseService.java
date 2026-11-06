package com.nexo.purchases;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Service;
import org.springframework.transaction.support.TransactionTemplate;
import org.springframework.web.client.RestClient;
import org.springframework.web.client.RestClientResponseException;
import org.springframework.web.server.ResponseStatusException;
import org.springframework.http.HttpStatus;
import org.springframework.jdbc.core.JdbcTemplate;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.OffsetDateTime;
import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import static com.nexo.purchases.Models.*;

@Service
public class PurchaseService {
    private final McpClient mcp;
    private final RestClient http;
    private final ObjectMapper mapper;
    private final JdbcTemplate jdbc;
    private final TransactionTemplate transaction;
    private final String apiKey, model, llmUrl;

    public PurchaseService(McpClient mcp, RestClient.Builder builder, ObjectMapper mapper, JdbcTemplate jdbc,
            TransactionTemplate transaction,
            @Value("${openrouter.api-key}") String apiKey,
            @Value("${openrouter.model}") String model,
            @Value("${openrouter.base-url}") String baseUrl) {
        this.mcp=mcp; this.http=builder.build(); this.mapper=mapper; this.jdbc=jdbc; this.transaction=transaction;
        this.apiKey=apiKey; this.model=model; this.llmUrl=baseUrl.replaceAll("/$", "") + "/chat/completions";
    }

    public CreateResponse create(String pedido) {
        if (pedido == null || pedido.trim().length() < 3 || pedido.length() > 2000)
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "O pedido deve ter entre 3 e 2.000 caracteres.");
        pedido = pedido.trim();
        if (apiKey.isBlank()) throw new ResponseStatusException(HttpStatus.SERVICE_UNAVAILABLE,
                "LLM não configurado. Defina OPENROUTER_API_KEY no arquivo .env e reinicie o serviço api.");
        AgentResult answer = runAgent(pedido);
        if (answer.items.isEmpty()) return new CreateResponse("nao_atendido", answer.message, null);
        BigDecimal total = answer.items.stream().map(Item::line_total).reduce(BigDecimal.ZERO, BigDecimal::add);
        String requestText = pedido;
        Long id = transaction.execute(status -> {
            Long newId = jdbc.queryForObject("INSERT INTO requisitions(request_text,response_text,total) VALUES (?,?,?) RETURNING id",
                    Long.class, requestText, answer.message, total);
            for (Item item : answer.items) jdbc.update("INSERT INTO requisition_items(requisition_id,product_id,title,quantity,unit_price,line_total) VALUES (?,?,?,?,?,?)",
                    newId, item.product_id(), item.title(), item.quantity(), item.unit_price(), item.line_total());
            return newId;
        });
        Requisition saved = getOne(Objects.requireNonNull(id));
        return new CreateResponse("ok", answer.message, saved);
    }

    private AgentResult runAgent(String pedido) {
        List<Map<String,Object>> tools;
        try { tools = mcp.discoverTools(); }
        catch (Exception e) { throw new ResponseStatusException(HttpStatus.BAD_GATEWAY, "Não foi possível conectar ao servidor MCP. Confira se os serviços estão ativos e tente novamente."); }
        Constraints constraints = parseConstraints(pedido);
        String system = """
                Você é um agente de compras. Use as ferramentas MCP para buscar e detalhar produtos reais do catálogo.
                Nunca invente produtos, IDs, preços, disponibilidade ou atributos. Use buscar_produtos para pesquisar e detalhar_produto para confirmar cada item escolhido. Só selecione itens que correspondam ao pedido e respeitem o teto de preço unitário explicitado. Se houver mais de uma opção que atenda aos critérios, escolha a melhor correspondência do catálogo; se forem igualmente adequadas, escolha a de menor preço. Não peça para a pessoa escolher entre modelos quando ela já informou o tipo de produto, quantidade e orçamento. Faça perguntas apenas quando faltar um critério essencial para identificar o tipo de produto ou a quantidade. Se não houver correspondência, explique objetivamente o que falta e não gere itens. Quando a seleção estiver confirmada, responda SOMENTE com JSON válido neste formato: {"status":"ok","mensagem":"...","itens":[{"product_id":1,"quantidade":2}]}. Para falta/ambiguidade: {"status":"nao_atendido","mensagem":"...","itens":[]}.
                Interprete valores no formato brasileiro: em um pedido em português, "US$ 2.000" representa dois mil dólares; mantenha os preços da resposta em USD.
                """ + "\nQuantidade explícita detectada: " + (constraints.quantity == null ? "não identificada" : constraints.quantity)
                + ". Teto unitário explícito USD: " + (constraints.maxPrice == null ? "não identificado" : constraints.maxPrice) + ".";
        List<Map<String,Object>> messages = new ArrayList<>();
        messages.add(Map.of("role", "system", "content", system));
        messages.add(Map.of("role", "user", "content", pedido));
        try {
            for (int i=0; i<7; i++) {
                Map<String,Object> body = new LinkedHashMap<>();
                body.put("model", model); body.put("messages", messages); body.put("tools", tools);
                body.put("tool_choice", "auto"); body.put("temperature", 0.1);
                body.put("provider", Map.of("zdr", true));
                JsonNode response = http.post().uri(llmUrl).header(HttpHeaders.AUTHORIZATION, "Bearer " + apiKey)
                        .header("HTTP-Referer", "http://localhost:5173").header("X-Title", "Agente de compras")
                        .contentType(MediaType.APPLICATION_JSON).body(body).retrieve().body(JsonNode.class);
                JsonNode msg = response.path("choices").path(0).path("message");
                messages.add(mapper.convertValue(msg, Map.class));
                JsonNode calls = msg.path("tool_calls");
                if (!calls.isArray() || calls.isEmpty()) return finish(msg.path("content").asText(""), constraints);
                for (JsonNode call : calls) {
                    String name = call.path("function").path("name").asText();
                    String rawArgs = call.path("function").path("arguments").asText("{}");
                    JsonNode arguments = mapper.readTree(rawArgs);
                    String output;
                    try { output = mcp.callTool(name, arguments).toString(); }
                    catch (Exception ex) { throw new ResponseStatusException(HttpStatus.BAD_GATEWAY,
                            "Não foi possível consultar o catálogo de produtos. " + userFacingCatalogError(ex)); }
                    messages.add(Map.of("role", "tool", "tool_call_id", call.path("id").asText(), "content", output));
                }
            }
            throw new IllegalStateException("O agente excedeu o limite de chamadas de ferramenta.");
        } catch (ResponseStatusException e) { throw e; }
        catch (RestClientResponseException e) {
            if (e.getStatusCode().value() == 401 || e.getStatusCode().value() == 403)
                throw new ResponseStatusException(HttpStatus.SERVICE_UNAVAILABLE,
                        "A chave da OpenRouter está inválida ou expirada. Atualize OPENROUTER_API_KEY no arquivo .env e reinicie a API.");
            if (e.getStatusCode().value() == 404)
                throw new ResponseStatusException(HttpStatus.BAD_GATEWAY,
                        "A OpenRouter não encontrou um endpoint do modelo compatível com as políticas de privacidade da conta. Confira OPENROUTER_MODEL e use um modelo com suporte a ferramentas e ZDR.");
            throw new ResponseStatusException(HttpStatus.BAD_GATEWAY,
                    "A OpenRouter respondeu com HTTP " + e.getStatusCode().value() + ". Confira a configuração e tente novamente.");
        }
        catch (Exception e) { throw new ResponseStatusException(HttpStatus.BAD_GATEWAY, "Não foi possível concluir o pedido com segurança. " + userFacingCatalogError(e)); }
    }

    private String userFacingCatalogError(Exception error) {
        String detail = error.getMessage();
        if (detail != null && (detail.contains("mais de 5 segundos") || detail.contains("demorou")))
            return "A consulta à DummyJSON demorou mais de 5 segundos. Tente novamente.";
        if (detail != null && (detail.contains("DummyJSON") || detail.contains("catálogo")))
            return detail;
        return "O catálogo DummyJSON está temporariamente indisponível. Tente novamente em alguns instantes.";
    }

    private AgentResult finish(String raw, Constraints constraints) throws Exception {
        int start = raw.indexOf('{'), end = raw.lastIndexOf('}');
        if (start < 0 || end < start) throw new IllegalStateException("Resposta do agente fora do formato JSON esperado.");
        JsonNode data = mapper.readTree(raw.substring(start, end + 1));
        String message = data.path("mensagem").asText("Não encontrei uma opção adequada.");
        if (!"ok".equals(data.path("status").asText())) return new AgentResult(message, List.of());
        List<Item> items = new ArrayList<>(); Set<Integer> ids = new HashSet<>(); int totalQuantity = 0;
        for (JsonNode selected : data.path("itens")) {
            int productId = selected.path("product_id").asInt();
            int qty = selected.path("quantidade").asInt();
            if (productId < 1 || !ids.add(productId) || qty < 1 || qty > 1000) throw new IllegalStateException("A lista de itens do agente é inválida.");
            JsonNode detail = mcp.callTool("detalhar_produto", mapper.createObjectNode().put("id", productId));
            if (detail.has("erro")) throw new IllegalStateException(detail.path("erro").asText());
            String title = detail.path("titulo").asText();
            BigDecimal price = detail.path("preco").decimalValue().setScale(2, RoundingMode.HALF_UP);
            if (constraints.maxPrice != null && price.compareTo(constraints.maxPrice) > 0)
                throw new IllegalStateException("O item " + title + " excede o limite de US$ " + constraints.maxPrice + " por unidade.");
            BigDecimal line = price.multiply(BigDecimal.valueOf(qty)).setScale(2, RoundingMode.HALF_UP);
            items.add(new Item(productId, title, qty, price, line)); totalQuantity += qty;
        }
        if (items.isEmpty()) throw new IllegalStateException("O agente não confirmou nenhum produto do catálogo.");
        if (constraints.quantity != null && totalQuantity != constraints.quantity)
            throw new IllegalStateException("A quantidade selecionada não corresponde ao pedido; a requisição não foi salva.");
        return new AgentResult(message, List.copyOf(items));
    }

    private Constraints parseConstraints(String text) {
        Matcher q = Pattern.compile("\\b(?:de|preciso de|quero|comprar)\\s+(\\d+)\\b", Pattern.CASE_INSENSITIVE).matcher(text);
        Matcher p = Pattern.compile("(?:até|no máximo|limite de)\\s*(?:US\\$|\\$)?\\s*([\\d.,]+)\\s*(?:cada|por unidade)?", Pattern.CASE_INSENSITIVE).matcher(text);
        Integer quantity = q.find() ? Integer.valueOf(q.group(1)) : null;
        BigDecimal price = null;
        if (p.find()) try { price = parseLocalizedAmount(p.group(1)); } catch (NumberFormatException ignored) {}
        return new Constraints(quantity, price);
    }

    private BigDecimal parseLocalizedAmount(String raw) {
        boolean hasDot = raw.contains("."), hasComma = raw.contains(",");
        String normalized = raw;
        if (hasDot && hasComma) {
            char decimalSeparator = raw.lastIndexOf('.') > raw.lastIndexOf(',') ? '.' : ',';
            char groupingSeparator = decimalSeparator == '.' ? ',' : '.';
            normalized = raw.replace(String.valueOf(groupingSeparator), "");
            if (decimalSeparator == ',') normalized = normalized.replace(',', '.');
        } else if (hasDot && raw.matches("\\d{1,3}(\\.\\d{3})+")) {
            normalized = raw.replace(".", "");
        } else if (hasComma && raw.matches("\\d{1,3}(,\\d{3})+")) {
            normalized = raw.replace(",", "");
        } else if (hasComma) {
            normalized = raw.replace(',', '.');
        }
        return new BigDecimal(normalized);
    }

    public List<Requisition> list() {
        List<Long> ids = jdbc.query("SELECT id FROM requisitions ORDER BY created_at DESC LIMIT 100", (rs, row) -> rs.getLong(1));
        return ids.stream().map(this::getOne).toList();
    }

    private Requisition getOne(long id) {
        return jdbc.queryForObject("SELECT id,request_text,response_text,total,created_at FROM requisitions WHERE id=?", (rs, row) -> {
            List<Item> items = jdbc.query("SELECT product_id,title,quantity,unit_price,line_total FROM requisition_items WHERE requisition_id=? ORDER BY id",
                    (itemRs, i) -> new Item(itemRs.getInt(1), itemRs.getString(2), itemRs.getInt(3), itemRs.getBigDecimal(4), itemRs.getBigDecimal(5)), id);
            return new Requisition(rs.getLong(1), rs.getString(2), rs.getString(3), rs.getBigDecimal(4), rs.getObject(5, OffsetDateTime.class), items);
        }, id);
    }

    public boolean llmConfigured() { return !apiKey.isBlank(); }
    private record Constraints(Integer quantity, BigDecimal maxPrice) {}
    private record AgentResult(String message, List<Item> items) {}
}
