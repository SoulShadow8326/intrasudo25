function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

async function loadHints() {
    try {
        const secret = await getSecret('GET');
        const response = await fetch('/api/question', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        if (!response.ok) {
            document.getElementById('hintsContainer').innerHTML = '<div class="hint-item" style="text-align: center; color: rgba(255, 255, 255, 0.7);">No question available yet</div>';
            return;
        }
        const data = await response.json();
        const container = document.getElementById('hintsContainer');
        if (data && data.question) {
            const q = data.question;
            if ((q.levelNumber !== undefined && q.markdown !== undefined && q.markdown !== null && q.markdown !== "")) {
                container.innerHTML = `<div class="hint-item"><h3 class="hint-title">Level ${q.levelNumber}</h3><p class="hint-text">${q.markdown}</p></div>`;
                return;
            }
        }
        document.getElementById('hintsContainer').innerHTML = '<div class="hint-item" style="text-align: center; color: rgba(255, 255, 255, 0.7);">No question available yet</div>';
    } catch (error) {
        document.getElementById('hintsContainer').innerHTML = '<div class="hint-item" style="text-align: center; color: #dc3545;">Failed to load hints</div>';
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
    } catch (error) {}
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

document.addEventListener('DOMContentLoaded', function() {
    checkAdminAccess();
    loadHints();
    checkNotifications();
    setInterval(checkNotifications, 30000);
});
