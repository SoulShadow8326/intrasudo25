let currentLevel = null;
let userSession = null;
let isSubmitting = false;

async function initializePage() {
    try {
        const sessionData = await loadUserSession();
        if (!sessionData) {
            window.location.href = '/auth';
            return;
        }

        await checkAdminAccess();
        await loadCurrentLevel();
        await checkNotifications();
        await updateHintsDisplay();
        
        setInterval(checkNotifications, 30000);
        setInterval(updateHintsDisplay, 30000);
        
    } catch (error) {
        window.location.href = '/auth';
    }
}

async function loadUserSession() {
    try {
        const response = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });

        if (response.ok) {
            userSession = await response.json();
            return userSession;
        } else {
            return null;
        }
    } catch (error) {
        return null;
    }
}

async function checkAdminAccess() {
    try {
        if (userSession && userSession.isAdmin) {
            const adminLink = document.getElementById('adminLink');
            const mobileAdminLink = document.getElementById('mobileAdminLink');
            if (adminLink) adminLink.style.display = 'inline-block';
            if (mobileAdminLink) mobileAdminLink.style.display = 'block';
        } else {
            const adminLink = document.getElementById('adminLink');
            const mobileAdminLink = document.getElementById('mobileAdminLink');
            if (adminLink) adminLink.style.display = 'none';
            if (mobileAdminLink) mobileAdminLink.style.display = 'none';
        }
    } catch (error) {
        const adminLink = document.getElementById('adminLink');
        const mobileAdminLink = document.getElementById('mobileAdminLink');
        if (adminLink) adminLink.style.display = 'none';
        if (mobileAdminLink) mobileAdminLink.style.display = 'none';
    }
}

async function loadCurrentLevel() {
    try {
        const response = await fetch('/api/user/current-level?' + Date.now(), {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || '',
                'Cache-Control': 'no-cache'
            }
        });

        if (!response.ok) {
            throw new Error(`API returned status ${response.status}`);
        }

        const newLevel = await response.json();
        currentLevel = newLevel;
        updateLevelDisplay();
        updateHintsDisplay();
        
    } catch (error) {
        console.error('Failed to load current level:', error);
        handleLevelLoadError(error);
        throw error;
    }
}

function updateLevelDisplay() {
    const levelTitle = document.getElementById('levelTitle');
    const existingDescription = document.getElementById('levelDescription');
    const mediaContainer = document.getElementById('levelMedia');
    const answerInput = document.getElementById('answerInput');
    const feedback = document.getElementById('feedback');
    
    if (existingDescription) {
        existingDescription.remove();
    }
    
    if (currentLevel && currentLevel.allCompleted) {
        if (levelTitle) {
            levelTitle.textContent = 'Congratulations!';
        }
        
        const levelContent = document.getElementById('levelContent');
        const completionMessage = document.createElement('div');
        completionMessage.id = 'levelDescription';
        completionMessage.style.cssText = 'margin-bottom: 2rem; text-align: center; color: var(--primary); font-size: 1.3rem; line-height: 1.6;';
        completionMessage.innerHTML = `
            <div style="margin-bottom: 20px;">
                <svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                    <polyline points="22 4 12 14.01 9 11.01"></polyline>
                </svg>
            </div>
            <p>${currentLevel.description}</p>
            <p style="margin-top: 20px; font-size: 1.1rem;">You've completed all ${currentLevel.maxLevel} levels!</p>
        `;
        
        if (levelContent) {
            levelContent.insertBefore(completionMessage, levelContent.firstChild);
        }
        
        if (mediaContainer) {
            mediaContainer.innerHTML = '';
        }
        
        if (answerInput) {
            answerInput.style.display = 'none';
        }
        
        if (feedback) {
            feedback.textContent = '';
        }
        
        return;
    }
    
    if (levelTitle) {
        levelTitle.textContent = `Level ${currentLevel.number}`;
    }
    
    if (currentLevel.mediaUrl) {
        if (mediaContainer) {
            if (currentLevel.mediaType === 'image') {
                mediaContainer.innerHTML = `<img src="${currentLevel.mediaUrl}" alt="Level ${currentLevel.number}" style="max-width: 100%; height: auto; margin: 1rem 0; border-radius: 8px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">`;
            } else if (currentLevel.mediaType === 'video') {
                mediaContainer.innerHTML = `<video controls style="max-width: 100%; height: auto; margin: 1rem 0; border-radius: 8px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);"><source src="${currentLevel.mediaUrl}" type="video/mp4"></video>`;
            }
        }
    } else {
        if (mediaContainer) {
            mediaContainer.innerHTML = '';
        }
    }
    
    if (feedback) {
        feedback.textContent = '';
    }
    
    if (answerInput) {
        answerInput.value = '';
        answerInput.focus();
    }
    
    updateHintsDisplay();
}

function handleLevelLoadError(error) {
    const levelTitle = document.getElementById('levelTitle');
    const levelDescription = document.getElementById('levelDescription');
    
    if (levelTitle) {
        levelTitle.textContent = 'Level Not Found';
    }
        
    if (!levelDescription) {
        levelDescription = document.createElement('div');
        levelDescription.id = 'levelDescription';
        levelDescription.style.cssText = 'margin-bottom: 2rem; text-align: center; color: var(--text-color); font-size: 1.1rem; line-height: 1.6;';
        
        const levelContent = document.getElementById('levelContent');
        if (levelContent) {
            levelContent.insertBefore(levelDescription, levelContent.firstChild);
        }
    }
    
    if (error.message.includes('401')) {
        levelDescription.textContent = 'Authentication required. Please log in again.';
        setTimeout(() => {
            window.location.href = '/auth';
        }, 1000);
    } else {
        levelDescription.textContent = 'Failed to load level content. Please try refreshing the page.';
    }
}

async function handleSubmit() {
    if (isSubmitting) {
        return;
    }
    
    if (!currentLevel) {
        const feedback = document.getElementById('feedback');
        feedback.textContent = 'No level loaded. Please refresh the page.';
        feedback.style.color = '#dc3545';
        return;
    }
    
    const answerInput = document.getElementById('answerInput');
    const feedback = document.getElementById('feedback');
    let answer = answerInput.value.trim();
    
    if (!answer) {
        feedback.textContent = 'Please enter an answer.';
        feedback.style.color = '#dc3545';
        setTimeout(() => {
            feedback.textContent = '';
            feedback.style.color = 'var(--primary)';
        }, 2000);
        return;
    }

    answer = answer.toLowerCase().replace(/\s+/g, '');
    if (!answer) {
        feedback.textContent = 'Answer cannot be empty after formatting. Please enter a valid answer.';
        feedback.style.color = '#dc3545';
        setTimeout(() => {
            feedback.textContent = '';
            feedback.style.color = 'var(--primary)';
        }, 2000);
        return;
    }

    if (!currentLevel) {
        feedback.textContent = 'No level loaded. Please refresh the page.';
        feedback.style.color = '#dc3545';
        return;
    }

    isSubmitting = true;
    
    const submitButton = document.querySelector('button[onclick="handleSubmit()"]');
    if (submitButton) {
        submitButton.disabled = true;
        submitButton.textContent = 'Submitting...';
    }
    
    try {
        const originalAnswer = answerInput.value.trim();
        if (originalAnswer !== answer) {
            feedback.textContent = `Validating answer: "${answer}" (formatted from "${originalAnswer}")...`;
        } else {
            feedback.textContent = 'Validating...';
        }
        feedback.style.color = 'var(--primary)';
        
        const response = await fetch('/api/submit-answer', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            },
            body: JSON.stringify({
                levelId: currentLevel.id,
                answer: answer
            })
        });

        if (!response.ok) {
            if (response.status === 401) {
                feedback.textContent = 'Session expired. Redirecting to login...';
                feedback.style.color = '#dc3545';
                setTimeout(() => {
                    window.location.href = '/auth';
                }, 1000);
                return;
            }
            throw new Error(`Server error: ${response.status}`);
        }

        const result = await response.json();
        
        if (result.correct) {
            feedback.textContent = 'Correct! Loading next level...';
            feedback.style.color = '#28a745';
            
            currentLevel = null;
            
            setTimeout(() => {
                window.location.reload();
            }, 150);
        } else if (result.reload_page) {
            feedback.textContent = result.message || 'Processing...';
            feedback.style.color = '#2977F5';
            setTimeout(() => {
                window.location.reload();
            }, 150);
        } else {
            feedback.textContent = result.message || 'Incorrect answer. Try again.';
            feedback.style.color = '#dc3545';
            
            const submitButton = document.querySelector('button[onclick="handleSubmit()"]');
            if (submitButton) {
                submitButton.disabled = false;
                submitButton.textContent = 'Submit Answer';
            }
            
            setTimeout(() => {
                feedback.textContent = '';
                feedback.style.color = 'var(--primary)';
                isSubmitting = false;
            }, 2000);
        }
    } catch (error) {
        feedback.textContent = 'Correct! Loading next level...';
        feedback.style.color = '#28a745';
        
        const submitButton = document.querySelector('button[onclick="handleSubmit()"]');
        if (submitButton) {
            submitButton.disabled = false;
            submitButton.textContent = 'Submit Answer';
        }
        
        setTimeout(() => {
            window.location.reload();
        }, 150);
    }
}

async function handleLogout() {
    try {
        await fetch('/api/auth/logout', { 
            method: 'POST',
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
    } catch (error) {
    } finally {
        window.location.href = '/auth';
    }
}

async function checkNotifications() {
    try {
        const response = await fetch('/api/notifications/unread-count', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            const logoNotification = document.getElementById('logoNotification');
            
            if (logoNotification) {
                logoNotification.style.display = 'none';
            }
        }
    } catch (error) {
    }
}

async function updateHintsDisplay() {
    try {
        const levelSpan = document.getElementById('levelNumber');
        const hintsSpan = document.getElementById('hintsStatus');
        
        if (currentLevel && !currentLevel.allCompleted) {
            if (levelSpan) {
                levelSpan.textContent = currentLevel.number;
            }
        }
        
        if (hintsSpan) {
            try {
                const response = await fetch('/api/hints', {
                    method: 'GET',
                    headers: {
                        'Content-Type': 'application/json',
                        'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
                    }
                });
                
                if (response.ok) {
                    const hints = await response.json();
                    const count = Array.isArray(hints) ? hints.length : 0;
                    
                    if (count === 0) {
                        hintsSpan.textContent = 'No hints posted';
                    } else if (count === 1) {
                        hintsSpan.textContent = '1 hint available';
                    } else {
                        hintsSpan.textContent = `${count} hints available`;
                    }
                } else {
                    hintsSpan.textContent = 'No hints posted';
                }
            } catch (error) {
                hintsSpan.textContent = 'No hints posted';
            }
        }
    } catch (error) {
    }
}

document.addEventListener('DOMContentLoaded', function() {
    initializePage();
});