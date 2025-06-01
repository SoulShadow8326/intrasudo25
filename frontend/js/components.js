function loadComponent(elementId, componentPath) {
    fetch(componentPath)
        .then(response => response.text())
        .then(html => {
            document.getElementById(elementId).innerHTML = html;
        })
        .catch(error => {
            console.error('Error loading component:', error);
        });
}

function loadFooter() {
    loadComponent('footer-placeholder', 'components/footer.html');
}

document.addEventListener('DOMContentLoaded', function() {
    loadFooter();
});
