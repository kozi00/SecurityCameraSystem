let currentPage = 1;
let pageSize = 24; // corresponds to 'limit' param on backend
let totalPages = 1;

function buildFilterQuery() {
    // If filter inputs exist (added on Pictures.html) include them
    const params = new URLSearchParams();
    const camEl = document.getElementById('filterCamera');
    const objEl = document.getElementById('filterObject');
    const afterEl = document.getElementById('filterAfter');
    const beforeEl = document.getElementById('filterBefore');
    if (camEl && camEl.value.trim()) params.set('camera', camEl.value.trim());
    if (objEl && objEl.value.trim()) params.set('object', objEl.value.trim());
    if (afterEl && afterEl.value) params.set('after', afterEl.value);
    if (beforeEl && beforeEl.value) params.set('before', beforeEl.value);
    return params.toString();
}


async function loadPictures(page = 1) {
    currentPage = page;
    const loadingEl = document.getElementById('loading');
    const errorEl = document.getElementById('error');
    const emptyEl = document.getElementById('empty');
    if (loadingEl) loadingEl.style.display = 'block';
    if (errorEl) errorEl.style.display = 'none';
    if (emptyEl) emptyEl.style.display = 'none';

    const filterQuery = getFiltersQuery();
    const baseUrl = `/api/pictures?page=${page}&limit=${pageSize}`;
    const url = filterQuery ? `${baseUrl}&${filterQuery}` : baseUrl;

    try {
        const response = await fetch(url);
        const data = await response.json();

        displayPictures(data);
        displayPagination(data);
        updateInfo(data);
        updateSizeBar(data);
    } catch (error) {
        if (errorEl) {
            errorEl.textContent = 'Błąd: ' + error.message;
            errorEl.style.display = 'block';
        }
    } finally {
        if (loadingEl) loadingEl.style.display = 'none';
    }
}

function displayPictures(data) {
    const gallery = document.getElementById('gallery');
    gallery.innerHTML = '';

    if (data.length === 0) {
        document.getElementById('empty').style.display = 'block';
        return;
    }

    data.pictures.forEach(picture => {
        const card = document.createElement('div');
        card.className = 'photo-card';
        
        card.innerHTML = `
            <img src="${data.imagesDir}/${picture}" 
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

    
    const maxSizeBytes = data.maxSize * 1024 * 1024 * 1024;
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
        sizeText.textContent = `${currentSizeGB.toFixed(2)} GB / ${data.maxSize} GB`;
    }
    
    if (sizeDetails) {
        const freeSpace = data.maxSize - currentSizeGB;
        sizeDetails.textContent = `Wolne: ${freeSpace.toFixed(2)} GB (${(100 - percentage).toFixed(1)}%)`;
    }
}

function changePageSize() {
    const el = document.getElementById('pageSize');
    if (el) {
        pageSize = parseInt(el.value, 10) || 24;
    }
    loadPictures(1);
}

function openPicture(filename) {
    window.open(`/api/pictures/view?image=${encodeURIComponent(filename)}`, '_blank');
}

function getFiltersQuery() {
        const c = document.getElementById('filterCamera').value.trim();
        const o = document.getElementById('filterObject').value.trim();
        const a = document.getElementById('filterAfter').value;
        const b = document.getElementById('filterBefore').value;
        const params = new URLSearchParams();
        if (c) params.set('camera', c);
        if (o) params.set('object', o);
        if (a) params.set('after', a);
        if (b) params.set('before', b);
        return params.toString();
    }

document.getElementById('applyFilters').addEventListener('click', () => loadPictures(1));
document.getElementById('resetFilters').addEventListener('click', () => {
    ['filterCamera','filterObject','filterAfter','filterBefore'].forEach(id => document.getElementById(id).value = '');
    loadPictures(1);
});

    

loadPictures(1);