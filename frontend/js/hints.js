// Hints page JavaScript
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

async function loadQuestions() {
    try {
        console.log('Loading questions from /api/question...');
        
        const response = await fetch('/api/question', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        
        console.log('Questions API response status:', response.status);
        
        const questionsList = document.getElementById('questionsList');
        
        if (!response.ok) {
            console.log('API response not ok, status:', response.status);
            if (response.status === 401) {
                console.log('User not authenticated, redirecting to auth page');
                window.location.href = '/auth';
                return;
            }
            questionsList.innerHTML = '<div class="no-questions"><p>No questions available yet</p></div>';
            return;
        }
        
        const data = await response.json();
        console.log('Questions API response data:', data);
        
        if (data && data.question && data.question.markdown && data.question.markdown.trim()) {
            console.log('Displaying level questions');
            
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
                        backslashEscapesHTMLTags: true
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
                    console.warn('Showdown library not loaded, displaying raw text');
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
            console.log('No question data in response');
            questionsList.innerHTML = '<div class="no-questions"><p>No questions available yet</p></div>';
        }
    } catch (error) {
        console.error('Error loading questions:', error);
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
        const notificationDot = document.getElementById('notificationDot');
        if (data.count > 0) {
            notificationDot.classList.add('show');
        } else {
            notificationDot.classList.remove('show');
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

document.addEventListener('DOMContentLoaded', function() {
    checkAdminAccess();
    loadQuestions();
    checkNotifications();
    setInterval(checkNotifications, 30000);
});
