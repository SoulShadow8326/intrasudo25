let currentLevel = null;
let userSession = null;
let isSubmitting = false;
let isRedirecting = false;

async function initializePage() {
    if (isRedirecting) return;
    
    try {
        const sessionData = await loadUserSession();
        if (!sessionData) {
            if (!isRedirecting) {
                isRedirecting = true;
                window.location.href = '/auth';
            }
            return;
        }

        await checkAdminAccess();
        await loadCurrentLevel();
        await checkNotifications();
        await updateHintsDisplay();
        
        setInterval(checkNotifications, 30000);
        setInterval(updateHintsDisplay, 30000);
        
    } catch (error) {
        console.error('Error initializing page:', error);
        if (!isRedirecting) {
            isRedirecting = true;
            window.location.href = '/auth';
        }
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
        } else if (response.status === 401) {
            console.log('Session expired, redirecting to auth');
            return null;
        } else {
            console.error('Session check failed with status:', response.status);
            return null;
        }
    } catch (error) {
        console.error('Error checking session:', error);
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
        await updateSourceHint();
        
        console.log('Level loaded:', currentLevel.number, 'updating source hint...');
        
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
        
        const levelQuestion = document.getElementById('levelQuestion');
        if (levelQuestion) {
            levelQuestion.style.display = 'none';
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
    
    const levelQuestion = document.getElementById('levelQuestion');
    if (levelQuestion) {
        if (currentLevel.markdown && currentLevel.markdown.trim()) {
            levelQuestion.textContent = currentLevel.markdown.trim();
            levelQuestion.style.display = 'block';
        } else {
            levelQuestion.style.display = 'none';
        }
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
    
    const levelQuestion = document.getElementById('levelQuestion');
    if (levelQuestion) {
        levelQuestion.style.display = 'none';
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

    if (/\s/.test(answer)) {
        feedback.textContent = 'Answer cannot contain spaces. Please enter a valid answer without spaces.';
        feedback.style.color = '#dc3545';
        setTimeout(() => {
            feedback.textContent = '';
            feedback.style.color = 'var(--primary)';
        }, 2000);
        return;
    }

    if (!/^[a-z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`]*$/.test(answer)) {
        feedback.textContent = 'Answer must be lowercase, alphanumeric with special characters only. No uppercase letters or spaces allowed.';
        feedback.style.color = '#dc3545';
        setTimeout(() => {
            feedback.textContent = '';
            feedback.style.color = 'var(--primary)';
        }, 2000);
        return;
    }

    if (!answer) {
        feedback.textContent = 'Answer cannot be empty. Please enter a valid answer.';
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
            
            setTimeout(() => {
                window.location.replace(window.location.pathname + '?v=' + Date.now());
            }, 800);
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
        console.error('Error submitting answer:', error);
        feedback.textContent = 'Error submitting answer. Please try again.';
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
                const response = await fetch('/api/announcements', {
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

async function updateSourceHint() {
    try {
        if (!currentLevel || !currentLevel.number) {
            const existingHints = document.querySelectorAll('meta[name="level-hint"]');
            existingHints.forEach(hint => hint.remove());
            
            for (let i = document.head.childNodes.length - 1; i >= 0; i--) {
                const node = document.head.childNodes[i];
                if (node.nodeType === 8 && node.data && node.data.trim().length > 0) {
                    document.head.removeChild(node);
                }
            }
            return;
        }

        const response = await fetch(`/api/user/level-hint/${currentLevel.number}`, {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });

        const head = document.head;
        
        const existingHints = document.querySelectorAll('meta[name="level-hint"]');
        existingHints.forEach(hint => hint.remove());
        
        for (let i = head.childNodes.length - 1; i >= 0; i--) {
            const node = head.childNodes[i];
            if (node.nodeType === 8 && node.data && node.data.trim().length > 0) {
                head.removeChild(node);
            }
        }

        if (response.ok) {
            const data = await response.json();
            if (data.hint && data.hint.trim() !== '') {
                const hintMeta = document.createElement('meta');
                hintMeta.setAttribute('name', 'level-hint');
                hintMeta.setAttribute('content', data.hint);
                hintMeta.setAttribute('data-level', currentLevel.number);
                head.appendChild(hintMeta);
                
                const hintComment = document.createComment(' ' + data.hint + ' ');
                head.appendChild(hintComment);
                
                console.log(`Updated source hint for level ${currentLevel.number}:`, data.hint);
            } else {
                console.log(`No source hint for level ${currentLevel.number}`);
            }
        }
    } catch (error) {
        console.error('Failed to update source hint:', error);
    }
}

document.addEventListener('DOMContentLoaded', function() {
    initializePage();
});