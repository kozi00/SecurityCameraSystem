  let currentLogType = '';
        let autoRefresh = false;
        let refreshInterval;

        // ✅ LOAD LOGS
        async function loadLogs(type) {
            currentLogType = type;
            const logContent = document.getElementById('log-content');
            
            // Show loading
            logContent.textContent = '🔄 Ładowanie logów...';
            
            try {
                const response = await fetch(`/logs/${type}`);
                const text = await response.text();
                
                if (response.ok) {
                    logContent.textContent = text || `Brak logów typu: ${type}`;
                } else {
                    logContent.textContent = `❌ Błąd ładowania logów: ${response.status}`;
                }
            } catch (error) {
                logContent.textContent = `❌ Błąd połączenia: ${error.message}`;
            }
        }

        // ✅ CLEAR LOGS
        async function clearLogs(type) {
            const confirmed = confirm(`🗑️ Czy na pewno chcesz wyczyścić logi typu "${type}"?`);
            if (!confirmed) return;

            try {
                const response = await fetch(`/logs/${type}/clear`, {
                    method: 'POST',
                });
                
                if (response.ok) {
                    alert(`✅ Logi typu "${type}" zostały wyczyszczone!`);
                    
                    // Jeśli obecnie wyświetlamy te logi, odśwież
                    if (currentLogType === type) {
                        loadLogs(type);
                    }
                } else {
                    alert(`❌ Błąd czyszczenia logów: ${response.status}`);
                }
            } catch (error) {
                alert(`❌ Błąd połączenia: ${error.message}`);
            }
        }
