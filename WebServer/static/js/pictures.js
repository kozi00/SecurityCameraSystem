let currentPage = 1;
let pageSize = 24;
let totalPages = 1;


async function loadPictures(page = 1) {
    document.getElementById('loading').style.display = 'block';
    document.getElementById('error').style.display = 'none';
    document.getElementById('empty').style.display = 'none';

    try {
        const response = await fetch(`/api/pictures?page=${page}&size=${pageSize}`);
        const data = await response.json();     
        
        displayPictures(data);
        displayPagination(data);
        updateInfo(data);
        updateSizeBar(data);
        
    } catch (error) {
        document.getElementById('error').textContent = 'Błąd: ' + error.message;
        document.getElementById('error').style.display = 'block';
    } finally {
        document.getElementById('loading').style.display = 'none';
    }
}

function displayPictures(data) {
    const gallery = document.getElementById('gallery');
    gallery.innerHTML = '';

    if (data.pictures.length === 0) {
        document.getElementById('empty').style.display = 'block';
        return;
    }

    data.pictures.forEach(picture => {
        const card = document.createElement('div');
        card.className = 'photo-card';
        
        card.innerHTML = `
            <img src="/static/images/${picture}" 
                    alt="${picture}"
                    onclick="openPicture('${picture}')"
                    onerror="this.style.display='none'">
            <div class="photo-info">
                <div class="photo-name">${picture}</div>
            </div>
        `;
        
        gallery.appendChild(card);
    });
}

function displayPagination(data) {
    const pagination = document.getElementById('pagination');
    pagination.innerHTML = '';
    
    if (data.totalPages <= 1) return;
    
    // Poprzednia
    const prevBtn = document.createElement('button');
    prevBtn.textContent = '←';
    prevBtn.disabled = data.currentPage <= 1;
    prevBtn.onclick = () => loadPictures(data.currentPage - 1);
    pagination.appendChild(prevBtn);
    
    // Numery stron
    const start = Math.max(1, data.currentPage - 2);
    const end = Math.min(data.totalPages, data.currentPage + 2);
    
    for (let i = start; i <= end; i++) {
        const btn = document.createElement('button');
        btn.textContent = i;
        btn.className = i === data.currentPage ? 'active' : '';
        btn.onclick = () => loadPictures(i);
        pagination.appendChild(btn);
    }
    
    // Następna
    const nextBtn = document.createElement('button');
    nextBtn.textContent = '→';
    nextBtn.disabled = data.currentPage >= data.totalPages;
    nextBtn.onclick = () => loadPictures(data.currentPage + 1);
    pagination.appendChild(nextBtn);
}

function updateInfo(data) {
    const info = document.getElementById('info');
    const start = (data.currentPage - 1) * data.pageSize + 1;
    const end = Math.min(data.currentPage * data.pageSize, data.length);
    info.textContent = `${start}-${end} z ${data.length}`;
}

function updateSizeBar(data) {
    const sizeBar = document.getElementById('sizeBar');
    if (!sizeBar) return;

    const maxSizeGB = 3; // 3GB maksymalnie
    const maxSizeBytes = maxSizeGB * 1024 * 1024 * 1024;
    const currentSizeBytes = data.size || 0;
    const currentSizeGB = currentSizeBytes / (1024 * 1024 * 1024);
    
    const percentage = Math.min((currentSizeBytes / maxSizeBytes) * 100, 100);
    
    const progressBar = sizeBar.querySelector('.size-progress');
    const sizeText = sizeBar.querySelector('.size-text');
    const sizeDetails = sizeBar.querySelector('.size-details');
    
    if (progressBar) {
        progressBar.style.width = percentage + '%';
        
        if (percentage < 50) {
            progressBar.style.backgroundColor = '#4CAF50'; 
        } else if (percentage < 80) {
            progressBar.style.backgroundColor = '#FF9800'; 
        } else {
            progressBar.style.backgroundColor = '#F44336'; 
        }
    }
    
    if (sizeText) {
        sizeText.textContent = `${currentSizeGB.toFixed(2)} GB / ${maxSizeGB} GB`;
    }
    
    if (sizeDetails) {
        const freeSpace = maxSizeGB - currentSizeGB;
        sizeDetails.textContent = `Wolne: ${freeSpace.toFixed(2)} GB (${(100 - percentage).toFixed(1)}%)`;
    }
}

function changePageSize() {
    pageSize = parseInt(document.getElementById('pageSize').value);
    loadPictures(1);
}

function openPicture(filename) {
    window.open(`/api/pictures/view?image=${encodeURIComponent(filename)}`, '_blank');
}

// Start
loadPictures(1);