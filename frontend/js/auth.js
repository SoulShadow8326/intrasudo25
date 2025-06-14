let userEmail = '';

async function handleEmailSubmit(event) {
    event.preventDefault();
    
    const email = document.getElementById('email').value.trim();
    
    if (!email) {
        showNotification('Please enter your email address', 'error');
        return;
    }
    
    if (!validateEmail(email)) {
        showNotification('Please enter a valid email address', 'error');
        return;
    }
    
    setEmailLoading(true);
    
    try {
        const params = new URLSearchParams();
        params.append('gmail', email);
        
        console.log('Submitting email:', email);
        
        const response = await fetch('/enter/email', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: params
        });
        
        console.log('Response status:', response.status);
        
        const data = await response.json();
        console.log('Response data:', data);
        
        if (response.ok) {
            userEmail = email;
            
            if (data.existing_user === "true") {
                showPopup('info', 'Account Found', 'You already have an account. Please enter your permanent 8-digit login code to continue.', () => {
                    showCodeForm();
                });
            } else {
                // For new users, show success message and code form
                showNotification('Code sent! Check your email for your permanent 8-digit login code.', 'success');
                setTimeout(() => {
                    showCodeForm();
                }, 1500); // Short delay to show the success message
            }
        } else {
            if (data.cooldown === "true") {
                showPopup('warning', 'Please Wait', data.error);
            } else {
                showNotification(data.error || 'Failed to send login code', 'error');
            }
        }
        
    } catch (error) {
        console.error('Email submission error:', error);
        showNotification(`Network error: ${error.message || 'Please try again.'}`, 'error');
    } finally {
        setEmailLoading(false);
    }
}

async function handleCodeSubmit(event) {
    event.preventDefault();
    
    const code = document.getElementById('verification-code').value.trim();
    
    if (!code || code.length !== 8) {
        showNotification('Please enter your 8-digit login code', 'error');
        return;
    }
    
    setCodeLoading(true);
    
    try {
        const params = new URLSearchParams();
        params.append('gmail', userEmail);
        params.append('vnum', code);
        
        console.log('Submitting code:', code, 'for email:', userEmail);
        
        const response = await fetch('/enter/email-verify', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: params
        });
        
        console.log('Code verification response status:', response.status);
        
        const data = await response.json();
        console.log('Code verification response data:', data);
        
        if (response.ok) {
            window.location.href = '/home';
        } else {
            showNotification(data.error || 'Invalid verification code', 'error');
        }
        
    } catch (error) {
        showNotification(`Network error: ${error.message || 'Please try again.'}`, 'error');
    } finally {
        setCodeLoading(false);
    }
}

function showCodeForm() {
    document.getElementById('email-form').style.display = 'none';
    document.getElementById('code-form').style.display = 'block';
    
    // Hide any success messages from email form
    const emailSuccess = document.getElementById('emailSuccess');
    if (emailSuccess) {
        emailSuccess.className = 'auth-success';
        emailSuccess.style.display = 'none';
    }
    
    document.getElementById('verification-code').focus();
}

function showEmailForm() {
    document.getElementById('code-form').style.display = 'none';
    document.getElementById('email-form').style.display = 'block';
    document.getElementById('email').value = '';
    document.getElementById('verification-code').value = '';
    
    userEmail = '';
    document.getElementById('email').focus();
}

function setEmailLoading(loading) {
    const emailButton = document.getElementById('emailButton');
    const emailButtonText = document.getElementById('emailButtonText');
    
    if (loading) {
        emailButton.disabled = true;
        emailButtonText.textContent = 'Sending Code...';
        emailButton.classList.add('loading');
    } else {
        emailButton.disabled = false;
        emailButtonText.textContent = 'Get Login Code';
        emailButton.classList.remove('loading');
    }
}

function setCodeLoading(loading) {
    const codeButton = document.getElementById('codeButton');
    const codeButtonText = document.getElementById('codeButtonText');
    
    if (loading) {
        codeButton.disabled = true;
        codeButtonText.textContent = 'Signing In...';
        codeButton.classList.add('loading');
    } else {
        codeButton.disabled = false;
        codeButtonText.textContent = 'Sign In';
        codeButton.classList.remove('loading');
    }
}

function validateEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

function showPopup(type, title, message, callback = null) {
    const modal = document.getElementById('authModal');
    const titleEl = document.getElementById('authModalTitle');
    const messageEl = document.getElementById('authModalMessage');
    const button = document.getElementById('authModalOkButton');
    
    titleEl.textContent = title;
    messageEl.textContent = message;
    modal.classList.add('show');
    
    const autoHideTimeout = setTimeout(() => {
        hidePopup();
        if (callback) callback();
    }, 5000);
    
    const handleClick = () => {
        clearTimeout(autoHideTimeout);
        hidePopup();
        if (callback) callback();
        button.removeEventListener('click', handleClick);
    };
    
    button.addEventListener('click', handleClick);
}

function hidePopup() {
    const modal = document.getElementById('authModal');
    modal.classList.remove('show');
}

async function checkExistingSession() {
    try {
        const response = await fetch('/api/user/session');
        if (response.ok) {
            const data = await response.json();
            if (data.userId) {
                window.location.href = '/home';
                return;
            }
        }
    } catch (error) {
        
    }
}

document.addEventListener('DOMContentLoaded', () => {
    
    const emailForm = document.getElementById('email-form');
    if (emailForm) {
        emailForm.addEventListener('submit', handleEmailSubmit);
    }
    
    const codeForm = document.getElementById('code-form');
    if (codeForm) {
        codeForm.addEventListener('submit', handleCodeSubmit);
    }
    
    const backButton = document.getElementById('backButton');
    if (backButton) {
        backButton.addEventListener('click', showEmailForm);
    }
    
    const emailInput = document.getElementById('email');
    if (emailInput) {
        emailInput.focus();
        emailInput.addEventListener('input', () => {});
    }
    
    const codeInput = document.getElementById('verification-code');
    if (codeInput) {
        codeInput.addEventListener('input', () => {});
    }
});
