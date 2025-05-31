let chatSocket = null;
let userId = null;
let userRole = null;
let isAdmin = false;
let chatOpen = false;
let hasNewMessages = false;

const messageTemplates = {
    admin: `
        <div class="chat-message admin">
            <img src="/assets/logo-blue.png" class="chat-message-avatar">
            <div class="chat-message-content">
                <p class="chat-message-sender">Exun Clan</p>
                <p class="chat-message-text">{content}</p>
            </div>
        </div>
    `,
    user: `
        <div class="chat-message user">
            <div class="chat-message-content">
                <p class="chat-message-sender">You</p>
                <p class="chat-message-text">{content}</p>
            </div>
            <img src="{avatar}" class="chat-message-avatar">
        </div>
    `
};

document.addEventListener('DOMContentLoaded', function() {
    initializeChatPopup();
});

async function initializeChatPopup() {
    try {
        const sessionResponse = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        if (!sessionResponse.ok) return;
        
        const sessionData = await sessionResponse.json();
        userId = sessionData.userId;
        userRole = sessionData.role;
        
        setupChatEventListeners();
        checkForNewMessages();
        
    } catch (error) {
        console.error('Failed to initialize chat:', error);
    }
}

function setupChatEventListeners() {
    const toggleBtn = document.getElementById('btn-message');
    const chatButton = document.querySelector('.chat-button');
    const chatPopup = document.getElementById('chatPopup');
    const chat_popup = document.getElementById('chat-popup');
    const closeBtn = document.getElementById('chatCloseBtn');
    const minimizeBtn = document.getElementById('chatMinimizeBtn');
    const maximizeBtn = document.getElementById('chatMaximizeBtn');
    const sendBtn = document.getElementById('chatSendButton');
    const chat_send = document.getElementById('chat-send');
    const chatInput = document.getElementById('chatInput');
    const chat_input = document.getElementById('chat-input');
    
    if (toggleBtn) {
        toggleBtn.addEventListener('click', toggleChat);
    }
    
    if (chatButton) {
        chatButton.addEventListener('click', toggleChat);
    }
    
    if (closeBtn) {
        closeBtn.addEventListener('click', closeChat);
    }
    
    if (minimizeBtn) {
        minimizeBtn.addEventListener('click', minimizeChat);
    }
    
    if (maximizeBtn) {
        maximizeBtn.addEventListener('click', maximizeChat);
    }
    
    if (sendBtn) {
        sendBtn.addEventListener('click', handleChatSubmit);
    }
    
    if (chat_send) {
        chat_send.addEventListener('click', handleChatSubmit);
    }
    
    if (chatInput) {
        chatInput.addEventListener('input', updateCharCount);
        chatInput.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleChatSubmit();
            }
        });
    }
    
    if (chat_input) {
        chat_input.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleChatSubmit();
            }
        });
    }
}

async function toggleChat() {
    const chatPopup = document.getElementById('chatPopup');
    const toggleBtnContainer = document.getElementById('chatToggleBtn');
    
    if (!chatOpen) {
        chatOpen = true;
        chatPopup.style.display = 'flex';
        
        toggleBtnContainer.style.opacity = '0';
        toggleBtnContainer.style.transform = 'scale(0)';
        
        setTimeout(() => {
            chatPopup.style.opacity = '1';
            chatPopup.style.transform = 'translateY(0px)';
            toggleBtnContainer.style.display = 'none';
        }, 10);
        
        await loadChatHistory();
        connectWebSocket();
        updateNotificationState(false);
        
        const messageContainer = document.getElementById('messagecontainer');
        setTimeout(() => {
            messageContainer.scrollTop = messageContainer.scrollHeight;
        }, 200);
    }
}

function closeChat() {
    const chatPopup = document.getElementById('chatPopup');
    const toggleBtnContainer = document.getElementById('chatToggleBtn');
    
    chatOpen = false;
    
    toggleBtnContainer.style.display = 'block';
    
    chatPopup.style.opacity = '0';
    chatPopup.style.transform = 'translateY(900px)';
    
    setTimeout(() => {
        chatPopup.style.display = 'none';
        
        toggleBtnContainer.style.opacity = '1';
        toggleBtnContainer.style.transform = 'scale(1)';
    }, 400);
    
}

function minimizeChat() {
    const chatPopup = document.getElementById('chatPopup');
    const toggleBtnContainer = document.getElementById('chatToggleBtn');
    
    chatOpen = false;
    
    // Minimize animation - scale down and fade out
    chatPopup.style.transform = 'scale(0.1) translateY(50px)';
    chatPopup.style.opacity = '0';
    
    setTimeout(() => {
        chatPopup.style.display = 'none';
        toggleBtnContainer.style.display = 'block';
        toggleBtnContainer.style.opacity = '1';
        toggleBtnContainer.style.transform = 'scale(1)';
    }, 300);
}

function maximizeChat() {
    const chatPopup = document.getElementById('chatPopup');
    
    // Toggle maximized state
    if (chatPopup.style.width === '90vw') {
        // Restore to normal size
        chatPopup.style.width = '350px';
        chatPopup.style.height = 'calc(100vh - 175px)';
        chatPopup.style.maxWidth = '350px';
        chatPopup.style.maxHeight = 'calc(100vh - 175px)';
    } else {
        // Maximize
        chatPopup.style.width = '90vw';
        chatPopup.style.height = '90vh';
        chatPopup.style.maxWidth = '90vw';
        chatPopup.style.maxHeight = '90vh';
    }
}

async function loadChatHistory() {
    try {
        const response = await fetch('/api/chat/messages', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        if (!response.ok) throw new Error('Failed to load chat history');
        
        const messages = await response.json();
        displayMessages(messages);
        
    } catch (error) {
        console.error('Failed to load chat history:', error);
    }
}

function displayMessages(messages) {
    const container = document.getElementById('messagecontainer');
    if (!container) return;
    
    container.innerHTML = '';
    
    messages.forEach(message => {
        const messageElement = createMessageElement(message);
        container.appendChild(messageElement);
    });
    
    container.scrollTop = container.scrollHeight;
}

function createMessageElement(message) {
    const div = document.createElement('div');
    
    let template;
    if (message.isAdmin) {
        template = messageTemplates.admin;
    } else {
        template = messageTemplates.user;
    }
    
    div.innerHTML = template
        .replace('{content}', escapeHtml(message.content))
        .replace('{avatar}', '/assets/logo_nobg.png');
    
    return div.firstElementChild;
}

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/chat/ws`;
    
    chatSocket = new WebSocket(wsUrl);
    
    chatSocket.onopen = function() {
        console.log('Chat WebSocket connected');
    };
    
    chatSocket.onmessage = function(event) {
        const message = JSON.parse(event.data);
        if (chatOpen) {
            addMessageToChat(message);
        } else {
            updateNotificationState(true);
        }
    };
    
    chatSocket.onclose = function() {
        console.log('Chat WebSocket disconnected');
        if (chatOpen) {
            setTimeout(connectWebSocket, 3000);
        }
    };
    
    chatSocket.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
}

function addMessageToChat(message) {
    const container = document.getElementById('messagecontainer');
    if (!container) return;
    
    const messageElement = createMessageElement(message);
    container.appendChild(messageElement);
    container.scrollTop = container.scrollHeight;
}

async function handleChatSubmit() {
    const input = document.getElementById('chatInput');
    if (!input) return;
    
    const content = input.value.trim();
    if (!content) return;
    
    if (content.length > 512) {
        alert('Message too long! Maximum 512 characters.');
        return;
    }
    
    try {
        const response = await fetch('/api/chat/send', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            },
            body: JSON.stringify({ content })
        });
        
        if (!response.ok) {
            throw new Error('Failed to send message');
        }
        
        input.value = '';
        updateCharCount();
        
    } catch (error) {
        console.error('Failed to send message:', error);
        alert('Failed to send message. Please try again.');
    }
}

function updateCharCount() {
    const input = document.getElementById('chatInput');
    const counter = document.getElementById('chatMsgLen');
    
    if (input && counter) {
        counter.textContent = input.value.length;
    }
}

async function checkForNewMessages() {
    try {
        const response = await fetch('/api/chat/unread-count', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            updateNotificationState(data.count > 0);
            if (data.count > 0) {
                updateMessageCount(data.count);
            }
        }
    } catch (error) {
        console.error('Failed to check for new messages:', error);
    }
}

function updateNotificationState(hasNotifications) {
    hasNewMessages = hasNotifications;
    
    const logoNotification = document.getElementById('logoNotification');
    const messageCountBadge = document.getElementById('message-count');
    
    if (logoNotification) {
        logoNotification.style.display = hasNotifications ? 'block' : 'none';
    }
    
    if (messageCountBadge) {
        messageCountBadge.style.display = hasNotifications ? 'inline-flex' : 'none';
    }
}

function updateMessageCount(count) {
    const messageCountBadge = document.getElementById('message-count');
    if (messageCountBadge && count > 0) {
        messageCountBadge.textContent = count > 99 ? '99+' : count.toString();
        messageCountBadge.style.display = 'inline-flex';
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
