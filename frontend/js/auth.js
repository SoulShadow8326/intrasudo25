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
            hideError('emailError');
            
            if (data.existing_user === "true") {
                showPopup('info', 'Account Found', 'You already have an account. Please enter your permanent 4-digit login code to continue.', () => {
                    showCodeForm();
                });
            } else {
                // For new users, show success message and code form
                showSuccess('emailSuccess', 'Code sent! Check your email for your permanent 4-digit login code.');
                setTimeout(() => {
                    showCodeForm();
                }, 1500); // Short delay to show the success message
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
            hideError('codeError');
            window.location.href = '/home';
        } else {
            showError('codeError', data.error || 'Invalid verification code');
        }
        
    } catch (error) {
        showError('codeError', `Network error: ${error.message || 'Please try again.'}`);
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
    
    // Hide all error and success messages
    hideError('emailError');
    hideError('codeError');
    const emailSuccess = document.getElementById('emailSuccess');
    const codeSuccess = document.getElementById('codeSuccess');
    if (emailSuccess) {
        emailSuccess.className = 'auth-success';
        emailSuccess.style.display = 'none';
    }
    if (codeSuccess) {
        codeSuccess.className = 'auth-success';
        codeSuccess.style.display = 'none';
    }
    
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
    errorElement.className = 'auth-error show';
    
    // Add smooth animation
    setTimeout(() => {
        errorElement.style.opacity = '1';
        errorElement.style.transform = 'translateY(0)';
    }, 10);
}

function hideError(elementId) {
    const errorElement = document.getElementById(elementId);
    errorElement.className = 'auth-error';
    errorElement.style.opacity = '0';
    errorElement.style.transform = 'translateY(-10px)';
    
    setTimeout(() => {
        errorElement.style.display = 'none';
    }, 300);
}

function showSuccess(elementId, message) {
    const successElement = document.getElementById(elementId);
    if (successElement) {
        successElement.textContent = message;
        successElement.className = 'auth-success show';
        
        setTimeout(() => {
            successElement.style.opacity = '1';
            successElement.style.transform = 'translateY(0)';
        }, 10);
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
        emailInput.addEventListener('input', () => {
            if (emailInput.value.trim() === '') {
                hideError('emailError');
            }
        });
    }
    
    const codeInput = document.getElementById('verification-code');
    if (codeInput) {
        codeInput.addEventListener('input', () => {
            if (codeInput.value.trim() === '') {
                hideError('codeError');
            }
        });
    }
});
