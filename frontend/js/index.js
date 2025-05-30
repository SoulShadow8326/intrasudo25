let currentLevel = null;
let userSession = null;

async function loadUserSession() {
    try {
        const secret = await getSecret('GET');
        const response = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        userSession = await response.json();
        checkAdminAccess();
        loadCurrentLevel();
    } catch (error) {
        console.error('Failed to load user session:', error);
        checkAdminAccess();
        loadCurrentLevel();
    }
}

async function checkAdminAccess() {
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

async function loadCurrentLevel() {
    try {
        const secret = await getSecret('GET');
        const response = await fetch('/api/user/current-level', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        currentLevel = await response.json();
        
        document.getElementById('levelTitle').textContent = `Level ${currentLevel.number}`;
        document.getElementById('levelDescription').textContent = currentLevel.description || '';
        
        if (currentLevel.mediaUrl) {
            const mediaContainer = document.getElementById('levelMedia');
            if (currentLevel.mediaType === 'image') {
                mediaContainer.innerHTML = `<img src="${currentLevel.mediaUrl}" alt="Level ${currentLevel.number}" style="max-width: 100%; height: auto; margin: 1rem 0;">`;
            } else if (currentLevel.mediaType === 'video') {
                mediaContainer.innerHTML = `<video controls style="max-width: 100%; height: auto; margin: 1rem 0;"><source src="${currentLevel.mediaUrl}" type="video/mp4"></video>`;
            }
        }
    } catch (error) {
        console.error('Failed to load current level:', error);
        document.getElementById('levelDescription').textContent = 'Failed to load level content.';
    }
}

async function handleSubmit() {
    const answerInput = document.getElementById('answerInput');
    const feedback = document.getElementById('feedback');
    const answer = answerInput.value.trim();
    
    if (!answer) return;

    try {
        feedback.textContent = 'Checking answer...';
        const secret = await getSecret('POST');
        const response = await fetch('/api/submit-answer', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            },
            body: JSON.stringify({
                levelId: currentLevel.id,
                answer: answer
            })
        });

        const result = await response.json();
        
        if (result.correct) {
            feedback.textContent = 'Correct! Loading next level...';
            feedback.style.color = '#28a745';
            setTimeout(() => {
                loadCurrentLevel();
                answerInput.value = '';
                feedback.textContent = '';
                feedback.style.color = 'var(--primary)';
            }, 2000);
        } else {
            feedback.textContent = result.message || 'Incorrect answer. Try again.';
            feedback.style.color = '#dc3545';
            setTimeout(() => {
                feedback.textContent = '';
                feedback.style.color = 'var(--primary)';
            }, 3000);
        }
    } catch (error) {
        console.error('Failed to submit answer:', error);
        feedback.textContent = 'Failed to submit answer. Please try again.';
        feedback.style.color = '#dc3545';
    }
}

async function handleLogout() {
    try {
        const secret = await getSecret('POST');
        await fetch('/api/auth/logout', { 
            method: 'POST',
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        window.location.href = '/auth';
    } catch (error) {
        console.error('Logout failed:', error);
        window.location.href = '/auth';
    }
}

async function checkNotifications() {
    try {
        const secret = await getSecret('GET');
        const response = await fetch('/api/notifications/unread-count', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        const data = await response.json();
        const notificationDot = document.getElementById('notificationDot');
        
        if (data.count > 0) {
            notificationDot.classList.add('show');
        } else {
            notificationDot.classList.remove('show');
        }
    } catch (error) {
        console.error('Failed to check notifications:', error);
    }
}

document.addEventListener('DOMContentLoaded', function() {
    loadUserSession();
    checkNotifications();
    setInterval(checkNotifications, 30000);
});