function Signal(key, initialValue) {
    let value = initialValue;
    let onChange = null;
    
    const signal = {
        Value: () => value,
        setValue: (newValue) => {
            value = newValue;
            if (onChange) onChange();
        },
        set onChange(callback) {
            onChange = callback;
        }
    };
    
    return signal;
}

let userId = null;
let userRole = null;
let userLevel = 1;
let isAdmin = false;

const chatSignal = Signal("chatOpenState", "close");
let checksum = null;
let announcements_Signal = null;

function cookie_get(key) {
    try {
        var cookies = {};
        for (var x in document.cookie.split("; ")) {
            var raw_data = document.cookie.split("; ")[x].split("=");
            cookies[raw_data[0]] = raw_data[1];
        }
        if (key in cookies) {
            return cookies[key];
        }
        return "";
    } catch {
        return "";
    }
}

function cookie_set(key, val) {
    try {
        document.cookie = `${key}=${val};expires=Thu, 01 Jan 2049 00:00:00 UTC`;
    } catch { }
}

var ignore = false;
var first = true;
var leads = false;

var messageMe = `
<div class="flex flex-row gap-2">
    <img src="/assets/logo-blue.png" class="w-7 h-7 rounded-full">
    <div
        class="bg-[var(--background)] border border-white/[0.35] rounded-lg w-full p-2 pt-1 text-xs flex flex-col gap-4">
        <p class="text-sm font-semibold text-[var(--primary)]">Exun Clan</p>
        <p class="text-white/80">
            {content}
        </p>
    </div>
</div>
`;

var messageYou = `
<div class="flex flex-row gap-2">
    <div
        class="bg-[var(--background)] border border-white/[0.35] rounded-lg w-full p-2 pt-1 text-xs flex flex-col gap-4">
        <p class="text-sm font-semibold text-white">You</p>
        <p class="text-white/80">
            {content}
        </p>
    </div>
    <img src="{avatar}" class="w-8 h-8 rounded-full">
</div>
`;

const messageTemplates = {
    loading: `
        <div class="chat-message loading" id="loadingMessage">
            <div class="chat-message-content">
                <div class="loading-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                </div>
            </div>
        </div>
    `,
    error: `
        <div class="chat-message error">
            <div class="chat-message-content">
                <p class="chat-message-text">Failed to load messages. <button class="retry-btn" onclick="retryLoadMessages()">Retry</button></p>
            </div>
        </div>
    `
};

document.addEventListener('DOMContentLoaded', function() {
    checksum = Signal("checksum", cookie_get("checksum"));
    announcements_Signal = Signal("announcements", cookie_get("announcements"));
    
    setupChecksumHandler();
    setupConnectionHandlers();
    setupVisibilityHandlers();
    setupChatSignalHandlers();
    setupMessageSubmission();
    
    initializeChatPopup();
    
    startChecksumPolling();
});

function setupChecksumHandler() {
    if (!checksum) return;
    
    checksum.onChange = async () => {
        var notyf = window.notyf || { success: function() {} };
        
        if (!ignore && !first) {
            if (chatSignal.Value() !== "open") {
                notyf.success({ 
                    position: { x: "center", y: "top" }, 
                    message: "You have got a new message!" 
                });
                showNotificationDot();
            }
        }
        if (ignore) {
            ignore = false;
        }
        
        cookie_set("checksum", checksum.Value());
        
        try {
            const response = await fetch("/api/leads");
            if (!response.ok) {
                return;
            }
            
            const data = await response.json();
            const chats = data || [];
            
            const hintsResponse = await fetch("/api/hints"); 
            if (!hintsResponse.ok) {
                return;
            }
            
            const hintsData = await hintsResponse.json();
            const hints = hintsData || [];
            
            updateChatContainers(chats, hints);
            updateBadgesInstantly(chats, hints);
        } catch (error) {
        }
    };
}
async function checkChecksum() {
    try {
        const response = await fetch("/api/chat/checksum", {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            },
            body: JSON.stringify({
                leadsHash: checksum ? checksum.Value() : '',
                hintsHash: announcements_Signal ? announcements_Signal.Value() : ''
            }),
            credentials: 'include'
        });
        const request = await response.json();
        
        leads = true;
        
        const leadsIndicator = document.getElementById("leads");
        const leadInput = document.getElementById("leadInput");
        const leadSendButton = document.getElementById("leadSendButton");
        
        if (leadsIndicator) leadsIndicator.style.backgroundColor = "#00da00";
        if (leadInput) leadInput.disabled = false;
        if (leadSendButton) leadSendButton.disabled = leadInput ? leadInput.value.trim().length === 0 : true;
        
        let hasNewLeads = false;
        let hasNewHints = false;
        
        if (checksum && checksum.Value() !== request["leadsHash"]) {
            hasNewLeads = !first && checksum.Value() !== '';
            checksum.setValue(request["leadsHash"]);
        }
        
        if (announcements_Signal && announcements_Signal.Value() !== request["hintsHash"]) {
            hasNewHints = !first && announcements_Signal.Value() !== '';
            announcements_Signal.setValue(request["hintsHash"]);
        }
        
        updateNotificationBadges(hasNewLeads, hasNewHints);
        
        if (first || request["changed"] || hasNewLeads || hasNewHints) {
            updateChatContainers(request["leads"] || [], request["hints"] || []);
        }
        
        if ((hasNewLeads || hasNewHints) && chatSignal.Value() !== "open") {
            var notyf = window.notyf || { success: function() {} };
            
            if (hasNewLeads) {
                notyf.success({ 
                    position: { x: "center", y: "top" }, 
                    message: "New message from leads!" 
                });
            }
            
            if (hasNewHints) {
                notyf.success({ 
                    position: { x: "center", y: "top" }, 
                    message: "New hint available!" 
                });
            }
            
            showNotificationDot();
        }
        
    } catch (error) {
    }
}

function setupMessageSubmission() {
    const leadSendButton = document.getElementById("leadSendButton");
    const leadInput = document.getElementById("leadInput");
    
    if (leadSendButton && leadInput) {
        leadSendButton.addEventListener("click", async (x) => {
            var text = leadInput.value.trim().trim("\n");
            if (text !== "") {
                leadInput.value = "";
                const leadMsgLen = document.getElementById("leadMsgLen");
                if (leadMsgLen) leadMsgLen.innerText = "0";
                
                var notyf = window.notyf || { error: function() {} };
                
                try {
                    // Immediately show the user's message
                    const leadsContainer = document.getElementById("leadsContainer");
                    if (leadsContainer) {
                        const userMessageHtml = `<div class="chat-message user">
                            <div class="chat-message-content">
                                <span class="chat-message-sender">You</span>
                                <div class="chat-message-text">${escapeHtml(text)}</div>
                            </div>
                        </div>`;
                        leadsContainer.innerHTML += userMessageHtml;
                        
                        setTimeout(() => {
                            leadsContainer.scrollTop = leadsContainer.scrollHeight;
                        }, 50);
                    }
                    
                    var response = await fetch("/submit_message", {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/json",
                            'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
                        },
                        body: JSON.stringify({ message: text })
                    });
                    var data = await response.json();
                    
                    if (!data.success) {
                        notyf.error({ 
                            position: { x: "center", y: "top" }, 
                            message: data.message || "Failed to send message"
                        });
                        // Remove the optimistically added message on error
                        if (leadsContainer) {
                            const messages = leadsContainer.querySelectorAll('.chat-message.user');
                            if (messages.length > 0) {
                                messages[messages.length - 1].remove();
                            }
                        }
                    } else {
                        // Check for immediate response without waiting for full sync
                        if (data.response) {
                            const adminMessageHtml = `<div class="chat-message admin">
                                <div class="chat-message-content">
                                    <span class="chat-message-sender">Admin</span>
                                    <div class="chat-message-text">${escapeHtml(data.response)}</div>
                                </div>
                            </div>`;
                            leadsContainer.innerHTML += adminMessageHtml;
                            
                            setTimeout(() => {
                                leadsContainer.scrollTop = leadsContainer.scrollHeight;
                            }, 50);
                        }
                        
                        // Still trigger the full sync but without waiting
                        ignore = true;
                        if (checksum && checksum.onChange) {
                            setTimeout(() => {
                                checksum.onChange();
                            }, 200); // Reduced from 1000ms to 200ms
                        }
                    }
                } catch (error) {
                    notyf.error({ 
                        position: { x: "center", y: "top" }, 
                        message: "Failed to send message" 
                    });
                    // Remove the optimistically added message on error
                    if (leadsContainer) {
                        const messages = leadsContainer.querySelectorAll('.chat-message.user');
                        if (messages.length > 0) {
                            messages[messages.length - 1].remove();
                        }
                    }
                }
            }
        });
        
        leadInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey && leadInput.value.trim()) {
                e.preventDefault();
                leadSendButton.click();
            }
        });
        
        leadInput.addEventListener('input', (e) => {
            if (e.target.value.trim().length >= 512) {
                leadInput.value = e.target.value.trim().slice(0, 512);
            }
            
            const leadMsgLen = document.getElementById("leadMsgLen");
            if (leadMsgLen) leadMsgLen.innerText = e.target.value.trim().length;
            
            if (leadSendButton) {
                leadSendButton.disabled = e.target.value.trim().length === 0;
            }
        });
    }
}

async function startChecksumPolling() {
    await checkChecksum();
    first = false;
    setInterval(checkChecksum, 500); // Reduced from 1000ms to 500ms for faster updates
}

function setupChatSignalHandlers() {
    const chatToggleBtn = document.getElementById("chatToggleBtn");
    const chatPopup = document.getElementById("chatPopup");
    const chatCloseBtn = document.getElementById("chatCloseBtn");
    const chatMinimizeBtn = document.getElementById("chatMinimizeBtn");
    
    if (!chatToggleBtn || !chatPopup) return;        chatSignal.onChange = () => {
        if (chatSignal.Value() === "open") {
            chatPopup.style.display = "flex";

            chatToggleBtn.style.opacity = 0;
            chatToggleBtn.style.transform = "scale(0)";

            setTimeout(() => {
                chatPopup.style.opacity = 1;
                chatPopup.style.transform = "translateY(0px)";
                chatToggleBtn.style.display = "none";
            }, 10);

            hideNotificationDot();
            
            const activeTab = document.querySelector('.chat-tab.active');
            if (activeTab) {
                const tabData = activeTab.getAttribute('data-tab');
                if (tabData) {
                    const badgeId = tabData + 'Badge';
                    const badge = document.getElementById(badgeId);
                    if (badge) {
                        badge.style.display = 'none';
                        badge.classList.remove('glowing');
                    }
                }
            }

            // Immediately refresh chat content when opening to ensure latest messages are shown
            refreshChatContent();

            const messagecontainer = document.getElementById("messagecontainer");
            setTimeout(() => {
                if (messagecontainer) {
                    messagecontainer.scrollTop = messagecontainer.scrollHeight;
                }
            }, 200);
        } else {
            chatToggleBtn.style.display = "block";

            chatPopup.style.opacity = 0;
            chatPopup.style.transform = "translateY(900px)";

            setTimeout(() => {
                chatPopup.style.display = "none";

                chatToggleBtn.style.opacity = 1;
                chatToggleBtn.style.transform = "scale(1)";
            }, 400);
        }
    };
    
    if (chatToggleBtn) {
        chatToggleBtn.addEventListener("click", (e) => {
            chatSignal.setValue("open");
            hideNotificationDot();
        });
    }
    
    if (chatCloseBtn) {
        chatCloseBtn.addEventListener("click", (e) => {
            chatSignal.setValue("close");
        });
    }
    
    if (chatMinimizeBtn) {
        chatMinimizeBtn.addEventListener("click", (e) => {
            chatSignal.setValue("close");
        });
    }
}

function setupConnectionHandlers() {
    window.addEventListener('online', function() {
        isOnline = true;
        updateConnectionStatus(true);
        checkChecksum();
    });
    
    window.addEventListener('offline', function() {
        isOnline = false;
        updateConnectionStatus(false);
    });
}

function setupVisibilityHandlers() {
    document.addEventListener('visibilitychange', function() {
        if (document.hidden) {
        } else {
            lastActivity = Date.now();
            checkChecksum();
            
            // Refresh chat content if chat is open when page becomes visible
            if (chatSignal.Value() === "open") {
                refreshChatContent();
            }
        }
    });
    
    ['click', 'keypress', 'scroll', 'mousemove'].forEach(event => {
        document.addEventListener(event, function() {
            lastActivity = Date.now();
        }, { passive: true });
    });
}

async function initializeChatPopup() {
    try {
        showLoadingState();
        
        const sessionResponse = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        
        if (!sessionResponse.ok) {
            throw new Error(`Session failed: ${sessionResponse.status}`);
        }
        
        const sessionData = await sessionResponse.json();
        userId = sessionData.userId;
        userRole = sessionData.role;
        userLevel = sessionData.level || 1;
        isAdmin = sessionData.role === 'admin';
        
        window.user = {
            id: sessionData.userId,
            name: sessionData.name,
            email: sessionData.email || sessionData.gmail,
            level: sessionData.level || 1
        };
        
        const levelDisplay = document.getElementById('currentLevel');
        if (levelDisplay) {
            levelDisplay.textContent = userLevel;
        }
        
        updateConnectionStatus(true);
        hideLoadingState();
        
    } catch (error) {
        updateConnectionStatus(false);
        showErrorState('Failed to initialize chat. Please refresh the page.');
        hideLoadingState();
    }
}

function showLoadingState() {
    const containers = ['leadsContainer', 'hintsContainer'];
    containers.forEach(containerId => {
        const container = document.getElementById(containerId);
        if (container) {
            container.innerHTML = messageTemplates.loading;
        }
    });
}

function hideLoadingState() {
    const loadingMessages = document.querySelectorAll('#loadingMessage');
    loadingMessages.forEach(msg => msg.remove());
}

function showErrorState(message) {
    const containers = ['leadsContainer', 'hintsContainer'];
    containers.forEach(containerId => {
        const container = document.getElementById(containerId);
        if (container && container.innerHTML.includes('loading-dots')) {
            container.innerHTML = messageTemplates.error.replace('{error}', message);
        }
    });
}

function retryLoadMessages() {
    connectionRetries = 0;
    initializeChatPopup();
}

function updateConnectionStatus(isConnected) {
    const statusDot = document.querySelector('.status-dot');
    const statusIndicator = document.querySelector('.status-indicator');
    const chatStatus = document.getElementById('chatStatus');
    
    if (statusDot) {
        statusDot.classList.toggle('disconnected', !isConnected);
    }
    
    if (statusIndicator) {
        statusIndicator.classList.toggle('offline', !isConnected);
    }
    
    if (chatStatus) {
        const statusEl = chatStatus.querySelector('.status-indicator');
        if (statusEl) {
            statusEl.classList.toggle('offline', !isConnected);
        }
    }
}

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return '';
}

function updateNotificationBadges(hasNewLeads, hasNewHints) {
    const leadsBadge = document.getElementById('leadsBadge');
    const hintsBadge = document.getElementById('hintsBadge');
    const chatToggleBtn = document.getElementById("chatToggleBtn");
    
    if (leadsBadge) {
        leadsBadge.style.display = 'none';
    }
    
    if (hintsBadge) {
        hintsBadge.style.display = 'none';
    }
    
    if (chatToggleBtn) {
        if (chatSignal.Value() !== "open" && (hasNewLeads || hasNewHints)) {
            showNotificationDot();
        }
    }
}

function switchChatTab(tabName) {
    const tabContents = document.querySelectorAll('.chat-tab-content');
    tabContents.forEach(content => {
        content.classList.remove('active');
        content.style.display = 'none';
    });
    
    const tabs = document.querySelectorAll('.chat-tab');
    tabs.forEach(tab => {
        tab.classList.remove('active');
    });
    
    const selectedContent = document.getElementById(tabName + 'Content');
    const selectedTab = document.querySelector(`[data-tab="${tabName}"]`);
    
    if (selectedContent) {
        selectedContent.classList.add('active');
        selectedContent.style.display = 'flex';
        
        const container = selectedContent.querySelector('.chat-messages-container');
        if (container) {
            setTimeout(() => {
                container.scrollTop = container.scrollHeight;
            }, 100);
        }
    }
    
    if (selectedTab) {
        selectedTab.classList.add('active');
        
        const badgeId = tabName + 'Badge';
        const badge = document.getElementById(badgeId);
        if (badge) {
            badge.style.display = 'none';
            badge.classList.remove('glowing');
        }
    }
    
    // Refresh content when switching tabs to ensure latest data is shown
    if (chatSignal.Value() === "open") {
        refreshChatContent();
    }
    
    setTimeout(() => {
        const container = selectedContent?.querySelector('.chat-messages-container');
        if (container) {
            container.scrollTop = container.scrollHeight;
        }
    }, 100);
}

function updateChatContainers(chats, hints) {
    renderMessages(chats, 'leadsContainer');
    if (!window.location.pathname.endsWith('/hints.html')) {
        renderMessages(hints, 'hintsContainer');
    }
}

function formatTime(timestamp) {
    if (!timestamp) return '';
    try {
        const date = new Date(timestamp);
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch (error) {
        return '';
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function updateBadgesInstantly(chats, hints) {
    const leadsBadge = document.getElementById('leadsBadge');
    const hintsBadge = document.getElementById('hintsBadge');
    
    if (leadsBadge) {
        leadsBadge.style.display = 'none';
    }
    
    if (hintsBadge) {
        hintsBadge.style.display = 'none';
    }
    
    if (chatSignal.Value() !== "open" && (chats.length > 0 || (hints && hints.length > 0))) {
        showNotificationDot();
    } else if (chatSignal.Value() === "open") {
        hideNotificationDot();
    }
}

let chatChecksum = '';

function showNotificationDot() {
    const chatToggleBtn = document.getElementById("chatToggleBtn");
    if (!chatToggleBtn) return;
    
    let dot = chatToggleBtn.querySelector('.notification-dot');
    if (!dot) {
        dot = document.createElement('span');
        dot.className = 'notification-dot show';
        chatToggleBtn.appendChild(dot);
    }
}

function hideNotificationDot() {
    const chatToggleBtn = document.getElementById("chatToggleBtn");
    if (!chatToggleBtn) return;
    
    const dot = chatToggleBtn.querySelector('.notification-dot');
    if (dot) {
        dot.remove();
    }
}

function renderMessages(messages, containerId) {
    const container = document.getElementById(containerId);
    if (!container) return;
    
    if (messages.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">${containerId === 'hintsContainer' ? '?' : 'INBOX'}</div>
                <p>No ${containerId === 'hintsContainer' ? 'hints' : 'messages'} yet</p>
            </div>
        `;
        return;
    }
    
    // Check if we can optimize by only appending new messages
    const existingMessages = container.querySelectorAll('.chat-message, .hint-message');
    const shouldOptimize = existingMessages.length > 0 && messages.length > existingMessages.length;
    
    if (shouldOptimize && containerId === 'leadsContainer') {
        // Only append new messages instead of re-rendering all
        const newMessages = messages.slice(existingMessages.length);
        const newHtml = newMessages.map(x => {
            const isOwnMessage = x.userEmail && window.user && (
                x.userEmail.toLowerCase() === window.user.email.toLowerCase() || 
                x.userEmail.toLowerCase() === window.user.gmail?.toLowerCase()
            );
            
            const isAdminMessage = x.isReply || x.sentBy || (x.userEmail && x.userEmail.includes('admin'));
            
            if (isAdminMessage) {
                return `<div class="chat-message admin">
                    <div class="chat-message-content">
                        <span class="chat-message-sender">Admin</span>
                        <div class="chat-message-text">${escapeHtml(x.message)}</div>
                    </div>
                </div>`;
            } else {
                return `<div class="chat-message user">
                    <div class="chat-message-content">
                        <span class="chat-message-sender">You</span>
                        <div class="chat-message-text">${escapeHtml(x.message)}</div>
                    </div>
                </div>`;
            }
        }).join('');
        
        container.innerHTML += newHtml;
    } else {
        // Full re-render for hints or when optimization isn't possible
        if (containerId === 'hintsContainer') {
            container.innerHTML = messages.map(hint => `
                <div class="hint-message">
                    <div class="message-content">
                        ${hint.message ? (typeof showdown !== 'undefined' ? 
                            new showdown.Converter().makeHtml(hint.message) : 
                            escapeHtml(hint.message)) : ''}
                    </div>
                </div>
            `).join('');
        } else {
            container.innerHTML = messages.map(x => {
                const isOwnMessage = x.userEmail && window.user && (
                    x.userEmail.toLowerCase() === window.user.email.toLowerCase() || 
                    x.userEmail.toLowerCase() === window.user.gmail?.toLowerCase()
                );
                
                const isAdminMessage = x.isReply || x.sentBy || (x.userEmail && x.userEmail.includes('admin'));
                
                if (isAdminMessage) {
                    return `<div class="chat-message admin">
                        <div class="chat-message-content">
                            <span class="chat-message-sender">Admin</span>
                            <div class="chat-message-text">${escapeHtml(x.message)}</div>
                        </div>
                    </div>`;
                } else {
                    return `<div class="chat-message user">
                        <div class="chat-message-content">
                            <span class="chat-message-sender">You</span>
                            <div class="chat-message-text">${escapeHtml(x.message)}</div>
                        </div>
                    </div>`;
                }
            }).join('');
        }
    }
    
    setTimeout(() => {
        container.scrollTop = container.scrollHeight;
    }, 50); // Reduced from 100ms to 50ms for faster scrolling
}

function updateChatMessages(messages) {
    if (messages.leads) {
        renderMessages(messages.leads, 'leadsContainer');
    }
    if (messages.hints) {
        renderMessages(messages.hints, 'hintsContainer');
    }
}

async function refreshChatContent() {
    try {
        // Fetch latest leads/chat messages
        const response = await fetch("/api/leads", {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        
        if (response.ok) {
            const chats = await response.json();
            renderMessages(chats || [], 'leadsContainer');
        }

        // Fetch latest hints (if not on hints page)
        if (!window.location.pathname.endsWith('/hints.html')) {
            const hintsResponse = await fetch("/api/hints", {
                headers: {
                    'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
                }
            });
            
            if (hintsResponse.ok) {
                const hints = await hintsResponse.json();
                renderMessages(hints || [], 'hintsContainer');
            }
        }

        // Scroll to bottom after loading content
        setTimeout(() => {
            const activeTab = document.querySelector('.chat-tab.active');
            if (activeTab) {
                const tabData = activeTab.getAttribute('data-tab');
                const container = document.getElementById(tabData + 'Container');
                if (container) {
                    container.scrollTop = container.scrollHeight;
                }
            }
        }, 100);
        
    } catch (error) {
        console.log('Failed to refresh chat content:', error);
    }
}
