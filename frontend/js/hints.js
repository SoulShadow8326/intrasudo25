function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

async function loadHints() {
    try {
        console.log('Loading hints from /api/question...');
        
        const response = await fetch('/api/question', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        
        console.log('Hints API response status:', response.status);
        
        const questionContainer = document.getElementById('levelQuestionContainer');
        const levelQuestionDiv = document.getElementById('levelQuestion');
        const hintsContainer = document.getElementById('hintsContainer');
        
        if (!response.ok) {
            console.log('API response not ok, status:', response.status);
            questionContainer.style.display = 'none';
            hintsContainer.innerHTML = '<div class="hint-item" style="text-align: center; color: rgba(255, 255, 255, 0.7);">No hints available yet</div>';
            return;
        }
        
        const data = await response.json();
        console.log('Hints API response data:', data);
        
        if (data && data.question) {
            const q = data.question;
            console.log('Level question data:', q);
            
            // Display the level question with markdown rendering
            if (q.description && q.description.trim()) {
                console.log('Displaying level question with markdown');
                questionContainer.style.display = 'block';
                
                // Initialize Showdown converter
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
                    levelQuestionDiv.innerHTML = converter.makeHtml(q.description);
                } else {
                    console.warn('Showdown library not loaded, displaying raw markdown');
                    levelQuestionDiv.innerHTML = `<pre>${q.description}</pre>`;
                }
            } else {
                questionContainer.style.display = 'none';
            }
            
            // Display hints if available
            if (q.sourceHint || q.consoleHint) {
                console.log('Displaying hints');
                
                const hintText = q.sourceHint || q.consoleHint;
                
                hintsContainer.innerHTML = `
                    <div class="hint-item" style="
                        text-align: center;
                        padding: 3rem;
                        margin: 2rem auto;
                        max-width: 600px;
                        background: rgba(255, 255, 255, 0.05);
                        border: 2px solid rgba(255, 255, 255, 0.1);
                        border-radius: 1rem;
                        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
                    ">
                        <p class="hint-text" style="
                            font-size: 3rem;
                            font-weight: bold;
                            color: var(--primary);
                            margin: 0;
                            letter-spacing: 0.1em;
                            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
                        ">${hintText}</p>
                    </div>
                `;
            } else {
                console.log('No specific hints available');
                hintsContainer.innerHTML = `
                    <div class="hint-item">
                        <h3 class="hint-title">General Hint</h3>
                        <p class="hint-text">Study the question carefully. Look for patterns, hidden meanings, or references that might lead you to the answer.</p>
                    </div>
                `;
            }
        } else {
            console.log('No question data in response');
            questionContainer.style.display = 'none';
            hintsContainer.innerHTML = `
                <div class="hint-item">
                    <h3 class="hint-title">General Hint</h3>
                    <p class="hint-text">Study the question carefully. Look for patterns, hidden meanings, or references that might lead you to the answer.</p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Error loading hints:', error);
        document.getElementById('levelQuestionContainer').style.display = 'none';
        document.getElementById('hintsContainer').innerHTML = '<div class="hint-item" style="text-align: center; color: #dc3545;">Failed to load hints</div>';
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
            if (userData.isAdmin) {
                document.getElementById('adminLink').style.display = 'inline-block';
            } else {
                document.getElementById('adminLink').style.display = 'none';
            }
        } else {
            document.getElementById('adminLink').style.display = 'none';
        }
    } catch (error) {
        document.getElementById('adminLink').style.display = 'none';
    }
}

document.addEventListener('DOMContentLoaded', function() {
    checkAdminAccess();
    loadHints();
    checkNotifications();
    setInterval(checkNotifications, 30000);
});
