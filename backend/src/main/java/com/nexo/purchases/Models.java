package com.nexo.purchases;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.List;

public final class Models {
    private Models() {}
    public record CreateRequest(String pedido) {}
    public record Item(int product_id, String title, int quantity, BigDecimal unit_price, BigDecimal line_total) {}
    public record Requisition(long id, String pedido, String mensagem, BigDecimal total, OffsetDateTime created_at, List<Item> itens) {}
    public record CreateResponse(String status, String mensagem, Requisition requisicao) {}
    public record Health(String status, boolean llm_configured) {}
}
