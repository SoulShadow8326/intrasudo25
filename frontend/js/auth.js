let userEmail = '';

async function handleEmailSubmit(event) {
    event.preventDefault();
    
    const email = document.getElementById('email').value.trim();
    
    if (!email) {
        showError('emailError', 'Please enter your email address');
        return;
    }
    
    if (!validateEmail(email)) {
        showError('emailError', 'Please enter a valid email address');
        return;
    }
    
    setEmailLoading(true);
    hideError('emailError');
    
    try {
        const secret = await getSecret('POST');
        if (!secret) {
            throw new Error('Authentication service unavailable');
        }
        
        const params = new URLSearchParams();
        params.append('gmail', email);
        
        console.log('Submitting email:', email);
        
        const response = await fetch('/enter/email', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'X-secret': secret
            },
            body: params
        });
        
        console.log('Response status:', response.status);
        
        const data = await response.json();
        console.log('Response data:', data);
        
        if (response.ok) {
            userEmail = email;
            
            if (data.existing_user === "true") {
                showPopup('info', 'Account Found', 'You already have an account. Please enter your permanent 4-digit login code to continue.', () => {
                    showCodeForm();
                });
            } else {
                showPopup('success', 'Code Sent!', 'Check your email for your permanent 4-digit login code.', () => {
                    showCodeForm();
                });
            }
        } else {
            if (data.cooldown === "true") {
                showPopup('warning', 'Please Wait', data.error);
            } else {
                showError('emailError', data.error || 'Failed to send login code');
            }
        }
        
    } catch (error) {
        console.error('Email submission error:', error);
        showError('emailError', `Network error: ${error.message || 'Please try again.'}`);
    } finally {
        setEmailLoading(false);
    }
}

async function handleCodeSubmit(event) {
    event.preventDefault();
    
    const code = document.getElementById('verification-code').value.trim();
    
    if (!code || code.length !== 4) {
        showError('codeError', 'Please enter your 4-digit login code');
        return;
    }
    
    setCodeLoading(true);
    hideError('codeError');
    
    try {
        const secret = await getSecret('POST');
        if (!secret) {
            throw new Error('Authentication service unavailable');
        }
        
        const params = new URLSearchParams();
        params.append('gmail', userEmail);
        params.append('vnum', code);
        
        console.log('Submitting code:', code, 'for email:', userEmail);
        
        const response = await fetch('/enter/email-verify', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'X-secret': secret
            },
            body: params
        });
        
        console.log('Code verification response status:', response.status);
        
        const data = await response.json();
        console.log('Code verification response data:', data);
        
        if (response.ok) {
            window.location.href = '/home';
        } else {
            showError('codeError', data.error || 'Invalid verification code');
        }
        
    } catch (error) {
        console.error('Code verification error:', error);
        showError('codeError', `Network error: ${error.message || 'Please try again.'}`);
    } finally {
        setCodeLoading(false);
    }
}

function showCodeForm() {
    document.getElementById('email-form').style.display = 'none';
    document.getElementById('code-form').style.display = 'block';
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

function showError(elementId, message) {
    const errorElement = document.getElementById(elementId);
    errorElement.textContent = message;
    errorElement.style.display = 'block';
    errorElement.style.color = '#dc3545';
}

function hideError(elementId) {
    const errorElement = document.getElementById(elementId);
    errorElement.style.display = 'none';
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
    
    const handleClick = () => {
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
        const secret = await getSecret('GET');
        if (!secret) return;
        
        const response = await fetch('/api/user/session', {
            headers: {
                'X-secret': secret
            }
        });
        if (response.ok) {
            const data = await response.json();
            if (data.userId) {
                window.location.href = '/home';
                return;
            }
        }
    } catch (error) {
        console.log('No existing session found');
    }
}

document.addEventListener('DOMContentLoaded', () => {
    console.log('Auth page loaded, setting up form handlers');
    
    const emailForm = document.getElementById('email-form');
    if (emailForm) {
        emailForm.addEventListener('submit', handleEmailSubmit);
        console.log('Email form handler attached');
    }
    
    const codeForm = document.getElementById('code-form');
    if (codeForm) {
        codeForm.addEventListener('submit', handleCodeSubmit);
        console.log('Code form handler attached');
    }
    
    const backButton = document.getElementById('backButton');
    if (backButton) {
        backButton.addEventListener('click', showEmailForm);
        console.log('Back button handler attached');
    }
    
    const emailInput = document.getElementById('email');
    if (emailInput) {
        emailInput.focus();
    }
    
    checkExistingSession();
});
