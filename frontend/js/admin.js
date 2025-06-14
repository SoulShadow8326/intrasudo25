let userSession = null;
let levelsData = {};

async function initializeAdmin() {
    try {
        await getUserSession();
        await Promise.all([
            loadStats(),
            loadLevels(),
            loadUsers(),
            loadAnnouncements()
        ]);
    } catch (error) {
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
        showNotification('Failed to load statistics', 'error');
    }
}

function updateStatsDisplay(stats) {
    document.getElementById('totalUsers').textContent = stats.totalUsers || 0;
    document.getElementById('totalLevels').textContent = stats.totalLevels || 0;
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
        loading.style.display = 'none';
        empty.style.display = 'block';
        hideQuestionStateSection();
        showNotification('Failed to load levels', 'error');
    }
}

function renderLevels(levels) {
    const list = document.getElementById('levelsList');
    
    levelsData = {};
    levels.forEach(level => {
        levelsData[level.id] = level;
    });
    
    list.innerHTML = levels.map(level => {
        const questions = (level.question || '').split('\n\n').filter(q => q.trim());
        return `
            <div class="level-card" data-level-id="${level.id}">
                <div class="level-card-header">
                    <div class="level-info-inline">
                        <div class="level-number-badge">${level.number}</div>
                        <h4 class="level-title-inline">Level ${level.number}</h4>
                    </div>
                    <div class="level-actions-compact">
                        <button class="btn-secondary" onclick="toggleEditLevel(${level.id})">Edit</button>
                        <button class="btn-danger" onclick="deleteLevel(${level.id})">Delete</button>
                    </div>
                </div>
                
                <div class="level-status-badge ${level.active ? 'status-active' : 'status-inactive'}">
                    ${level.active ? 'Active' : 'Inactive'}
                </div>
                
                <div class="level-content">
                    <div class="level-questions-preview">
                        <div class="question-preview">
                            <div class="question-label">Question</div>
                            <p class="question-text">${level.question || 'No question set'}</p>
                        </div>
                    </div>
                    
                    <div class="level-answer-section">
                        <div class="answer-label">Answer</div>
                        <p class="answer-text">${level.answer}</p>
                    </div>
                </div>

                <div class="edit-form-inline" id="editForm_${level.id}">
                    <div class="form-group">
                        <label class="form-label">Level Number:</label>
                        <input type="number" class="form-input" id="editNumber_${level.id}" value="${level.number}" min="1">
                    </div>
                    <div class="form-group">
                        <label class="form-label">Question:</label>
                        <textarea class="form-input form-textarea edit-question-input" id="editQuestion_${level.id}" placeholder="Enter level question">${level.question || ''}</textarea>
                    </div>
                    <div class="form-group">
                        <label class="form-label">Answer:</label>
                        <input type="text" class="form-input" id="editAnswer_${level.id}" value="${level.answer}">
                    </div>
                    <div class="form-group">
                        <label class="form-label">
                            <input type="checkbox" id="editActive_${level.id}" ${level.active ? 'checked' : ''}> Active
                        </label>
                    </div>
                    <div class="form-actions">
                        <button class="btn-secondary" onclick="cancelEditLevel(${level.id})">Cancel</button>
                        <button class="btn-primary" onclick="updateLevel(${level.id})">Update Level</button>
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

async function createLevel() {
    const levelNumber = document.getElementById('levelNumber').value;
    const levelQuestion = document.getElementById('levelQuestion').value.trim();
    const levelAnswer = document.getElementById('levelAnswer').value.trim();

    if (!levelNumber || !levelAnswer) {
        showNotification('Please fill in level number and answer.', 'error');
        return;
    }

    if (!levelQuestion) {
        showNotification('Please add a level question.', 'error');
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
                showNotification('Failed to delete level. Please try again.', 'error');
            }
        }
    );
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
                        ${!user.IsAdmin ? `
                            <button class="btn-secondary" onclick="resetUserLevel('${user.Gmail}')">Reset Level</button>
                            <button class="btn-warning" onclick="banUserEmail('${user.Gmail}')">Ban Email</button>
                            <button class="btn-danger" onclick="deleteUser('${user.Gmail}')">Delete</button>
                        ` : ''}
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
                showNotification('Failed to delete user. Please try again.', 'error');
            }
        }
    );
}

async function banUserEmail(email) {
    showConfirmModal(
        'Ban Email', 
        `Are you sure you want to ban the email ${email}? This will prevent this email from registering or logging in.`,
        async function() {
            try {
                const response = await fetch(`/api/admin/users/${encodeURIComponent(email)}/ban`, {
                    method: 'POST',
                    headers: {
                        'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
                    }
                });

                if (response.ok) {
                    showNotification('Email banned successfully!', 'success');
                    loadUsers();
                } else {
                    throw new Error('Failed to ban email');
                }
            } catch (error) {
                showNotification('Failed to ban email. Please try again.', 'error');
            }
        }
    );
}

async function resetUserLevel(email) {
    showConfirmModal(
        'Reset User Level', 
        `Are you sure you want to reset the level for user ${email}? This will set their level back to 1.`,
        async function() {
            try {
                const csrfToken = getCookie('X-CSRF_COOKIE');
                const response = await fetch(`/api/admin/users/${encodeURIComponent(email)}/reset-level`, {
                    method: 'POST',
                    headers: {
                        'CSRFtok': csrfToken
                    }
                });

                if (response.ok) {
                    showNotification(`Level for ${email} has been reset successfully!`, 'success');
                    loadUsers();
                } else {
                    throw new Error('Failed to reset user level');
                }
            } catch (error) {
                showNotification('Failed to reset user level. Please try again.', 'error');
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
                    'CSRFtok': getCookie('X_CSRF_COOKIE') || userSession?.csrfToken || ''
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
            showNotification('Failed to update question states. Please try again.', 'error');
        }
    });
}

function toggleAddLevelForm() {
    const form = document.getElementById('addLevelForm');
    if (form.classList.contains('show')) {
        form.classList.remove('show');
    } else {
        form.classList.add('show');
        clearAddLevelForm();
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
    document.getElementById('editLevelAnswer').value = '';
    document.getElementById('editLevelActive').checked = false;
    delete document.getElementById('editLevelForm').dataset.levelId;
    const container = document.getElementById('editQuestionsContainer');
    container.innerHTML = `
        <div class="question-input-group">
            <textarea class="form-input form-textarea level-question" placeholder="Enter the level question or description"></textarea>
            <button type="button" class="btn-add-question" onclick="addEditQuestionInput()">+</button>
        </div>
    `;
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
            window.location.href = '/auth';
        }
    } catch (error) {
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

function toggleLevelExpand(button) {
    const levelItem = button.closest('.level-item');
    const expandedSection = levelItem.querySelector('.level-questions-expanded');
    const expandIcon = button.querySelector('.expand-icon');
    const expandText = button.querySelector('.expand-text');
    
    if (expandedSection.style.display === 'none') {
        expandedSection.style.display = 'block';
        expandIcon.textContent = '▲';
        expandText.textContent = 'Hide Questions';
    } else {
        expandedSection.style.display = 'none';
        expandIcon.textContent = '▼';
        expandText.textContent = 'Show Questions';
    }
}

// Announcement management functions
async function loadAnnouncements() {
    try {
        const response = await fetch('/api/admin/announcements');
        if (response.ok) {
            const announcements = await response.json();
            displayAnnouncements(announcements);
        } else {
            throw new Error('Failed to load announcements');
        }
    } catch (error) {
        showNotification('Failed to load announcements', 'error');
    }
}

function displayAnnouncements(announcements) {
    const container = document.getElementById('announcementsContainer');
    
    if (!announcements || announcements.length === 0) {
        container.innerHTML = '<div class="empty-state">No announcements found</div>';
        return;
    }

    const announcementsHTML = announcements.map(announcement => `
        <div class="announcement-item" data-id="${announcement.id}">
            <div class="announcement-content">
                <h3 class="announcement-heading">${escapeHtml(announcement.heading)}</h3>
                <div class="announcement-meta">
                    Created: ${new Date(announcement.created_at).toLocaleDateString()}
                    ${announcement.updated_at !== announcement.created_at ? 
                        `• Updated: ${new Date(announcement.updated_at).toLocaleDateString()}` : ''}
                </div>
            </div>
            <div class="announcement-actions">
                <button class="btn-edit-announcement" onclick="editAnnouncement(${announcement.id}, '${escapeHtml(announcement.heading).replace(/'/g, "\\'")}')">
                    Edit
                </button>
                <button class="btn-delete-announcement" onclick="deleteAnnouncement(${announcement.id})">
                    Delete
                </button>
            </div>
        </div>
    `).join('');

    container.innerHTML = announcementsHTML;
}

function toggleAddAnnouncementForm() {
    const addForm = document.getElementById('addAnnouncementForm');
    const editForm = document.getElementById('editAnnouncementForm');
    
    editForm.style.display = 'none';
    addForm.classList.toggle('show');
    
    if (addForm.classList.contains('show')) {
        document.getElementById('announcementHeading').focus();
    } else {
        clearAddAnnouncementForm();
    }
}

function clearAddAnnouncementForm() {
    document.getElementById('announcementHeading').value = '';
}

function cancelAddAnnouncement() {
    const addForm = document.getElementById('addAnnouncementForm');
    const editForm = document.getElementById('editAnnouncementForm');
    
    addForm.classList.remove('show');
    editForm.style.display = 'none';
    clearAddAnnouncementForm();
}

async function createAnnouncement() {
    const heading = document.getElementById('announcementHeading').value.trim();
    
    if (!heading) {
        showNotification('Please enter an announcement heading', 'error');
        return;
    }

    try {
        const response = await fetch('/api/admin/announcements', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || userSession?.csrfToken || ''
            },
            body: JSON.stringify({ heading })
        });

        if (response.ok) {
            showNotification('Announcement created successfully', 'success');
            cancelAddAnnouncement();
            await loadAnnouncements();
        } else {
            const error = await response.text();
            throw new Error(error);
        }
    } catch (error) {
        showNotification('Failed to create announcement', 'error');
    }
}

async function editAnnouncement(id, currentHeading) {
    const addForm = document.getElementById('addAnnouncementForm');
    const editForm = document.getElementById('editAnnouncementForm');
    
    addForm.classList.remove('show');
    editForm.style.display = 'block';
    editForm.dataset.announcementId = id;
    
    document.getElementById('editAnnouncementHeading').value = currentHeading;
    document.getElementById('editAnnouncementHeading').focus();
}

function cancelEditAnnouncement() {
    const editForm = document.getElementById('editAnnouncementForm');
    editForm.style.display = 'none';
    delete editForm.dataset.announcementId;
    document.getElementById('editAnnouncementHeading').value = '';
}

async function updateAnnouncement() {
    const editForm = document.getElementById('editAnnouncementForm');
    const id = editForm.dataset.announcementId;
    const newHeading = document.getElementById('editAnnouncementHeading').value.trim();
    
    if (!newHeading) {
        showNotification('Please enter an announcement heading', 'error');
        return;
    }

    try {
        const response = await fetch(`/api/admin/announcements/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X_CSRF_COOKIE') || userSession?.csrfToken || ''
            },
            body: JSON.stringify({ heading: newHeading })
        });

        if (response.ok) {
            showNotification('Announcement updated successfully', 'success');
            cancelEditAnnouncement();
            await loadAnnouncements();
        } else {
            const error = await response.text();
            throw new Error(error);
        }
    } catch (error) {
        showNotification('Failed to update announcement', 'error');
    }
}

async function deleteAnnouncement(id) {
    showConfirmModal(
        'Delete Announcement',
        'Are you sure you want to delete this announcement? This action cannot be undone.',
        async function() {
            try {
                const response = await fetch(`/api/admin/announcements/${id}`, {
                    method: 'DELETE',
                    headers: {
                        'CSRFtok': getCookie('X_CSRF_COOKIE') || userSession?.csrfToken || ''
                    }
                });

                if (response.ok) {
                    showNotification('Announcement deleted successfully', 'success');
                    await loadAnnouncements();
                } else {
                    const error = await response.text();
                    throw new Error(error);
                }
            } catch (error) {
                showNotification('Failed to delete announcement', 'error');
            }
        }
    );
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

async function resetMyLevel() {
    showConfirmModal(
        'Reset Level',
        'Are you sure you want to reset your game level to 1? This action cannot be undone.',
        async function() {
            try {
                const response = await fetch('/api/admin/users/reset-my-level', {
                    method: 'POST',
                    headers: {
                        'CSRFtok': getCookie('X_CSRF_COOKIE') || userSession?.csrfToken || ''
                    }
                });

                if (response.ok) {
                    showNotification('Your level has been reset to 1', 'success');
                } else {
                    const error = await response.text();
                    throw new Error(error);
                }
            } catch (error) {
                showNotification('Failed to reset level', 'error');
            }
        }
    );
}

function toggleEditLevel(levelId) {
    const editForm = document.getElementById(`editForm_${levelId}`);
    const allEditForms = document.querySelectorAll('.edit-form-inline');
    
    allEditForms.forEach(form => {
        if (form.id !== `editForm_${levelId}`) {
            form.classList.remove('show');
        }
    });
    
    editForm.classList.toggle('show');
}

function cancelEditLevel(levelId) {
    const editForm = document.getElementById(`editForm_${levelId}`);
    editForm.classList.remove('show');
}

async function updateLevel(levelId) {
    const levelQuestion = document.getElementById(`editQuestion_${levelId}`).value.trim();
    const levelNumber = document.getElementById(`editNumber_${levelId}`).value;
    const levelAnswer = document.getElementById(`editAnswer_${levelId}`).value.trim();
    const levelActive = document.getElementById(`editActive_${levelId}`).checked;

    if (!levelNumber || !levelAnswer) {
        showNotification('Please fill in level number and answer.', 'error');
        return;
    }

    if (!levelQuestion) {
        showNotification('Please add a level question.', 'error');
        return;
    }

    const requestData = {
        level_number: levelNumber,
        title: `Level ${levelNumber}`,
        markdown: levelQuestion,
        answer: levelAnswer,
        active: levelActive.toString()
    };

    try {
        const response = await fetch(`/api/admin/levels/${levelId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X_CSRF_COOKIE') || userSession?.csrfToken || ''
            },
            body: JSON.stringify(requestData)
        });

        if (response.ok) {
            showNotification('Level updated successfully!', 'success');
            cancelEditLevel(levelId);
            loadLevels();
            loadStats();
        } else {
            const errorData = await response.json();
            showNotification(errorData.error || 'Failed to update level', 'error');
        }
    } catch (error) {
        showNotification('Failed to update level. Please try again.', 'error');
    }
}
