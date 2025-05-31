

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

async function checkUserSession() {
    try {
        const response = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        if (response.ok) {
            return await response.json();
        }
    } catch (error) {
        console.error('Failed to check user session:', error);
    }
    return null;
}

async function checkAdminAccess() {
    try {
        const response = await fetch('/api/user/session', {
            headers: {
                'CSRFtok': getCookie('X-CSRF_COOKIE') || ''
            }
        });
        if (response.ok) {
            const userData = await response.json();
            if (userData.isAdmin) {
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

function toggleMobileMenu() {
    const mobileMenu = document.getElementById('mobileNavMenu');
    const menuToggle = document.querySelector('.mobile-menu-toggle');
    
    if (mobileMenu && menuToggle) {
        const isActive = mobileMenu.classList.contains('active');
        
        if (isActive) {
            mobileMenu.classList.remove('active');
            menuToggle.classList.remove('active');
            document.body.classList.remove('mobile-menu-open');
        } else {
            mobileMenu.classList.add('active');
            menuToggle.classList.add('active');
            document.body.classList.add('mobile-menu-open');
        }
    }
}

document.addEventListener('click', function(event) {
    const mobileMenu = document.getElementById('mobileNavMenu');
    const menuToggle = document.querySelector('.mobile-menu-toggle');
    
    if (mobileMenu && menuToggle && 
        !mobileMenu.contains(event.target) && 
        !menuToggle.contains(event.target) &&
        mobileMenu.classList.contains('active')) {
        mobileMenu.classList.remove('active');
        menuToggle.classList.remove('active');
        document.body.classList.remove('mobile-menu-open');
    }
});

document.addEventListener('DOMContentLoaded', function() {
    const mobileNavLinks = document.querySelectorAll('.mobile-nav-links .nav-link');
    mobileNavLinks.forEach(link => {
        link.addEventListener('click', function() {
            const mobileMenu = document.getElementById('mobileNavMenu');
            const menuToggle = document.querySelector('.mobile-menu-toggle');
            
            if (mobileMenu && menuToggle) {
                mobileMenu.classList.remove('active');
                menuToggle.classList.remove('active');
            }
        });
    });
});

async function handleLogout() {
    try {
        const response = await fetch('/api/auth/logout', {
            method: 'POST'
        });
        
        if (response.ok) {
            window.location.href = '/auth';
        } else {
            console.error('Logout failed');
            window.location.href = '/auth';
        }
    } catch (error) {
        console.error('Error during logout:', error);
        window.location.href = '/auth';
    }
}
