// ==========================================================================
// Chatwoot Lite UI Application Script
// ==========================================================================

// Fluent UI System Icons (SVG Paths)
const ICONS = {
  chat: '<svg viewBox="0 0 24 24"><path d="M12 2c5.523 0 10 4.477 10 10s-4.477 10-10 10a9.96 9.96 0 0 1-4.587-1.112l-3.826 1.067a1.25 1.25 0 0 1-1.54-1.54l1.068-3.823A9.96 9.96 0 0 1 2 12C2 6.477 6.477 2 12 2Zm0 1.5A8.5 8.5 0 0 0 3.5 12c0 1.47.373 2.883 1.073 4.137l.15.27-1.112 3.984 3.987-1.112.27.15A8.5 8.5 0 1 0 12 3.5ZM8.75 13h4.498a.75.75 0 0 1 .102 1.493l-.102.007H8.75a.75.75 0 0 1-.102-1.493L8.75 13h4.498H8.75Zm0-3.5h6.505a.75.75 0 0 1 .101 1.493l-.101.007H8.75a.75.75 0 0 1-.102-1.493L8.75 9.5h6.505H8.75Z"/></svg>',
  search: '<svg viewBox="0 0 24 24"><path d="M10 2.75a7.25 7.25 0 0 1 5.63 11.819l4.9 4.9a.75.75 0 0 1-.976 1.134l-.084-.073-4.901-4.9A7.25 7.25 0 1 1 10 2.75Zm0 1.5a5.75 5.75 0 1 0 0 11.5 5.75 5.75 0 0 0 0-11.5Z"/></svg>',
  send: '<svg viewBox="0 0 24 24"><path d="M5.694 12 2.299 3.272c-.236-.607.356-1.188.942-.982l.093.04 18 9a.75.75 0 0 1 .097 1.283l-.097.058-18 9c-.583.291-1.217-.244-1.065-.847l.03-.096L5.694 12 2.299 3.272 5.694 12ZM4.402 4.54l2.61 6.71h6.627a.75.75 0 0 1 .743.648l.007.102a.75.75 0 0 1-.649.743l-.101.007H7.01l-2.609 6.71L19.322 12 4.401 4.54Z"/></svg>',
  settings: '<svg viewBox="0 0 24 24"><path d="M12.012 2.25c.734.008 1.465.093 2.182.253a.75.75 0 0 1 .582.649l.17 1.527a1.384 1.384 0 0 0 1.927 1.116l1.4-.615a.75.75 0 0 1 .85.174a9.8 9.8 0 0 1 2.205 3.792a.75.75 0 0 1-.272.825l-1.241.916a1.38 1.38 0 0 0 0 2.226l1.243.915a.75.75 0 0 1 .272.826a9.8 9.8 0 0 1-2.204 3.792a.75.75 0 0 1-.849.175l-1.406-.617a1.38 1.38 0 0 0-1.926 1.114l-.17 1.526a.75.75 0 0 1-.571.647a9.5 9.5 0 0 1-4.406 0a.75.75 0 0 1-.572-.647l-.169-1.524a1.382 1.382 0 0 0-1.925-1.11l-1.406.616a.75.75 0 0 1-.85-.175a9.8 9.8 0 0 1-2.203-3.796a.75.75 0 0 1 .272-.826l1.243-.916a1.38 1.38 0 0 0 0-2.226l-1.243-.914a.75.75 0 0 1-.272-.826a9.8 9.8 0 0 1 2.205-3.792a.75.75 0 0 1 .85-.174l1.4.615a1.387 1.387 0 0 0 1.93-1.118l.17-1.526a.75.75 0 0 1 .583-.65q1.074-.238 2.201-.252M12 9a3 3 0 1 0 0 6a3 3 0 0 0 0-6"/></svg>',
  check: '<svg viewBox="0 0 24 24"><path d="M4.53 12.97a.75.75 0 0 0-1.06 1.06l4.5 4.5a.75.75 0 0 0 1.06 0l11-11a.75.75 0 0 0-1.06-1.06L8.5 16.94l-3.97-3.97Z"/></svg>',
  doubleCheck: '<svg viewBox="0 0 24 24"><path d="M1.53 12.97a.75.75 0 0 0-1.06 1.06l4.5 4.5a.75.75 0 0 0 1.06 0l11-11a.75.75 0 0 0-1.06-1.06L5.5 16.94l-3.97-3.97Zm6.5 0a.75.75 0 0 0-1.06 1.06l4.5 4.5a.75.75 0 0 0 1.06 0l11-11a.75.75 0 0 0-1.06-1.06L12 16.94l-3.97-3.97Z"/></svg>',
  checkCircle: '<svg viewBox="0 0 24 24"><path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20Zm-1.25 13.72-3.47-3.47a.75.75 0 1 1 1.06-1.06l2.41 2.41 5.72-5.72a.75.75 0 1 1 1.06 1.06l-6.25 6.25a.75.75 0 0 1-.53.53Z"/></svg>',
  alertCircle: '<svg viewBox="0 0 24 24"><path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20Zm0 5a.75.75 0 0 1 .75.75v5.5a.75.75 0 0 1-1.5 0v-5.5A.75.75 0 0 1 12 7Zm0 9a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"/></svg>',
  copy: '<svg viewBox="0 0 24 24"><path d="M7.5 2A2.5 2.5 0 0 0 5 4.5v11A2.5 2.5 0 0 0 7.5 18h9a2.5 2.5 0 0 0 2.5-2.5v-11A2.5 2.5 0 0 0 16.5 2h-9ZM6.5 4.5A1 1 0 0 1 7.5 3.5h9a1 1 0 0 1 1 1v11a1 1 0 0 1-1 1h-9a1 1 0 0 1-1-1v-11ZM4 6.5A1.5 1.5 0 0 0 2.5 8v11A2.5 2.5 0 0 0 5 21.5h10a1.5 1.5 0 0 0 1.5-1.5v-.5h-10A1.5 1.5 0 0 1 5 18V6.5h-1Z"/></svg>',
  image: '<svg viewBox="0 0 24 24"><path d="M4.5 3A2.5 2.5 0 0 0 2 5.5v13A2.5 2.5 0 0 0 4.5 21h15a2.5 2.5 0 0 0 2.5-2.5v-13A2.5 2.5 0 0 0 19.5 3h-15ZM3.5 5.5a1 1 0 0 1 1-1h15a1 1 0 0 1 1 1v8.88l-4.72-4.72a1.75 1.75 0 0 0-2.47 0L7 16.03l-1.78-1.78a1.75 1.75 0 0 0-2.47 0L3.5 15v-9.5Zm4.75 4.5a1.75 1.75 0 1 1 0-3.5 1.75 1.75 0 0 1 0 3.5Z"/></svg>',
  document: '<svg viewBox="0 0 24 24"><path d="M6.5 2A2.5 2.5 0 0 0 4 4.5v15A2.5 2.5 0 0 0 6.5 22h11a2.5 2.5 0 0 0 2.5-2.5V8.5a1 1 0 0 0-.29-.7l-5.5-5.5a1 1 0 0 0-.71-.3H6.5ZM5.5 4.5a1 1 0 0 1 1-1H13v4a1.5 1.5 0 0 0 1.5 1.5h4v10.5a1 1 0 0 1-1 1h-11a1 1 0 0 1-1-1v-15ZM8 12.75a.75.75 0 0 1 .75-.75h6.5a.75.75 0 0 1 0 1.5h-6.5a.75.75 0 0 1-.75-.75Zm0 3.5a.75.75 0 0 1 .75-.75h6.5a.75.75 0 0 1 0 1.5h-6.5a.75.75 0 0 1-.75-.75Z"/></svg>',
  location: '<svg viewBox="0 0 24 24"><path d="M12 2a7.5 7.5 0 0 0-7.5 7.5c0 5.25 6.75 12 7.05 12.3a.6.6 0 0 0 .9 0c.3-.3 7.05-7.05 7.05-12.3A7.5 7.5 0 0 0 12 2Zm0 10.5a3 3 0 1 1 0-6 3 3 0 0 1 0 6Z"/></svg>',
  contact: '<svg viewBox="0 0 24 24"><path d="M12 2a5 5 0 1 0 0 10 5 5 0 0 0 0-10Zm0 1.5a3.5 3.5 0 1 1 0 7 3.5 3.5 0 0 1 0-7Zm-7 13.5A3.5 3.5 0 0 1 8.5 13.5h7a3.5 3.5 0 0 1 3.5 3.5v1.5a1.5 1.5 0 0 1-1.5 1.5h-11A1.5 1.5 0 0 1 5 18.5V17Zm1.5 0a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1.5H6.5V17Z"/></svg>',
  audio: '<svg viewBox="0 0 24 24"><path d="M12 2a3.5 3.5 0 0 0-3.5 3.5v6a3.5 3.5 0 0 0 7 0v-6A3.5 3.5 0 0 0 12 2Zm2 9.5a2 2 0 1 1-4 0v-6a2 2 0 1 1 4 0v6Zm-7-3.5a.75.75 0 0 1 .75.75 4.25 4.25 0 0 0 8.5 0 .75.75 0 0 1 1.5 0 5.75 5.75 0 0 1-5 5.69v2.31h2a.75.75 0 0 1 0 1.5h-5.5a.75.75 0 0 1 0-1.5h2v-2.31a5.75 5.75 0 0 1-5-5.69.75.75 0 0 1 .75-.75Z"/></svg>',
  video: '<svg viewBox="0 0 24 24"><path d="M3.5 4A2.5 2.5 0 0 0 1 6.5v11A2.5 2.5 0 0 0 3.5 20h11a2.5 2.5 0 0 0 2.5-2.5v-2.17l3.8 2.53A1.25 1.25 0 0 0 23 16.82V7.18a1.25 1.25 0 0 0-2.2-.84L17 8.67V6.5A2.5 2.5 0 0 0 14.5 4h-11Zm0 1.5h11a1 1 0 0 1 1 1v11a1 1 0 0 1-1 1h-11a1 1 0 0 1-1-1v-11a1 1 0 0 1 1-1Zm15 4.38 3-2v8.24l-3-2V9.88Z"/></svg>'
};

// Soft Pastel Avatar Colors
const AVATAR_COLORS = [
  { bg: '#C4B5FD', text: '#4C1D95' }, // Purple
  { bg: '#93C5FD', text: '#1E40AF' }, // Blue
  { bg: '#86EFAC', text: '#14532D' }, // Green
  { bg: '#FCA5A5', text: '#7F1D1D' }, // Red
  { bg: '#FDE68A', text: '#78350F' }, // Yellow
  { bg: '#FBCFE8', text: '#831843' }, // Pink
  { bg: '#A7F3D0', text: '#064E3B' }, // Teal
  { bg: '#FED7AA', text: '#7C2D12' }  // Orange
];

function getAvatarStyle(name) {
  let hash = 0;
  const str = name || 'User';
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  const color = AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
  return `background-color: ${color.bg}; color: ${color.text};`;
}

function getInitials(name) {
  if (!name) return 'C';
  const clean = name.replace(/[^\w\s]/gi, '').trim();
  if (!clean) return 'W';
  const parts = clean.split(/\s+/);
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return clean.substring(0, 2).toUpperCase();
}

// Global State
let currentConversationId = null;
let conversationsList = [];
let pollInterval = null;
let apiAccessToken = '';

async function authFetch(url, options = {}) {
  options.headers = options.headers || {};
  if (apiAccessToken) {
    if (options.headers instanceof Headers) {
      options.headers.set('api_access_token', apiAccessToken);
    } else {
      options.headers['api_access_token'] = apiAccessToken;
    }
  }
  return fetch(url, options);
}

// Initialize on DOM Ready
document.addEventListener('DOMContentLoaded', async () => {
  initTabs();
  await loadMetaConfig();
  loadConversations();
  loadHealthStatus();
  loadWebhooks();

  // Search filter
  document.getElementById('conv-search').addEventListener('input', (e) => {
    filterConversations(e.target.value);
  });

  // Message submission form
  document.getElementById('message-form').addEventListener('submit', handleSendMessage);
  const msgInput = document.getElementById('message-input');
  msgInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage(e);
    }
  });

  // Webhook creation form
  document.getElementById('new-webhook-form').addEventListener('submit', handleCreateWebhook);

  // Poll for updates every 3 seconds
  pollInterval = setInterval(() => {
    loadConversations(true);
    if (currentConversationId) {
      loadMessages(currentConversationId, true);
    }
  }, 3000);
});

// Tab Navigation
function initTabs() {
  const btnConvs = document.getElementById('tab-convs-btn');
  const btnWebhooks = document.getElementById('tab-webhooks-btn');
  const viewConvs = document.getElementById('view-conversations');
  const viewWebhooks = document.getElementById('view-webhooks');

  btnConvs.addEventListener('click', () => {
    btnConvs.classList.add('active');
    btnWebhooks.classList.remove('active');
    viewConvs.classList.add('active');
    viewWebhooks.classList.remove('active');
  });

  btnWebhooks.addEventListener('click', () => {
    btnWebhooks.classList.add('active');
    btnConvs.classList.remove('active');
    viewWebhooks.classList.add('active');
    viewConvs.classList.remove('active');
    loadHealthStatus();
    loadMetaConfig();
    loadWebhooks();
  });
}

// Load Conversations List
async function loadConversations(isBackground = false) {
  try {
    const res = await authFetch('/api/conversations');
    if (!res.ok) return;
    const data = await res.json();
    conversationsList = data.data || [];
    renderConversations(conversationsList);
  } catch (err) {
    if (!isBackground) {
      console.error('Failed to load conversations:', err);
    }
  }
}

function renderConversations(list) {
  const container = document.getElementById('conversations-list');
  if (list.length === 0) {
    container.innerHTML = '<div class="empty-list-state">Nenhuma conversa encontrada</div>';
    return;
  }

  container.innerHTML = list.map(c => {
    const contactName = c.contact ? (c.contact.name || c.contact.phone_number) : 'Desconhecido';
    const lastSnippet = c.last_message ? (c.last_message.content || 'Anexo recebido') : 'Nova conversa';
    const timeStr = formatTime(c.last_activity_at || c.created_at);
    const activeClass = (c.id === currentConversationId) ? 'active' : '';
    const initials = getInitials(contactName);
    const avatarStyle = getAvatarStyle(contactName);

    return `
      <div class="conv-item ${activeClass}" onclick="selectConversation(${c.id})">
        <div class="avatar" style="${avatarStyle}">${initials}</div>
        <div class="conv-item-content">
          <div class="conv-item-top">
            <span class="conv-name">${escapeHtml(contactName)}</span>
            <span class="conv-time">${timeStr}</span>
          </div>
          <div class="conv-snippet">${escapeHtml(lastSnippet)}</div>
        </div>
      </div>
    `;
  }).join('');
}

function filterConversations(term) {
  term = term.toLowerCase().trim();
  const filtered = conversationsList.filter(c => {
    const name = (c.contact && c.contact.name) ? c.contact.name.toLowerCase() : '';
    const phone = (c.contact && c.contact.phone_number) ? c.contact.phone_number.toLowerCase() : '';
    return name.includes(term) || phone.includes(term);
  });
  renderConversations(filtered);
}

// Select Conversation
async function selectConversation(id) {
  currentConversationId = id;
  renderConversations(conversationsList);

  const emptyState = document.getElementById('empty-chat-state');
  const activeContainer = document.getElementById('active-chat-container');
  emptyState.style.display = 'none';
  activeContainer.classList.remove('hidden');

  const conv = conversationsList.find(c => c.id === id);
  if (conv && conv.contact) {
    const contactName = conv.contact.name || conv.contact.phone_number;
    document.getElementById('current-contact-name').textContent = contactName;
    document.getElementById('current-contact-phone').textContent = conv.contact.phone_number;
    
    const avatarEl = document.getElementById('current-avatar');
    avatarEl.textContent = getInitials(contactName);
    avatarEl.style = getAvatarStyle(contactName);

    const statusEl = document.getElementById('current-conv-status');
    if (conv.status === 1) {
      statusEl.textContent = 'Resolvida';
      statusEl.className = 'badge-status resolved';
    } else if (conv.status === 3) {
      statusEl.textContent = 'Adiada';
      statusEl.className = 'badge-status snoozed';
    } else {
      statusEl.textContent = 'Aberta';
      statusEl.className = 'badge-status open';
    }
  }

  await loadMessages(id);
}

// Load Messages for a Conversation
async function loadMessages(convId, isBackground = false) {
  try {
    const res = await authFetch(`/api/conversations/${convId}/messages`);
    if (!res.ok) return;
    const data = await res.json();
    const messages = data.data || [];

    const container = document.getElementById('messages-container');
    const wasAtBottom = container.scrollHeight - container.scrollTop <= container.clientHeight + 100;

    container.innerHTML = messages.map(m => {
      const isOutgoing = m.message_type === 1; // Outgoing
      const bubbleClass = isOutgoing ? 'outgoing' : 'incoming';
      const timeStr = formatTime(m.created_at);
      const statusIcon = isOutgoing ? getStatusIcon(m.status) : '';

      let attachmentsHtml = '';
      if (m.attachments && m.attachments.length > 0) {
        attachmentsHtml = m.attachments.map(att => renderAttachment(att)).join('');
      }

      return `
        <div class="message-bubble ${bubbleClass}">
          ${m.content ? `<div class="msg-content">${escapeHtml(m.content)}</div>` : ''}
          ${attachmentsHtml}
          <div class="msg-meta">
            <span>${timeStr}</span>
            ${statusIcon}
          </div>
        </div>
      `;
    }).join('');

    if (!isBackground || wasAtBottom) {
      container.scrollTop = container.scrollHeight;
    }
  } catch (err) {
    if (!isBackground) console.error('Failed to load messages:', err);
  }
}

function renderAttachment(att) {
  switch (att.file_type) {
    case 0: // Image
      return `
        <div class="attachment-preview">
          ${ICONS.image}
          <img src="${escapeHtml(att.external_url)}" style="max-width:220px; border-radius:6px; margin-top:4px;" alt="Imagem" onerror="this.outerHTML='<span>Imagem anexada</span>'">
        </div>`;
    case 1: // Audio
      return `<div class="attachment-preview">${ICONS.audio} <span>Mensagem de voz</span></div>`;
    case 2: // Video
      return `<div class="attachment-preview">${ICONS.video} <span>Vídeo anexado</span></div>`;
    case 3: // File
      return `<div class="attachment-preview">${ICONS.document} <span>${escapeHtml(att.fallback_title || 'Documento')}</span></div>`;
    case 4: // Location
      return `<div class="attachment-preview">${ICONS.location} <span>Localização (${att.coordinates_lat}, ${att.coordinates_long})</span></div>`;
    case 7: // Contact
      return `<div class="attachment-preview">${ICONS.contact} <span>Contato: ${escapeHtml(att.fallback_title || 'Contato')}</span></div>`;
    default:
      return `<div class="attachment-preview">${ICONS.document} <span>Arquivo anexado</span></div>`;
  }
}

// Send Message
async function handleSendMessage(e) {
  e.preventDefault();
  if (!currentConversationId) return;

  const input = document.getElementById('message-input');
  const text = input.value.trim();
  if (!text) return;

  input.value = '';
  input.focus();

  try {
    const res = await authFetch(`/api/conversations/${currentConversationId}/messages`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: text })
    });

    if (res.ok) {
      await loadMessages(currentConversationId);
      await loadConversations(true);
    } else {
      const err = await res.json();
      alert('Erro ao enviar mensagem: ' + (err.error || 'Desconhecido'));
    }
  } catch (err) {
    console.error('Error sending message:', err);
  }
}

// Load Health Status
async function loadHealthStatus() {
  try {
    const res = await fetch('/ready');
    const data = await res.json();
    const pgElem = document.getElementById('health-postgres');
    const rdElem = document.getElementById('health-redis');
    const statusDot = document.getElementById('system-status-dot');

    if (data.checks) {
      pgElem.innerHTML = data.checks.postgres === 'connected' 
        ? `<span class="metric-status-ok">${ICONS.checkCircle} Conectado</span>` 
        : `<span style="color:#EF4444;">${ICONS.alertCircle} Erro</span>`;
        
      rdElem.innerHTML = data.checks.redis === 'connected' 
        ? `<span class="metric-status-ok">${ICONS.checkCircle} Conectado</span>` 
        : `<span style="color:#EF4444;">${ICONS.alertCircle} Erro</span>`;
    }

    if (data.status === 'ready') {
      statusDot.className = 'status-indicator';
      statusDot.style.backgroundColor = 'var(--green-resolved)';
    } else {
      statusDot.className = 'status-indicator';
      statusDot.style.backgroundColor = '#EF4444';
    }
  } catch (err) {
    console.error('Health check error:', err);
  }
}

// Load Webhooks
async function loadWebhooks() {
  try {
    const res = await authFetch('/api/webhooks');
    if (!res.ok) return;
    const data = await res.json();
    const tbody = document.getElementById('webhooks-table-body');
    const webhooks = data.data || [];

    if (webhooks.length === 0) {
      tbody.innerHTML = '<tr><td colspan="4" class="text-center">Nenhum webhook registrado</td></tr>';
      return;
    }

    tbody.innerHTML = webhooks.map(w => {
      let subsStr = 'Todos';
      if (w.subscriptions) {
        try {
          const subs = typeof w.subscriptions === 'string' ? JSON.parse(w.subscriptions) : w.subscriptions;
          subsStr = subs.join(', ');
        } catch (e) {}
      }
      return `
        <tr>
          <td><strong>${escapeHtml(w.name || 'Sem nome')}</strong></td>
          <td><code>${escapeHtml(w.url)}</code></td>
          <td><span class="badge-channel">${escapeHtml(subsStr)}</span></td>
          <td>${formatTime(w.created_at)}</td>
        </tr>
      `;
    }).join('');
  } catch (err) {
    console.error('Error loading webhooks:', err);
  }
}

// Create Webhook
async function handleCreateWebhook(e) {
  e.preventDefault();
  const name = document.getElementById('wh-name').value.trim();
  const url = document.getElementById('wh-url').value.trim();
  const secret = document.getElementById('wh-secret').value.trim();

  if (!url) return;

  try {
    const res = await authFetch('/api/webhooks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name,
        url,
        secret,
        subscriptions: ['message_created', 'message_updated', 'conversation_created', 'conversation_updated']
      })
    });

    if (res.ok) {
      document.getElementById('new-webhook-form').reset();
      loadWebhooks();
    } else {
      const err = await res.json();
      alert('Erro ao cadastrar webhook: ' + (err.error || 'Desconhecido'));
    }
  } catch (err) {
    console.error('Error creating webhook:', err);
  }
}

// Load Meta Webhook Config
async function loadMetaConfig() {
  try {
    const res = await fetch('/api/config/meta-webhook');
    if (!res.ok) return;
    const data = await res.json();
    if (data.webhook_url) {
      document.getElementById('meta-callback-url').value = data.webhook_url;
    }
    if (data.verify_token) {
      document.getElementById('meta-verify-token').value = data.verify_token;
    }
    if (data.api_access_token) {
      apiAccessToken = data.api_access_token;
    }
  } catch (err) {
    console.error('Failed to load meta config:', err);
  }
}

// Copy Field Helper
function copyField(fieldId) {
  const input = document.getElementById(fieldId);
  if (!input) return;
  input.select();
  input.setSelectionRange(0, 99999);
  navigator.clipboard.writeText(input.value).then(() => {
    // Show subtle feedback
  }).catch(() => {
    document.execCommand('copy');
  });
}

// Helpers
function formatTime(isoString) {
  if (!isoString) return '';
  const d = new Date(isoString);
  if (isNaN(d.getTime())) return '';
  const now = new Date();
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  return d.toLocaleDateString([], { day: '2-digit', month: '2-digit' }) + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function getStatusIcon(status) {
  switch (status) {
    case 0: // Sent
      return `<span class="status-icon-sent" title="Enviado">${ICONS.check}</span>`;
    case 1: // Delivered
      return `<span class="status-icon-sent" title="Entregue">${ICONS.doubleCheck}</span>`;
    case 2: // Read
      return `<span class="status-icon-read" title="Lido">${ICONS.doubleCheck}</span>`;
    case 3: // Failed
      return `<span style="color:#EF4444;" title="Falha">${ICONS.alertCircle}</span>`;
    default:
      return '';
  }
}

function escapeHtml(str) {
  if (!str) return '';
  return str.replace(/[&<>'"]/g, 
    tag => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[tag] || tag));
}
