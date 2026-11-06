import React, { useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { ArrowUp, PackageSearch, ShoppingBag, Plus, Clock3, Check, Sparkles, Menu, X, CircleAlert } from 'lucide-react';
import './style.css';

const API = import.meta.env.VITE_API_URL || 'http://localhost:8000';
const money = value => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value || 0);

function App() {
  const [message, setMessage] = useState('');
  const [busy, setBusy] = useState(false);
  const [history, setHistory] = useState([]);
  const [active, setActive] = useState(null);
  const [items, setItems] = useState([]);
  const [toast, setToast] = useState('');
  const [drawer, setDrawer] = useState(false);
  const [ready, setReady] = useState(false);

  async function loadHistory() {
    try { const r = await fetch(`${API}/requisicoes`); if (r.ok) setHistory(await r.json()); }
    catch { /* API health state is shown separately */ }
  }
  useEffect(() => {
    fetch(`${API}/health`).then(r => r.json()).then(d => setReady(d.llm_configured)).catch(() => setReady(false));
    loadHistory();
  }, []);

  async function submit(text = message) {
    if (!text.trim() || busy) return;
    setBusy(true); setToast(''); setActive(null); setItems([]);
    try {
      const response = await fetch(`${API}/requisicoes`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ pedido: text.trim() }) });
      const data = await response.json();
      if (!response.ok) throw new Error(data.detail || data.message || data.error || 'Não foi possível processar o pedido.');
      setMessage(''); setActive({ ...data, pedido: text.trim() }); setItems(data.requisicao?.itens || []);
      if (data.requisicao) await loadHistory();
    } catch (e) { setToast(e.message); }
    finally { setBusy(false); }
  }
  function newRequest() { setActive(null); setItems([]); setMessage(''); setToast(''); setDrawer(false); }
  function showOld(row) { setActive({ status: 'ok', mensagem: row.mensagem, requisicao: row }); setItems(row.itens); setDrawer(false); }

  return <div className="app-shell">
    <aside className={`sidebar ${drawer ? 'open' : ''}`}>
      <div className="brand"><div className="brand-mark"><ShoppingBag size={18}/></div><span>FDE<span className="brand-dot">.</span></span><button className="close-drawer" onClick={() => setDrawer(false)}><X size={18}/></button></div>
      <button className="new-request" onClick={newRequest}><Plus size={16}/> Nova requisição</button>
      <div className="history-heading"><span>RECENTES</span><Clock3 size={14}/></div>
      <div className="history-list">
        {history.map(row => <button key={row.id} className={`history-item ${active?.requisicao?.id === row.id ? 'selected' : ''}`} onClick={() => showOld(row)}>
          <span className="history-icon"><ShoppingBag size={15}/></span><span className="history-copy"><strong>Requisição #{row.id}</strong><small>{row.pedido}</small></span>
        </button>)}
        {!history.length && <p className="empty-history">Suas requisições salvas aparecem aqui.</p>}
      </div>
      <div className="sidebar-bottom"><span className={`status-dot ${ready ? 'online' : ''}`}></span><span>{ready ? 'Agente conectado' : 'Configure sua chave de API'}</span></div>
    </aside>
    {drawer && <button className="scrim" aria-label="Fechar menu" onClick={() => setDrawer(false)} />}
    <main className="main">
      <header className="topbar"><button className="menu-button" onClick={() => setDrawer(true)}><Menu size={20}/></button><div className="breadcrumb">Workspace <span>/</span> <strong>Compras</strong></div><div className="avatar">FD</div></header>
      <section className={`conversation ${active ? 'has-result' : ''}`}>
        {!active ? <div className="welcome">
          <div className="welcome-icon"><PackageSearch size={22}/></div>
          <div className="eyebrow"><Sparkles size={13}/> ASSISTENTE DE COMPRAS</div>
          <h1>O que sua equipe<br/><span>precisa hoje?</span></h1>
          <p>Descreva o que você procura. Eu busco no catálogo e organizo a requisição para você.</p>
          <div className="suggestions"><span>EXPERIMENTE</span><button onClick={() => submit('Preciso de 3 laptops para o time de vendas, até US$ 2.000 cada.')}>3 laptops para vendas <span>↗</span></button><button onClick={() => { setMessage('Preciso de 2 smartphones para a equipe'); document.querySelector('textarea')?.focus(); }}>Smartphones para a equipe <span>↗</span></button></div>
        </div> : <div className="result-wrap">
          <div className="request-bubble"><span>SEU PEDIDO</span><p>{active.requisicao?.pedido || active.pedido || message || history.find(h => h.id === active.requisicao?.id)?.pedido}</p></div>
          <div className="agent-answer"><div className="agent-avatar"><Sparkles size={16}/></div><div className="answer-content"><span className="answer-label">FDE <i>·</i> AGENTE DE COMPRAS</span><p>{active.mensagem}</p>
          {active.status === 'ok' && active.requisicao && <div className="receipt"><div className="receipt-head"><div><Check size={15}/><strong>Requisição #{active.requisicao.id}</strong></div><span>REGISTRADA</span></div>
          {items.map((item, i) => <div className="line-item" key={item.product_id + '-' + i}><div className="item-symbol"><PackageSearch size={17}/></div><div className="item-info"><strong>{item.title}</strong><small>{item.quantity} un. × {money(item.unit_price)}</small></div><b>{money(item.line_total)}</b></div>)}
          <div className="receipt-total"><span>Total estimado</span><strong>{money(active.requisicao.total)}</strong></div><small className="receipt-note">Valores consultados no catálogo DummyJSON · USD</small></div>}
          {active.status === 'nao_atendido' && <div className="not-found"><CircleAlert size={16}/> Nenhuma requisição foi criada</div>}
          </div></div>
          </div>}
      </section>
      <div className="composer-area">
        {toast && <div className="toast"><CircleAlert size={16}/>{toast}<button onClick={() => setToast('')}>×</button></div>}
        <form className="composer" onSubmit={e => { e.preventDefault(); submit(); }}><textarea value={message} onChange={e => setMessage(e.target.value)} placeholder="Ex.: Preciso de 3 laptops para o time de vendas, até US$ 2.000 cada." rows="1" onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); submit(); } }}/><div className="composer-foot"><span><span className="tiny-spark"><Sparkles size={12}/></span> Buscas verificadas no catálogo</span><button type="submit" disabled={busy || !message.trim()} aria-label="Enviar pedido">{busy ? <span className="loader"/> : <ArrowUp size={18}/>}</button></div></form>
        <div className="disclaimer">O agente confirma produtos e preços no catálogo antes de registrar cada requisição.</div>
      </div>
    </main>
  </div>;
}

createRoot(document.getElementById('root')).render(<App/>);
