let userSession = null;

async function initializeAdmin() {
    try {
        await getUserSession();
        await Promise.all([
            loadStats(),
            loadLevels(),
            loadUsers()
        ]);
    } catch (error) {
        console.error('Failed to initialize admin dashboard:', error);
        showNotification('Failed to load admin dashboard', 'error');
    }
}

async function getUserSession() {
    try {
        const response = await fetch('/api/user/session');
        if (response.ok) {
            userSession = await response.json();
        }
    } catch (error) {
        console.error('Failed to get user session:', error);
    }
}

async function loadStats() {
    try {
        const response = await fetch('/api/admin/stats');
        if (response.ok) {
            const stats = await response.json();
            updateStatsDisplay(stats);
        } else {
            throw new Error('Failed to load stats');
        }
    } catch (error) {
        console.error('Failed to load stats:', error);
        showNotification('Failed to load statistics', 'error');
    }
}

function updateStatsDisplay(stats) {
    document.getElementById('totalUsers').textContent = stats.totalUsers || 0;
    document.getElementById('totalLevels').textContent = stats.totalLevels || 0;
    document.getElementById('activeUsers').textContent = stats.activeUsers || 0;
}

async function loadLevels() {
    const container = document.getElementById('levelsContainer');
    const loading = document.getElementById('levelsLoading');
    const list = document.getElementById('levelsList');
    const empty = document.getElementById('levelsEmpty');

    try {
        loading.style.display = 'block';
        list.style.display = 'none';
        empty.style.display = 'none';

        const response = await fetch('/api/admin/levels');
        if (response.ok) {
            const levels = await response.json();
            
            if (levels && levels.length > 0) {
                renderLevels(levels);
                loading.style.display = 'none';
                list.style.display = 'block';
                showQuestionStateSection();
                renderQuestionStates(levels);
            } else {
                loading.style.display = 'none';
                empty.style.display = 'block';
                hideQuestionStateSection();
            }
        } else {
            throw new Error('Failed to load levels');
        }
    } catch (error) {
        console.error('Failed to load levels:', error);
        loading.style.display = 'none';
        empty.style.display = 'block';
        hideQuestionStateSection();
        showNotification('Failed to load levels', 'error');
    }
}

function renderLevels(levels) {
    const list = document.getElementById('levelsList');
    list.innerHTML = levels.map(level => `
        <div class="level-item">
            <div class="level-info">
                <div class="level-number">${level.number}</div>
                <div class="level-details">
                    <h4 class="level-title">${level.title}</h4>
                    <p class="level-description">${level.question.substring(0, 50)}${level.question.length > 50 ? '...' : ''}</p>
                    <div class="level-meta">
                        <span class="status-badge ${level.active ? 'status-active' : 'status-inactive'}">
                            ${level.active ? 'Active' : 'Inactive'}
                        </span>
                    </div>
                </div>
            </div>
            <div class="level-actions">
                <button class="btn-secondary" onclick="editLevel(${level.id}, '${level.question.replace(/'/g, "\\'")}', '${level.answer.replace(/'/g, "\\'")}', ${level.active})">Edit</button>
                <button class="btn-danger" onclick="deleteLevel(${level.id})">Delete</button>
            </div>
        </div>
    `).join('');
}

async function createLevel() {
    const levelNumber = document.getElementById('levelNumber').value;
    const levelQuestion = document.getElementById('levelQuestion').value.trim();
    const levelAnswer = document.getElementById('levelAnswer').value.trim();

    if (!levelNumber || !levelQuestion || !levelAnswer) {
        showNotification('Please fill in all required fields.', 'error');
        return;
    }

    const requestData = {
        level_number: levelNumber,
        title: `Level ${levelNumber}`,
        markdown: levelQuestion,
        answer: levelAnswer,
        active: "true"
    };

    try {
        const response = await fetch('/api/admin/levels', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || userSession?.csrfToken || ''
            },
            body: JSON.stringify(requestData)
        });

        if (response.ok) {
            showNotification('Level created successfully!', 'success');
            cancelAddLevel();
            loadLevels();
            loadStats();
        } else {
            const errorData = await response.json();
            showNotification(errorData.error || 'Failed to create level', 'error');
        }
    } catch (error) {
        console.error('Error creating level:', error);
        showNotification('Failed to create level. Please try again.', 'error');
    }
}

async function deleteLevel(levelId) {
    showConfirmModal(
        'Delete Level', 
        'Are you sure you want to delete this level? This action cannot be undone.',
        async function() {
            try {
                const response = await fetch(`/api/admin/levels/${levelId}`, {
                    method: 'DELETE',
                    headers: {
                        'Content-Type': 'application/json',
                        'CSRFtok': getCookie('X-CSRF_COOKIE') || userSession?.csrfToken || ''
                    }
                });

                if (response.ok) {
                    showNotification('Level deleted successfully!', 'success');
                    loadLevels();
                    loadStats();
                } else {
                    throw new Error('Failed to delete level');
                }
            } catch (error) {
                console.error('Error deleting level:', error);
                showNotification('Failed to delete level. Please try again.', 'error');
            }
        }
    );
}

function editLevel(levelId, question, answer, active) {
    document.getElementById('editLevelNumber').value = levelId;
    document.getElementById('editLevelQuestion').value = question;
    document.getElementById('editLevelAnswer').value = answer;
    document.getElementById('editLevelActive').checked = active;
    document.getElementById('editLevelForm').dataset.levelId = levelId;
    document.getElementById('addLevelForm').classList.remove('show');
    document.getElementById('editLevelForm').style.display = 'block';
}

async function updateLevel() {
    const levelId = document.getElementById('editLevelForm').dataset.levelId;
    const levelQuestion = document.getElementById('editLevelQuestion').value.trim();
    const levelAnswer = document.getElementById('editLevelAnswer').value.trim();
    const levelActive = document.getElementById('editLevelActive').checked;

    if (!levelQuestion || !levelAnswer) {
        showNotification('Please fill in all required fields.', 'error');
        return;
    }

    const requestData = {
        markdown: levelQuestion,
        answer: levelAnswer,
        active: levelActive.toString()
    };

    try {
        const response = await fetch(`/api/admin/levels/${levelId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || userSession?.csrfToken || ''
            },
            body: JSON.stringify(requestData)
        });

        if (response.ok) {
            showNotification('Level updated successfully!', 'success');
            cancelEditLevel();
            loadLevels();
            loadStats();
        } else {
            const errorData = await response.json();
            showNotification(errorData.error || 'Failed to update level', 'error');
        }
    } catch (error) {
        console.error('Error updating level:', error);
        showNotification('Failed to update level. Please try again.', 'error');
    }
}

async function loadUsers() {
    const container = document.getElementById('usersContainer');
    
    try {
        container.innerHTML = '<div class="loading-state">Loading users...</div>';
        
        const response = await fetch('/api/admin/users');
        if (response.ok) {
            const data = await response.json();
            const users = data.users || data;
            renderUsers(users);
        } else {
            throw new Error('Failed to load users');
        }
    } catch (error) {
        console.error('Failed to load users:', error);
        container.innerHTML = '<div class="error-state">Unable to load users. Please try again.</div>';
        showNotification('Failed to load users', 'error');
    }
}

function renderUsers(users) {
    const container = document.getElementById('usersContainer');
    
    if (users.length === 0) {
        container.innerHTML = '<div class="empty-state"><div class="empty-state-icon">Users</div><p>No users found.</p></div>';
        return;
    }
    
    container.innerHTML = `
        <div class="user-list">
            ${users.map(user => `
                <div class="user-item">
                    <div class="user-info">
                        <div class="user-details">
                            <h4 class="user-name">${user.Gmail}</h4>
                            <p class="user-email">${user.Gmail}</p>
                            <p class="user-level">Current Level: ${user.On}</p>
                            <p class="user-status ${user.Verified ? 'verified' : 'unverified'}">
                                ${user.Verified ? 'Verified' : 'Unverified'}
                            </p>
                            ${user.IsAdmin ? '<p class="user-admin">Admin</p>' : ''}
                        </div>
                    </div>
                    <div class="user-actions">
                        ${!user.IsAdmin ? `<button class="btn-danger" onclick="deleteUser('${user.Gmail}')">Delete</button>` : ''}
                    </div>
                </div>
            `).join('')}
        </div>
    `;
}

async function deleteUser(email) {
    showConfirmModal(
        'Delete User', 
        `Are you sure you want to delete user ${email}? This action cannot be undone.`,
        async function() {
            try {
                const response = await fetch(`/api/admin/users/${encodeURIComponent(email)}`, {
                    method: 'DELETE',
                    headers: {
                        'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
                    }
                });

                if (response.ok) {
                    showNotification('User deleted successfully!', 'success');
                    loadUsers();
                    loadStats();
                } else {
                    throw new Error('Failed to delete user');
                }
            } catch (error) {
                console.error('Error deleting user:', error);
                showNotification('Failed to delete user. Please try again.', 'error');
            }
        }
    );
}

function showQuestionStateSection() {
    document.getElementById('questionStateSection').style.display = 'block';
}

function hideQuestionStateSection() {
    document.getElementById('questionStateSection').style.display = 'none';
}

function renderQuestionStates(levels) {
    const list = document.getElementById('questionStateList');
    list.innerHTML = levels.map(level => `
        <div class="question-state-item">
            <div class="question-info">
                <div class="question-number">Level ${level.number}</div>
                <div class="question-title">${level.title}</div>
            </div>
            <div class="question-state-toggle">
                <label class="toggle-switch">
                    <input type="checkbox" ${level.enabled ? 'checked' : ''} onchange="toggleQuestionState(${level.id}, this.checked)">
                    <span class="toggle-slider"></span>
                </label>
                <span class="toggle-label">${level.enabled ? 'Enabled' : 'Disabled'}</span>
            </div>
        </div>
    `).join('');
}

async function toggleQuestionState(levelId, enabled) {
    try {
        const response = await fetch(`/api/admin/levels/${levelId}/state`, {
            method: 'PATCH',
            headers: { 
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || userSession?.csrfToken || ''
            },
            body: JSON.stringify({ enabled })
        });

        if (response.ok) {
            showNotification(`Question ${enabled ? 'enabled' : 'disabled'} successfully!`, 'success');
            loadLevels();
        } else {
            throw new Error('Failed to update question state');
        }
    } catch (error) {
        console.error('Error updating question state:', error);
        showNotification('Failed to update question state. Please try again.', 'error');
        loadLevels();
    }
}

async function toggleAllQuestions(enabled) {
    const confirmTitle = enabled ? 'Enable All Questions' : 'Disable All Questions';
    const confirmMessage = enabled ? 
        'Are you sure you want to enable all questions?' : 
        'Are you sure you want to disable all questions?';
    
    showConfirmModal(confirmTitle, confirmMessage, async function() {
        try {
            const response = await fetch('/api/admin/levels/bulk-state', {
                method: 'PATCH',
                headers: { 
                    'Content-Type': 'application/json',
                    'CSRFtok': getCookie('X-CSRF_COOKIE') || userSession?.csrfToken || ''
                },
                body: JSON.stringify({ enabled })
            });

            if (response.ok) {
                showNotification(`All questions ${enabled ? 'enabled' : 'disabled'} successfully!`, 'success');
                loadLevels();
            } else {
                throw new Error('Failed to update question states');
            }
        } catch (error) {
            console.error('Error updating question states:', error);
            showNotification('Failed to update question states. Please try again.', 'error');
        }
    });
}

function toggleAddLevelForm() {
    const form = document.getElementById('addLevelForm');
    const isVisible = form.classList.contains('show');
    
    if (isVisible) {
        form.classList.remove('show');
    } else {
        form.classList.add('show');
        document.getElementById('levelNumber').focus();
    }
}

function cancelAddLevel() {
    const form = document.getElementById('addLevelForm');
    form.classList.remove('show');
    clearAddLevelForm();
}

function clearAddLevelForm() {
    document.getElementById('levelNumber').value = '';
    document.getElementById('levelQuestion').value = '';
    document.getElementById('levelAnswer').value = '';
}

function cancelEditLevel() {
    document.getElementById('editLevelForm').style.display = 'none';
    clearEditLevelForm();
}

function clearEditLevelForm() {
    document.getElementById('editLevelNumber').value = '';
    document.getElementById('editLevelQuestion').value = '';
    document.getElementById('editLevelAnswer').value = '';
    document.getElementById('editLevelActive').checked = false;
    delete document.getElementById('editLevelForm').dataset.levelId;
}

async function refreshUsers() {
    await loadUsers();
    showNotification('Users refreshed!', 'success');
}

async function handleLogout() {
    try {
        const response = await fetch('/api/auth/logout', {
            method: 'POST'
        });
        
        if (response.ok) {
            window.location.href = '/auth';
        } else {
            console.error('Logout failed');
        }
    } catch (error) {
        console.error('Error during logout:', error);
        window.location.href = '/auth';
    }
}

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    
    const bgColor = type === 'success' ? '#22c55e' : type === 'error' ? '#ef4444' : '#2977f5';
    const borderColor = type === 'success' ? 'rgba(34, 197, 94, 0.3)' : type === 'error' ? 'rgba(239, 68, 68, 0.3)' : 'rgba(41, 119, 245, 0.3)';
    
    notification.style.cssText = `
        position: fixed;
        top: 80px;
        right: 20px;
        padding: 1rem 1.5rem;
        border-radius: 0.75rem;
        color: white;
        z-index: 1001;
        background: ${bgColor};
        border: 1px solid ${borderColor};
        animation: slideInRight 0.3s ease-out;
        font-weight: 500;
        font-size: 0.9rem;
        min-width: 200px;
        max-width: 300px;
        backdrop-filter: blur(10px);
    `;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.animation = 'slideOutRight 0.3s ease-in';
        setTimeout(() => notification.remove(), 300);
    }, 4000);
}

let confirmCallback = null;

function showConfirmModal(title, message, confirmAction) {
    const modal = document.getElementById('confirmModal');
    const titleEl = document.getElementById('confirmTitle');
    const messageEl = document.getElementById('confirmMessage');
    const confirmBtn = document.getElementById('confirmActionBtn');
    
    titleEl.textContent = title;
    messageEl.textContent = message;
    confirmCallback = confirmAction;
    
    modal.classList.add('show');
    document.body.style.overflow = 'hidden';
    
    setTimeout(() => confirmBtn.focus(), 100);
    document.addEventListener('keydown', handleEscKey);
}

function closeConfirmModal() {
    const modal = document.getElementById('confirmModal');
    modal.classList.remove('show');
    document.body.style.overflow = '';
    confirmCallback = null;
    
    document.removeEventListener('keydown', handleEscKey);
}

function handleEscKey(event) {
    if (event.key === 'Escape') {
        closeConfirmModal();
    }
}

function confirmAction() {
    if (typeof confirmCallback === 'function') {
        confirmCallback();
    }
    closeConfirmModal();
}

document.addEventListener('DOMContentLoaded', function() {
    initializeAdmin();
});
