let currentPage = 1;
let pageSize = 24; 
let totalPages = 1;

function buildFilterQuery() {
    const params = new URLSearchParams();
    const camEl = document.getElementById('filterCamera');
    const objEl = document.getElementById('filterObject');
    const timeAfterEl = document.getElementById('filterTimeAfter');
    const timeBeforeEl = document.getElementById('filterTimeBefore');
    const dateAfterEl = document.getElementById('filterDateAfter');
    const dateBeforeEl = document.getElementById('filterDateBefore');

    if (camEl && camEl.value.trim()) params.set('camera', camEl.value.trim());
    if (objEl && objEl.value.trim()) params.set('object', objEl.value.trim());
    if (timeAfterEl && timeAfterEl.value) params.set('timeAfter', timeAfterEl.value);
    if (timeBeforeEl && timeBeforeEl.value) params.set('timeBefore', timeBeforeEl.value);
    if (dateAfterEl && dateAfterEl.value) params.set('dateAfter', dateAfterEl.value);
    if (dateBeforeEl && dateBeforeEl.value) params.set('dateBefore', dateBeforeEl.value);
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
    const emptyMessage = document.getElementById('empty');
    
    gallery.innerHTML = '';

    const pictures = data.images || [];

    if (pictures.length === 0) {
        emptyMessage.style.display = 'block';
        return;
    } else {
        emptyMessage.style.display = 'none';
    }

    pictures.forEach(picture => {
        const card = document.createElement('div');
        card.className = 'photo-card';
        card.dataset.filename = picture.name;
        
        const imagePath = `${data.imagesDir}/${picture.name}`;

        card.innerHTML = `
            <div class="photo-image-container">
                <img src="${imagePath}" 
                        alt="${picture.name}"
                        onclick="openPicture('${picture.name}')"
                        onerror="this.parentElement.innerHTML='<div class=\'image-error\'>Błąd ładowania</div>'">
                <div class="photo-overlay">
                    <button class="btn-view" onclick="openPicture('${picture.name}')" title="Otwórz w nowej karcie">🔎</button>
                    <button class="btn-delete" onclick="confirmDeletePicture('${picture.name}')" title="Usuń zdjęcie">🗑️</button>
                </div>
            </div>
            <div class="photo-info">
                <div class="photo-info-row"><span class="label">Data:</span> ${picture.date}</div>
                <div class="photo-info-row"><span class="label">Godzina:</span> ${picture.timeOfDay}</div>
                <div class="photo-info-row"><span class="label">Kamera:</span> ${picture.camera}</div>
                <div class="photo-info-row">
                    <span class="label">Obiekt:</span> 
                    ${picture.objects && picture.objects.length > 0 ? picture.objects.join(", ") : "brak"}
                </div>
            </div>
        `;
        
        gallery.appendChild(card);
    });
}

function displayPagination(data) {
    const pagination = document.getElementById('pagination');
    pagination.innerHTML = '';
    
    if (data.totalPages <= 1) return;
    
    const prevBtn = document.createElement('button');
    prevBtn.textContent = '←';
    prevBtn.disabled = data.currentPage <= 1;
    prevBtn.onclick = () => loadPictures(data.currentPage - 1);
    pagination.appendChild(prevBtn);
    
    const start = Math.max(1, data.currentPage - 2);
    const end = Math.min(data.totalPages, data.currentPage + 2);
    
    for (let i = start; i <= end; i++) {
        const btn = document.createElement('button');
        btn.textContent = i;
        btn.className = i === data.currentPage ? 'active' : '';
    btn.onclick = () => loadPictures(i);
        pagination.appendChild(btn);
    }
    
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

let pictureToDelete = null;

function confirmDeletePicture(filename) {
    pictureToDelete = filename;
    const modal = document.getElementById('deleteModal');
    const filenameEl = document.getElementById('deleteFilename');
    if (filenameEl) filenameEl.textContent = filename;
    if (modal) modal.style.display = 'flex';
}

function closeDeleteModal() {
    const modal = document.getElementById('deleteModal');
    if (modal) modal.style.display = 'none';
    pictureToDelete = null;
}

async function deletePicture() {
    if (!pictureToDelete) return;
    
    const filename = pictureToDelete;
    const deleteBtn = document.getElementById('confirmDeleteBtn');
    if (deleteBtn) {
        deleteBtn.disabled = true;
        deleteBtn.textContent = 'Usuwanie...';
    }
    
    try {
        const response = await fetch(`/api/pictures/delete?filename=${encodeURIComponent(filename)}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error('Błąd podczas usuwania: ' + response.status);
        }
        
        // Animacja usunięcia karty
        const card = document.querySelector(`.photo-card[data-filename="${filename}"]`);
        if (card) {
            card.classList.add('deleting');
            setTimeout(() => card.remove(), 300);
        }
        
        showNotification('Zdjęcie zostało usunięte', 'success');
        
        // Odśwież listę po krótkim opóźnieniu
        setTimeout(() => loadPictures(currentPage), 500);
        
    } catch (error) {
        showNotification('Błąd: ' + error.message, 'error');
    } finally {
        closeDeleteModal();
        if (deleteBtn) {
            deleteBtn.disabled = false;
            deleteBtn.textContent = 'Usuń';
        }
    }
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <span>${message}</span>
        <button onclick="this.parentElement.remove()">&times;</button>
    `;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.classList.add('fade-out');
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

function getFiltersQuery() {
        const c = document.getElementById('filterCamera').value.trim();
        const o = document.getElementById('filterObject').value.trim();
        const ta = document.getElementById('filterTimeAfter').value;
        const tb = document.getElementById('filterTimeBefore').value;
        const da = document.getElementById('filterDateAfter').value;
        const db = document.getElementById('filterDateBefore').value;
        const params = new URLSearchParams();
        if (c) params.set('camera', c);
        if (o) params.set('object', o);
        if (ta) params.set('timeAfter', ta);
        if (tb) params.set('timeBefore', tb);
        if (da) params.set('dateAfter', da);
        if (db) params.set('dateBefore', db);
        return params.toString();
    }

document.getElementById('applyFilters').addEventListener('click', () => loadPictures(1));
document.getElementById('resetFilters').addEventListener('click', () => {
    ['filterCamera','filterObject','filterTimeAfter','filterTimeBefore','filterDateAfter','filterDateBefore'].forEach(id => document.getElementById(id).value = '');
    loadPictures(1);
});

document.getElementById('clearAllPictures').addEventListener('click', async ()=>{
            if (!confirm('Na pewno usunąć wszystkie zdjęcia? Tej operacji nie można cofnąć.')) return;
            try {
                const res = await fetch('/api/pictures/clear', { method: 'POST' });
                if (!res.ok) throw new Error('Błąd czyszczenia: ' + res.status);
                showNotification('Wszystkie zdjęcia zostały usunięte', 'success');
                loadPictures(1);
            } catch(e) {
                showNotification('Błąd: ' + e.message, 'error');
            }
        });

// Zamknij modal klikając poza nim
document.getElementById('deleteModal')?.addEventListener('click', (e) => {
    if (e.target.id === 'deleteModal') {
        closeDeleteModal();
    }
});

// Zamknij modal klawiszem Escape
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        closeDeleteModal();
    }
});

loadPictures(1);