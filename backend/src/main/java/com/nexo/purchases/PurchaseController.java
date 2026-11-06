package com.nexo.purchases;

import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.web.bind.annotation.*;
import static com.nexo.purchases.Models.*;
import java.util.List;

@RestController
public class PurchaseController {
    private final PurchaseService service;
    private final JdbcTemplate jdbc;
    public PurchaseController(PurchaseService service, JdbcTemplate jdbc) { this.service=service; this.jdbc=jdbc; }
    @GetMapping("/health") public Health health() { jdbc.queryForObject("SELECT 1", Integer.class); return new Health("ok", service.llmConfigured()); }
    @PostMapping("/requisicoes") public CreateResponse create(@RequestBody CreateRequest body) { return service.create(body.pedido()); }
    @GetMapping("/requisicoes") public List<Requisition> list() { return service.list(); }
}
