const urlParams = new URLSearchParams(window.location.search);
const errorType = urlParams.get('type');
const errorCode = urlParams.get('code');

switch (errorType) {
    case 'access_denied':
        document.getElementById('errorTitle').textContent = 'Access Denied';
        document.getElementById('errorMessage').textContent = 'You don\'t have permission to access this page. Please sign in with appropriate credentials.';
        break;
    case 'unauthorized':
        document.getElementById('errorTitle').textContent = 'Authorization Required';
        document.getElementById('errorMessage').textContent = 'You need to be signed in to access this page.';
        break;
    case 'admin_required':
        document.getElementById('errorTitle').textContent = 'Admin Access Required';
        document.getElementById('errorMessage').textContent = 'This page requires administrator privileges. Contact the organizing team if you believe this is an error.';
        break;
    case 'level_not_found':
        document.getElementById('errorTitle').textContent = 'Level Not Found';
        document.getElementById('errorMessage').textContent = 'The requested level doesn\'t exist or isn\'t available yet.';
        break;
    case 'session_expired':
        document.getElementById('errorTitle').textContent = 'Session Expired';
        document.getElementById('errorMessage').textContent = 'Your session has expired. Please sign in again to continue.';
        break;
    default:
        break;
}

async function checkAdminAccess() {
    try {
        const secret = await getSecret();
        const response = await fetch('/api/user/session', {
            headers: {
                'X-secret': secret
            }
        });
        if (response.ok) {
            const userData = await response.json();
            if (userData.isAdmin) {
                document.getElementById('adminLink').style.display = 'inline-block';
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

async function checkAuthAndUpdateButtons() {
    try {
        const secret = await getSecret();
        const response = await fetch('/api/user/session', {
            headers: {
                'X-secret': secret
            }
        });
        if (response.ok) {
            const userData = await response.json();
            if (userData.userId) {
                const signInBtn = document.querySelector('a[href="/auth"]');
                if (signInBtn) {
                    signInBtn.style.display = 'none';
                }
            }
            if (userData.isAdmin && document.getElementById('adminLink')) {
                document.getElementById('adminLink').style.display = 'inline-block';
            }
        }
    } catch (error) {
        console.log('User not authenticated');
    }
}

document.addEventListener('DOMContentLoaded', function() {
    checkAdminAccess();
    checkAuthAndUpdateButtons();
});
