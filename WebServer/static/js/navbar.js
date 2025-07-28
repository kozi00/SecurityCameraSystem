class Navbar {
    constructor() {
        this.currentPage = this.getCurrentPage();
        this.init();
    }
    getCurrentPage() {
        const path = window.location.pathname;
        if (path === '/' || path === '/index.html') return 'podglad';
        if (path.includes('/pictures')) return 'zdjecia';
        if (path.includes('/settings')) return 'ustawienia';
        return 'podglad';
    }

    init() {
        this.createNavbar();
        this.setActiveLink();
        this.addEventListeners();
    }

    createNavbar() {
        const navbar = document.createElement('nav');
        navbar.className = 'navbar';
        
        navbar.innerHTML = `
        
            <div class="navbar-container">
                <a href="/" class="navbar-logo">
                    <span class="navbar-icon">🔒</span>
                    <span>System Kamer</span>
                </a>
                
                <button class="navbar-toggle" onclick="toggleMobileMenu()">
                    ☰
                </button>
                
                <ul class="navbar-menu" id="navbar-menu">
                    <li class="navbar-item">
                        <a href="/" class="navbar-link" data-page="podglad">
                            <span class="navbar-icon">👁️</span>
                            <span>Podgląd</span>
                        </a>
                    </li>
                    <li class="navbar-item">
                        <a href="/pictures" class="navbar-link" data-page="zdjecia">
                            <span class="navbar-icon">📷</span>
                            <span>Zdjęcia</span>
                        </a>
                    </li>
                    <li class="navbar-item">
                        <a href="/settings" class="navbar-link" data-page="ustawienia">
                            <span class="navbar-icon">⚙️</span>
                            <span>Ustawienia</span>
                        </a>
                    </li>
                </ul>
            </div>
        `;
        
        // Wstaw navbar na początek body
        document.body.insertBefore(navbar, document.body.firstChild);
    }

    setActiveLink() {
        const links = document.querySelectorAll('.navbar-link');
        links.forEach(link => {
            link.classList.remove('active');
            if (link.dataset.page === this.currentPage) {
                link.classList.add('active');
            }
        });
    }

    addEventListeners() {
        // Obsługa kliknięć na linki
        document.addEventListener('click', (e) => {
            if (e.target.closest('.navbar-link')) {
                const link = e.target.closest('.navbar-link');
                this.handleNavigation(link);
            }
        });
    }

    handleNavigation(link) {
        // Usuń active ze wszystkich linków
        document.querySelectorAll('.navbar-link').forEach(l => l.classList.remove('active'));
        // Dodaj active do klikniętego
        link.classList.add('active');
        
        // Zamknij menu mobile
        const menu = document.getElementById('navbar-menu');
        menu.classList.remove('active');
    }
}

// Funkcja globalna dla przycisku mobile
function toggleMobileMenu() {
    const menu = document.getElementById('navbar-menu');
    menu.classList.toggle('active');
}

// Inicjalizacja nawigacji
document.addEventListener('DOMContentLoaded', () => {
    new Navbar();
});