let lastCountdownChecksum = null;

async function getCountdownChecksum() {
    try {
        const response = await fetch('/api/countdown-checksum?' + Date.now());
        const data = await response.json();
        return data.checksum;
    } catch (error) {
        console.error('Countdown checksum error:', error);
        return null;
    }
}

async function checkCountdownStatus() {
    try {
        const response = await fetch('/api/countdown-status?' + Date.now());
        const data = await response.json();
        
        if (data.status === 'not_started' || data.status === 'ended') {
            if (window.location.pathname !== '/status') {
                window.location.href = '/status';
            }
        } else if (data.status === 'active') {
            if (window.location.pathname === '/status') {
                window.location.href = '/landing';
            }
        }
    } catch (error) {
        console.error('Countdown status error:', error);
    }
}

async function pollCountdownStatus() {
    const checksum = await getCountdownChecksum();
    
    if (checksum && checksum !== lastCountdownChecksum) {
        lastCountdownChecksum = checksum;
        await checkCountdownStatus();
    }
}

checkCountdownStatus();
setInterval(pollCountdownStatus, 3000);
