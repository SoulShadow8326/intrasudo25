function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

let chatSocket = null;
let userId = null;
let userRole = null;
let isAdmin = false;

async function initializeChat() {
    try {
        const secret = await getSecret('GET');
        const sessionResponse = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        if (!sessionResponse.ok) {
            window.location.href = 'auth.html';
            return;
        }
        
        const sessionData = await sessionResponse.json();
        userId = sessionData.userId;
        userRole = sessionData.role;
        
        await loadChatHistory();
        connectWebSocket();
        enableChatInput();
        
    } catch (error) {
        console.error('Failed to initialize chat:', error);
        showError('Failed to initialize chat');
    }
}

async function loadChatHistory() {
    try {
        const secret = await getSecret('GET');
        const response = await fetch('/api/chat/messages', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        if (!response.ok) throw new Error('Failed to load chat history');
        
        const messages = await response.json();
        displayMessages(messages);
        
        document.getElementById('chatLoadingState').style.display = 'none';
        
    } catch (error) {
        console.error('Failed to load chat history:', error);
        document.getElementById('chatLoadingState').textContent = 'Failed to load messages';
    }
}

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/chat`;
    
    chatSocket = new WebSocket(wsUrl);
    
    chatSocket.onopen = function() {
        console.log('Chat WebSocket connected');
    };
    
    chatSocket.onmessage = function(event) {
        const message = JSON.parse(event.data);
        displayMessage(message);
    };
    
    chatSocket.onclose = function() {
        console.log('Chat WebSocket disconnected');
        setTimeout(connectWebSocket, 3000);
    };
    
    chatSocket.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
}

function displayMessages(messages) {
    const container = document.getElementById('chatMessages');
    container.innerHTML = '';
    
    messages.forEach(message => {
        displayMessage(message);
    });
    
    scrollToBottom();
}

function displayMessage(message) {
    const container = document.getElementById('chatMessages');
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${message.isAdmin ? 'admin-message' : 'user-message'}`;
    
    const timeString = new Date(message.timestamp).toLocaleTimeString();
    
    messageDiv.innerHTML = `
        <div class="message-header">
            <span class="message-author">${message.authorName}</span>
            <span class="message-time">${timeString}</span>
        </div>
        <div class="message-text">${escapeHtml(message.content)}</div>
    `;
    
    container.appendChild(messageDiv);
    scrollToBottom();
}

function enableChatInput() {
    const input = document.getElementById('chatInput');
    input.disabled = false;
    input.focus();
}

async function handleChatSubmit(event) {
    event.preventDefault();
    const input = document.getElementById('chatInput');
    const message = input.value.trim();
    
    if (!message) return;
    
    try {
        const secret = await getSecret('POST');
        if (!secret) {
            throw new Error('Authentication service unavailable');
        }
        
        const response = await fetch('/api/chat/send', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            },
            body: JSON.stringify({
                content: message,
                userId: userId
            })
        });
        
        if (!response.ok) throw new Error('Failed to send message');
        
        input.value = '';
        
    } catch (error) {
        console.error('Failed to send message:', error);
        showError('Failed to send message');
    }
}

function scrollToBottom() {
    const container = document.getElementById('chatMessages');
    container.scrollTop = container.scrollHeight;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showError(message) {
    const container = document.getElementById('chatMessages');
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-message';
    errorDiv.textContent = message;
    container.appendChild(errorDiv);
    scrollToBottom();
}

async function checkAdminStatus() {
    try {
        const secret = await getSecret('GET');
        const response = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        if (response.ok) {
            const userData = await response.json();
            isAdmin = !!userData.isAdmin;
            if (isAdmin && document.getElementById('adminLink')) {
                document.getElementById('adminLink').style.display = 'inline-block';
            }
        } else {
            isAdmin = false;
            if (document.getElementById('adminLink')) {
                document.getElementById('adminLink').style.display = 'none';
            }
        }
    } catch (error) {
        isAdmin = false;
        if (document.getElementById('adminLink')) {
            document.getElementById('adminLink').style.display = 'none';
        }
    }
}

document.addEventListener('DOMContentLoaded', async function() {
    await checkAdminStatus();
    initializeChat();
});
