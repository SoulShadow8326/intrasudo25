function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

async function loadLeaderboard() {
    try {
        const secret = await getSecret();
        const response = await fetch('/api/leaderboard', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        const leaderboard = await response.json();
        const listContainer = document.getElementById('leaderboardList');
        
        checkAdminAccess();
        
        if (leaderboard.length === 0) {
            listContainer.innerHTML = '<div class="leaderboard-entry" style="text-align: center; padding: 2rem; color: rgba(255, 255, 255, 0.7);">No participants yet</div>';
            return;
        }
        
        listContainer.innerHTML = leaderboard.map((entry, index) => `
            <div class="leaderboard-entry">
                <div class="entry-left">
                    <span class="rank">${index + 1}</span>
                    <span class="name">${entry.username}</span>
                </div>
                <div class="entry-right">
                    <span class="level">${entry.currentLevel}</span>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Failed to load leaderboard:', error);
        document.getElementById('leaderboardList').innerHTML = 
            '<div class="leaderboard-entry" style="text-align: center; padding: 2rem; color: #dc3545;">Failed to load leaderboard</div>';
    }
}

async function checkNotifications() {
    try {
        const secret = await getSecret();
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

async function checkAdminAccess() {
    try {
        const secret = await getSecret();
        const response = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        if (response.ok) {
            const userData = await response.json();
            if (userData.isAdmin) {
                document.getElementById('adminLink').style.display = 'block';
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
    loadLeaderboard();
    checkNotifications();
    setInterval(loadLeaderboard, 30000);
    setInterval(checkNotifications, 30000);
});
