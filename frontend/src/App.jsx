
import React, { useState, useRef, useEffect } from 'react';
import {
  Send,
  ShoppingBag,
  Plus,
  Package,
  CheckCircle2,
  Bot,
  User,
  Loader2,
  Trash2
} from 'lucide-react';

export default function App() {
  const [mensagem, setMensagem] = useState('');
  const [loading, setLoading] = useState(false);
  const [historico, setHistorico] = useState([]);
  const [conversa, setConversa] = useState([]);

  const mensagensEndRef = useRef(null);
  const inputRef = useRef(null);

  const sugestoes = [
    "Preciso de 3 smartphones para a equipe",
    "Preciso de 2 laptops para o time, até US$ 1500 cada",
    "Quero cotar 1 tablet para o design"
  ];

  useEffect(() => {
    mensagensEndRef.current?.scrollIntoView({
      behavior: 'smooth'
    });
  }, [conversa, loading]);

  const enviarRequisicao = async (textoInput) => {
    const texto = (textoInput || mensagem).trim();

    if (!texto || loading) return;

    const mensagemUsuario = {
      id: Date.now(),
      tipo: 'usuario',
      texto
    };

    setConversa(prev => [...prev, mensagemUsuario]);
    setMensagem('');
    setLoading(true);

    try {
      const response = await fetch(
        'http://localhost:8080/requisicoes',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            mensagem: texto
          })
        }
      );

      if (!response.ok) {
      const erro = await response.text();
      throw new Error(`API Go (${response.status}): ${erro}`);
      } 

      const data = await response.json();

      setHistorico(prev => [data, ...prev]);

      const respostaAgente = {
        id: data.id,
        tipo: 'agente',
        texto: data.resposta,
        status: data.status,
        itens: data.itens || [],
        requisicao: data
      };

      setConversa(prev => [...prev, respostaAgente]);

    } catch (err) {
      setConversa(prev => [
        ...prev,
        {
          id: Date.now(),
          tipo: 'erro',
          texto: `Não foi possível processar sua solicitação: ${err.message}`
        }
      ]);
    } finally {
      setLoading(false);
      inputRef.current?.focus();
    }
  };

  const novaConversa = () => {
    if (loading) return;

    setConversa([]);
    setMensagem('');
    inputRef.current?.focus();
  };

  const formatarValor = (valor) => {
    return Number(valor || 0).toLocaleString('en-US', {
      style: 'currency',
      currency: 'USD'
    });
  };

  const totalItens = historico.reduce((acc, req) => {
    return acc + (req.itens || []).reduce(
      (sum, item) => sum + item.quantidade,
      0
    );
  }, 0);

  return (
    <div className="h-screen bg-[#212121] text-white flex overflow-hidden font-sans">

      {/* SIDEBAR */}
      <aside className="hidden md:flex w-64 bg-[#171717] flex-col border-r border-white/5">

        <div className="p-4">
          <button
            onClick={novaConversa}
            className="w-full flex items-center gap-3 p-3 rounded-xl border border-white/15 hover:bg-white/10 transition text-sm"
          >
            <Plus size={18} />
            Nova conversa
          </button>
        </div>

        <div className="px-4 pt-4">
          <p className="text-xs text-gray-500 font-semibold mb-3">
            COMPRAS
          </p>

          <div className="flex items-center gap-3 p-2 text-sm text-gray-300">
            <ShoppingBag size={17} />
            Procurement Agent
          </div>

          <div className="flex items-center gap-3 p-2 text-sm text-gray-400">
            <Package size={17} />
            {historico.length} requisições processadas
          </div>
        </div>

        <div className="mt-6 px-4 flex-1 overflow-y-auto">
          <p className="text-xs text-gray-500 font-semibold mb-3">
            HISTÓRICO RECENTE
          </p>

          {historico.map((req) => (
            <div
              key={req.id}
              className="p-2 text-xs text-gray-400 truncate rounded-lg hover:bg-white/5"
            >
              {req.mensagem}
            </div>
          ))}
        </div>

        <div className="p-4 border-t border-white/10">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-indigo-600 rounded-full flex items-center justify-center">
              <Bot size={17} />
            </div>

            <div>
              <p className="text-sm font-medium">
                FDE Procurement
              </p>
              <p className="text-xs text-gray-500">
                Agente de compras inteligente
              </p>
            </div>
          </div>
        </div>

      </aside>

      {/* CHAT PRINCIPAL */}
      <main className="flex-1 flex flex-col min-w-0">

        {/* HEADER */}
        <header className="h-14 border-b border-white/10 flex items-center justify-between px-5 shrink-0">

          <div className="flex items-center gap-2">
            <ShoppingBag size={19} className="text-indigo-400" />

            <span className="font-semibold text-sm">
              FDE Procurement Agent
            </span>

            <span className="hidden sm:inline text-xs text-gray-500">
              / Assistente de compras
            </span>
          </div>

          <div className="flex items-center gap-2 text-xs text-emerald-400">
            <span className="w-2 h-2 rounded-full bg-emerald-400" />
            Sistema online
          </div>

        </header>

        {/* ÁREA DE MENSAGENS */}
        <section className="flex-1 overflow-y-auto">

          <div className="max-w-3xl mx-auto px-4 py-8">

            {/* TELA INICIAL */}
            {conversa.length === 0 && (
              <div className="min-h-[65vh] flex flex-col items-center justify-center text-center">

                <div className="w-16 h-16 rounded-2xl bg-indigo-600/20 border border-indigo-500/30 flex items-center justify-center mb-6">
                  <ShoppingBag
                    size={30}
                    className="text-indigo-400"
                  />
                </div>

                <h1 className="text-3xl font-semibold mb-3">
                  Como posso ajudar?
                </h1>

                <p className="text-gray-400 text-sm max-w-md mb-10">
                  Sou seu agente inteligente de compras.
                  Descreva o que sua equipe precisa e vou
                  processar sua requisição.
                </p>

                <div className="grid sm:grid-cols-2 gap-3 w-full max-w-xl">

                  {sugestoes.map((sug, index) => (
                    <button
                      key={index}
                      onClick={() => enviarRequisicao(sug)}
                      className="text-left p-4 rounded-xl border border-white/10 hover:bg-white/5 transition text-sm text-gray-300"
                    >
                      <ShoppingBag
                        size={17}
                        className="text-indigo-400 mb-3"
                      />

                      {sug}
                    </button>
                  ))}

                </div>
              </div>
            )}

            {/* MENSAGENS */}
            <div className="space-y-8">

              {conversa.map((msg) => (

                <div
                  key={msg.id}
                  className={`flex gap-4 ${
                    msg.tipo === 'usuario'
                      ? 'justify-end'
                      : 'justify-start'
                  }`}
                >

                  {/* ÍCONE DO AGENTE */}
                  {msg.tipo !== 'usuario' && (
                    <div className="w-8 h-8 rounded-full bg-indigo-600 flex items-center justify-center shrink-0">
                      {msg.tipo === 'erro'
                        ? <Bot size={17} />
                        : <Bot size={17} />
                      }
                    </div>
                  )}

                  <div className={`max-w-[85%] min-w-0 ${
                    msg.tipo === 'usuario'
                      ? 'bg-[#303030] rounded-2xl px-4 py-3'
                      : 'pt-1 flex-1'
                  }`}>

                    <p className="text-sm leading-7 whitespace-pre-wrap">
                      {msg.texto}
                    </p>

                    {/* STATUS DA REQUISIÇÃO */}
                    {msg.tipo === 'agente' && (
                      <div className="mt-4">

                        <div className="flex items-center gap-2 mb-4">
                          <CheckCircle2
                            size={15}
                            className="text-emerald-400"
                          />

                          <span className="text-xs text-emerald-400">
                            {msg.status}
                          </span>

                          <span className="text-xs text-gray-600">
                            | ID #{msg.id}
                          </span>
                        </div>

                        {/* PRODUTOS */}
                        {msg.itens.length > 0 && (
                          <div className="border border-white/10 rounded-xl overflow-hidden">

                            <div className="bg-[#2a2a2a] p-3 flex items-center gap-2">
                              <Package
                                size={16}
                                className="text-indigo-400"
                              />

                              <span className="text-sm font-medium">
                                Itens da requisição
                              </span>
                            </div>

                            <div className="divide-y divide-white/10">

                              {msg.itens.map((item, index) => (
                                <div
                                  key={index}
                                  className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
                                >

                                  <div>
                                    <p className="text-sm font-medium">
                                      {item.nome_produto}
                                    </p>

                                    <p className="text-xs text-gray-500 mt-1">
                                      Código: #{item.produto_id}
                                    </p>
                                  </div>

                                  <div className="flex items-center justify-between sm:gap-8 text-xs">

                                    <span className="text-gray-400">
                                      Qtd: {item.quantidade}
                                    </span>

                                    <span className="text-gray-400">
                                      {formatarValor(item.preco_unitario)}
                                    </span>

                                    <span className="text-emerald-400 font-semibold">
                                      {formatarValor(item.valor_total)}
                                    </span>

                                  </div>

                                </div>
                              ))}

                            </div>

                            <div className="bg-[#2a2a2a] p-4 flex justify-between items-center">
                              <span className="text-sm text-gray-400">
                                Total estimado
                              </span>

                              <span className="text-base font-semibold text-emerald-400">
                                {formatarValor(
                                  msg.itens.reduce(
                                    (sum, item) => sum + Number(item.valor_total || 0),
                                    0
                                  )
                                )}
                              </span>
                            </div>

                          </div>
                        )}

                      </div>
                    )}

                    {msg.tipo === 'erro' && (
                      <p className="text-xs text-red-400 mt-3">
                        Verifique se o backend Go está disponível.
                      </p>
                    )}

                  </div>

                  {/* ÍCONE DO USUÁRIO */}
                  {msg.tipo === 'usuario' && (
                    <div className="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center shrink-0">
                      <User size={17} />
                    </div>
                  )}

                </div>

              ))}

              {/* LOADING */}
              {loading && (
                <div className="flex gap-4 items-start">

                  <div className="w-8 h-8 rounded-full bg-indigo-600 flex items-center justify-center shrink-0">
                    <Bot size={17} />
                  </div>

                  <div className="flex items-center gap-2 text-sm text-gray-400 pt-2">
                    <Loader2
                      size={16}
                      className="animate-spin text-indigo-400"
                    />
                    Analisando sua solicitação...
                  </div>

                </div>
              )}

              <div ref={mensagensEndRef} />

            </div>
          </div>

        </section>

        {/* CAMPO DE MENSAGEM */}
        <footer className="px-4 pb-5 pt-3 shrink-0">

          <div className="max-w-3xl mx-auto">

            <form
              onSubmit={(e) => {
                e.preventDefault();
                enviarRequisicao();
              }}
              className="relative bg-[#303030] border border-white/10 rounded-2xl p-2 flex items-end gap-2 focus-within:border-white/25 transition"
            >

              <textarea
                ref={inputRef}
                value={mensagem}
                onChange={(e) => setMensagem(e.target.value)}
                onKeyDown={(e) => {
                  if (
                    e.key === 'Enter' &&
                    !e.shiftKey
                  ) {
                    e.preventDefault();
                    enviarRequisicao();
                  }
                }}
                placeholder="Descreva o que você precisa comprar..."
                rows={1}
                disabled={loading}
                className="flex-1 bg-transparent resize-none outline-none text-sm text-white placeholder-gray-500 px-3 py-3 max-h-36"
              />

              <button
                type="submit"
                disabled={loading || !mensagem.trim()}
                className="w-10 h-10 rounded-xl bg-white text-black flex items-center justify-center hover:bg-gray-200 disabled:opacity-30 transition shrink-0"
              >
                {loading
                  ? <Loader2 size={18} className="animate-spin" />
                  : <Send size={18} />
                }
              </button>

            </form>

            <p className="text-center text-[10px] text-gray-500 mt-3">
              FDE Procurement Agent pode cometer erros.
              Verifique os valores antes de aprovar compras.
            </p>

          </div>

        </footer>

      </main>

    </div>
  );
}
