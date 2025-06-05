function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

async function loadQuestions() {
    try {
        const response = await fetch('/api/question', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        
        const questionsList = document.getElementById('questionsList');
        
        if (!response.ok) {
            if (response.status === 401) {
                window.location.href = '/auth';
                return;
            }
            questionsList.innerHTML = '<div class="no-questions"><p>No questions available yet</p></div>';
            return;
        }
        
        const data = await response.json();
        
        if (data && data.question && data.question.markdown && data.question.markdown.trim()) {
            const questions = data.question.markdown.split('\n\n').filter(question => question.trim());
            
            if (questions.length > 0) {
                if (typeof showdown !== 'undefined') {
                    const converter = new showdown.Converter({
                        tables: true,
                        strikethrough: true,
                        ghCodeBlocks: true,
                        tasklists: true,
                        simpleLineBreaks: true,
                        openLinksInNewWindow: true,
                        backslashEscapesHTMLTags: true,
                        emoji: true,
                        underline: true,
                        completeHTMLDocument: false,
                        metadata: false,
                        splitAdjacentBlockquotes: true,
                        smartIndentationFix: true,
                        disableForced4SpacesIndentedSublists: true,
                        literalMidWordUnderscores: true
                    });
                    
                    questionsList.innerHTML = questions.map((question, index) => `
                        <div class="question-item">
                            <div class="question-header">
                                <div class="question-number">${index + 1}</div>
                                <div class="question-content">
                                    ${converter.makeHtml(question.trim())}
                                </div>
                            </div>
                        </div>
                    `).join('');
                } else {
                    questionsList.innerHTML = questions.map((question, index) => `
                        <div class="question-item">
                            <div class="question-header">
                                <div class="question-number">${index + 1}</div>
                                <div class="question-content">
                                    <pre>${question.trim()}</pre>
                                </div>
                            </div>
                        </div>
                    `).join('');
                }
            } else {
                questionsList.innerHTML = '<div class="no-questions"><p>No questions available yet</p></div>';
            }
        } else {
            questionsList.innerHTML = '<div class="no-questions"><p>No questions available yet</p></div>';
        }
    } catch (error) {
        document.getElementById('questionsList').innerHTML = '<div class="no-questions"><p>Failed to load questions</p></div>';
    }
}

async function checkNotifications() {
    try {
        const response = await fetch('/api/notifications/unread-count', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        const data = await response.json();
        const logoNotification = document.getElementById('logoNotification');
        if (logoNotification) {
            if (data.count > 0) {
                logoNotification.style.display = 'block';
            } else {
                logoNotification.style.display = 'none';
            }
        }
    } catch (error) {}
}

async function checkAdminAccess() {
    try {
        const response = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        if (response.ok) {
            const userData = await response.json();
            const adminLinks = ['adminLink', 'mobileAdminLink'];
            adminLinks.forEach(linkId => {
                const element = document.getElementById(linkId);
                if (element) {
                    element.style.display = userData.isAdmin ? 'inline-block' : 'none';
                }
            });
        } else {
            const adminLinks = ['adminLink', 'mobileAdminLink'];
            adminLinks.forEach(linkId => {
                const element = document.getElementById(linkId);
                if (element) {
                    element.style.display = 'none';
                }
            });
        }
    } catch (error) {
        const adminLinks = ['adminLink', 'mobileAdminLink'];
        adminLinks.forEach(linkId => {
            const element = document.getElementById(linkId);
            if (element) {
                element.style.display = 'none';
            }
        });
    }
}

// Announcements functions
async function loadAnnouncements() {
    try {
        const response = await fetch('/api/announcements');
        
        if (!response.ok) {
            return;
        }
        
        const announcements = await response.json();
        
        const announcementsContainer = document.getElementById('announcementsContainer');
        
        if (announcements && announcements.length > 0) {
            displayAnnouncements(announcements);
            announcementsContainer.style.display = 'flex';
        } else {
            announcementsContainer.style.display = 'none';
        }
    } catch (error) {
        document.getElementById('announcementsContainer').style.display = 'none';
    }
}

function displayAnnouncements(announcements) {
    const announcementsList = document.getElementById('announcementsList');
    
    const announcementsHTML = announcements.map(announcement => `
        <div class="announcement-banner">
            <h2 class="announcement-heading">
                ${escapeHtml(announcement.heading)}
            </h2>
        </div>
    `).join('');
    
    announcementsList.innerHTML = announcementsHTML;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

async function loadHints() {
    try {
        const response = await fetch('/api/hints', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        
        const hintsContainer = document.getElementById('hintsContainer');
        
        if (!response.ok) {
            if (response.status === 401) {
                window.location.href = '/auth';
                return;
            }
            hintsContainer.innerHTML = '<div class="empty-state"><div class="empty-icon">?</div><p>No hints available yet.</p></div>';
            return;
        }
        
        const hints = await response.json();
        
        if (hints && Array.isArray(hints) && hints.length > 0) {
            displayHints(hints);
        } else {
            hintsContainer.innerHTML = '<div class="empty-state"><div class="empty-icon">?</div><p>No hints available yet.</p></div>';
        }
    } catch (error) {
        document.getElementById('hintsContainer').innerHTML = '<div class="empty-state"><div class="empty-icon">?</div><p>Failed to load hints.</p></div>';
    }
}

function displayHints(hints) {
    const hintsContainer = document.getElementById('hintsContainer');
    
    const hintsHTML = hints.map(hint => `
        <div class="chat-message hint-message">
            <div class="message-header">
                <span class="message-author">Admin</span>
                <span class="message-time">${formatTime(hint.timestamp)}</span>
            </div>
            <div class="message-content">
                ${hint.message ? (typeof showdown !== 'undefined' ? 
                    new showdown.Converter().makeHtml(hint.message) : 
                    escapeHtml(hint.message)) : ''}
            </div>
        </div>
    `).join('');
    
    hintsContainer.innerHTML = hintsHTML;
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

document.addEventListener('DOMContentLoaded', function() {
    checkAdminAccess();
    loadQuestions();
    checkNotifications();
    loadAnnouncements();
    loadHints();
    setInterval(checkNotifications, 30000);
    setInterval(loadHints, 30000);
});
