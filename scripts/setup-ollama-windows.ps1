# HomeLib — Установка Ollama на Windows GPU-машину
# Запуск: powershell -ExecutionPolicy Bypass -File setup-ollama-windows.ps1

Write-Host "=== Установка Ollama для HomeLib ===" -ForegroundColor Cyan

# 1. Установка Ollama
Write-Host "`n1. Установка Ollama..."
winget install Ollama.Ollama

# 2. Скачивание модели для эмбеддингов
Write-Host "`n2. Скачивание модели nomic-embed-text..."
ollama pull nomic-embed-text

# 3. Разрешить подключения из сети
Write-Host "`n3. Настройка сетевого доступа..."
[System.Environment]::SetEnvironmentVariable("OLLAMA_HOST", "0.0.0.0:11434", "User")

# 4. Увеличить параллелизм
Write-Host "`n4. Настройка параллелизма..."
[System.Environment]::SetEnvironmentVariable("OLLAMA_NUM_PARALLEL", "4", "User")

# 5. Настройка фаервола
$ServerIP = Read-Host "`nВведите IP-адрес сервера HomeLib (например, 192.168.1.100)"
if ($ServerIP) {
    Write-Host "5. Настройка фаервола..."
    New-NetFirewallRule -DisplayName "Ollama HomeLib" `
        -Direction Inbound -Protocol TCP -LocalPort 11434 `
        -RemoteAddress $ServerIP -Action Allow
    Write-Host "Правило фаервола создано для IP: $ServerIP" -ForegroundColor Green
}

Write-Host "`n=== Готово! ===" -ForegroundColor Green
Write-Host "Перезагрузите систему или перелогиньтесь для применения настроек."
Write-Host "После перезагрузки проверьте: curl http://localhost:11434/api/tags"
