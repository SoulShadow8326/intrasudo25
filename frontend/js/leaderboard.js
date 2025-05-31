function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

async function loadLeaderboard() {
    try {
        const secret = await getSecret('GET');
        if (!secret) {
            throw new Error('Authentication service unavailable');
        }
        
        const response = await fetch('/api/leaderboard', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'X-secret': secret
            }
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error('Response not OK:', response.status, errorText);
            throw new Error(`Failed to fetch leaderboard data: ${response.status} ${errorText}`);
        }
        
        const data = await response.json();
        console.log('Leaderboard data received:', data);
        const leaderboardData = data.leaderboard || [];
        const listContainer = document.getElementById('leaderboardList');
        
        await checkAdminAccess();
        
        if (leaderboardData.length === 0) {
            listContainer.innerHTML = '<div class="leaderboard-entry" style="text-align: center; padding: 2rem; color: rgba(255, 255, 255, 0.7);">No participants yet</div>';
            return Promise.resolve();
        }
        
        listContainer.innerHTML = leaderboardData.map((entry, index) => {
            const rank = index + 1;
            let rankClass = '';
            
            if (rank === 1) {
                rankClass = 'rank-first';
            } else if (rank === 2) {
                rankClass = 'rank-second';
            } else if (rank === 3) {
                rankClass = 'rank-third';
            }
            
            const username = entry.Gmail.split('@')[0];
            
            return `
                <div class="leaderboard-entry ${rank <= 3 ? 'top-three' : ''}">
                    <span class="rank ${rankClass}">${rank}</span>
                    <span class="name" title="${username}">${username}</span>
                    <span class="level">Level ${entry.On || 1}</span>
                </div>
            `;
        }).join('');
    } catch (error) {
        console.error('Failed to load leaderboard:', error);
        document.getElementById('leaderboardList').innerHTML = 
            '<div class="leaderboard-entry" style="text-align: center; padding: 2rem; color: #dc3545;">Failed to load leaderboard</div>';
        return Promise.resolve();
    }
    return Promise.resolve();
}

async function checkNotifications() {
    try {
        const secret = await getSecret('GET');
        if (!secret) return;
        
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

function forceReflow() {
    const wrapper = document.querySelector('.leaderboard-wrapper');
    const container = document.querySelector('.leaderboard-container');
    const header = document.querySelector('.leaderboard-header');
    const body = document.querySelector('.leaderboard-body');
    const list = document.querySelector('.leaderboard-list');
    const entries = document.querySelectorAll('.leaderboard-entry');
    
    if (wrapper && container && header && body) {
        const isMobile = window.innerWidth <= 768;
        
        if (isMobile) {
            wrapper.style.width = '100%';
            wrapper.style.minWidth = '320px';
            wrapper.style.maxWidth = '100%';
            container.style.width = '100%';
            container.style.minWidth = '320px';
            container.style.maxWidth = '100%';
            header.style.width = '100%';
            body.style.width = '100%';
            if (list) {
                list.style.width = '100%';
            }
            entries.forEach(entry => {
                entry.style.width = '100%';
                const nameCell = entry.querySelector('.name');
                if (nameCell) {
                    nameCell.style.width = 'auto';
                    nameCell.style.maxWidth = 'none';
                    nameCell.style.flex = '1';
                    nameCell.style.minWidth = '0';
                    nameCell.style.whiteSpace = 'nowrap';
                    nameCell.style.overflow = 'hidden';
                    nameCell.style.textOverflow = 'ellipsis';
                }
            });
        } else {
            const fixedWidth = 800;
            wrapper.style.width = `${fixedWidth}px`;
            wrapper.style.minWidth = `${fixedWidth}px`;
            wrapper.style.maxWidth = `${fixedWidth}px`;
            container.style.width = `${fixedWidth}px`;
            container.style.minWidth = `${fixedWidth}px`;
            container.style.maxWidth = `${fixedWidth}px`;
            header.style.width = '100%';
            body.style.width = '100%';
            if (list) {
                list.style.width = '100%';
            }
            entries.forEach(entry => {
                entry.style.width = '100%';
                const nameCell = entry.querySelector('.name');
                if (nameCell) {
                    nameCell.style.width = '400px';
                    nameCell.style.maxWidth = '400px';
                    nameCell.style.flex = '0 0 400px';
                    nameCell.style.minWidth = '400px';
                    nameCell.style.whiteSpace = 'nowrap';
                    nameCell.style.overflow = 'hidden';
                    nameCell.style.textOverflow = 'ellipsis';
                }
            });
        }
        
        container.style.opacity = '0.99';
        setTimeout(() => {
            container.style.opacity = '1';
        }, 50);
    }
}

document.addEventListener('DOMContentLoaded', function() {
    checkAdminAccess();
    loadLeaderboard().then(() => {
        setTimeout(forceReflow, 100);
        setTimeout(forceReflow, 500);
    });
    checkNotifications();
    setInterval(loadLeaderboard, 30000);
    setInterval(checkNotifications, 30000);
    window.addEventListener('resize', forceReflow);
    window.addEventListener('load', function() {
        forceReflow();
        setTimeout(forceReflow, 300);
    });
    if (typeof ResizeObserver !== 'undefined') {
        const wrapper = document.querySelector('.leaderboard-wrapper');
        if (wrapper) {
            const resizeObserver = new ResizeObserver(entries => {
                for (let entry of entries) {
                    if (entry.target === wrapper) {
                        forceReflow();
                    }
                }
            });
            resizeObserver.observe(wrapper);
        }
    }
});
